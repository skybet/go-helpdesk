package server

import (
	"net/http"
	"strings"
	"github.com/nlopes/slack"
)

// LogFunc is an abstraction that allows using any external logger with a Print signature
// Set to nil to disable logging completely
type LogFunc func(...interface{})

// LogfFunc is an abstraction that allows using any external logger with a Printf signature
// Set to nil to disable logging completely
type LogfFunc func(string, ...interface{})

// SlackHandlerFunc is an http.HandlerFunc which can return an error and has a context
// Context varies depending on the request type and is for injecting arbitrary data in
// at routing time
type SlackHandlerFunc func(res *Response, req *Request, ctx interface{}) error

// Route is a handler which is invoked when a path is matched
type Route struct {
	CallbackID, Path, Command, InteractionType string
	Handler                                    SlackHandlerFunc
}

// SlackHandler is a function executed when a route is invoked
type SlackHandler struct {
	Log          LogFunc
	Logf         LogfFunc
	Routes       []*Route
	DefaultRoute SlackHandlerFunc
	basePath     string
	appToken     string
	secretToken  string
	dnHeader     *string	// Used for Mutual TLS
}

// NewSlackHandler returns an initialised SlackHandler
func NewSlackHandler(basePath, appToken, secretToken string, dnHeader *string, l LogFunc, lf LogfFunc) *SlackHandler {
	return &SlackHandler{
		DefaultRoute: func(res *Response, req *Request, ctx interface{}) error {
			res.Text(http.StatusNotFound, "Not found")
			return nil
		},
		Log:         l,
		Logf:        lf,
		basePath:    basePath,
		appToken:    appToken,
		secretToken: secretToken,
		dnHeader: dnHeader,
	}
}

// HandleCallback registers a handler to be executed when a specific
// InteractionType / CallbackID pair is present in the request
// payload sent to the BasePath
func (h *SlackHandler) HandleCallback(it, cid string, f SlackHandlerFunc) {
	r := &Route{Path: h.basePath, CallbackID: cid, InteractionType: it, Handler: f}
	h.handle(r)
}

// HandleCommand registers a handler to be executed when a slash command
// request is sent to the BasePath
func (h *SlackHandler) HandleCommand(c string, f SlackHandlerFunc) {
	r := &Route{Path: h.basePath, Command: c, Handler: f}
	h.handle(r)
}

// HandlePath registers handlers for specific paths
func (h *SlackHandler) HandlePath(p string, f SlackHandlerFunc) {
	r := &Route{Path: p, Handler: f}
	h.handle(r)
}

func (h *SlackHandler) handle(r *Route) {
	// TODO: validate no duplicates
	h.Routes = append(h.Routes, r)
}

// ServeHTTP satisfies http.Handler interface
func (h *SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{Request: r}
	res := &Response{w}

	// Generic serve function which captures and logs handler errors
	serve := func(f SlackHandlerFunc, ctx interface{}) {
		if err := f(res, req, ctx); err != nil {
			h.Logf("HTTP handler error: %s", err)
		}
	}

	// If the request did not look like it came from slack, 400 and abort
	if err := req.Validate(h.secretToken, h.dnHeader); err != nil {
		h.Logf("Bad request from slack: %s", err)
		res.Text(400, "invalid slack request")
		return
	}

	// First check if path matches our BasePath and has valid form data
	// If yes then attempt to decode it to match on Command or CallbackID / InteractionType
	// If no then match custom paths
	err := r.ParseForm()
  	if strings.HasPrefix(r.URL.Path, h.basePath) && err == nil {
		// Is it a slash command?
		if r.Form.Get("command") != "" {
			sc, _ := slack.SlashCommandParse(r)
			// Loop through all our routes and attempt a match on the Command
			for _, rt := range h.Routes {
				if rt.Command == sc.Command {
					// Send the SlackCommand struct as context
					serve(rt.Handler, sc)
					return
				}
			}
		}

		// Does it have a valid callback payload? - If so, it's a callback
		payload, err := req.CallbackPayload()
		if err != nil {
			h.Logf("Error parsing payload: %s", err)
		}
		// Loop through all our routes and attempt a match on the InteractionType / CallbackID pair
		for _, rt := range h.Routes {
			if payload.MatchRoute(rt) {
				// Send the payload as context
				t, err := payload.Mutate()
				if err != nil {
					h.Logf("Error mutating %s payload: %s", payload["type"], err)
				}
				serve(rt.Handler, t)
				return
			}
		}
	} else {
		// If nothing else works, loop through all our routes and attempt a match on the path
		for _, rt := range h.Routes {
			if rt.Path == r.URL.Path {
				serve(rt.Handler, nil)
				return
			}
		}
	}

	// No matches - 404
	serve(h.DefaultRoute, nil)
}
