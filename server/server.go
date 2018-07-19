package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

// LogFunc is an abstraction that allows using any external logger with a Printf signature
// Set to nil to disable logging completely
type LogFunc func(string, ...interface{})

// SlackHandlerFunc is an http.HandlerFunc which can return an error and has a context
// Context varies depending on the request type and is for injecting arbitrary data in
// at routing time
type SlackHandlerFunc func(res *Response, req *Request, ctx interface{}) error

// Route is a handler which is invoked when a path is matched
type Route struct {
	CallbackID, Path, Command string
	Handler                   SlackHandlerFunc
}

// Request wraps http.Request
type Request struct {
	*http.Request
}

// Response wraps http.ResponseWriter
type Response struct {
	http.ResponseWriter
}

// Text is a convenience method for sending a response
func (r *Response) Text(code int, body string) {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(code)

	io.WriteString(r, fmt.Sprintf("%s\n", body))
}

// SlackHandler is a function executed when a route is invoked
type SlackHandler struct {
	Log          LogFunc
	Routes       []*Route
	DefaultRoute SlackHandlerFunc
	basePath     string
	appToken     string
}

// NewSlackHandler returns an initialised SlackHandler
func NewSlackHandler(basePath, appToken string, l LogFunc) *SlackHandler {
	return &SlackHandler{
		DefaultRoute: func(res *Response, req *Request, ctx interface{}) error {
			res.Text(http.StatusNotFound, "Not found")
			return nil
		},
		Log:      l,
		basePath: basePath,
		appToken: appToken,
	}
}

// HandleCallback registers a handler to be executed when a specific
// CallbackID is present in the request payload sent to the BasePath
func (h *SlackHandler) HandleCallback(cid string, f SlackHandlerFunc) {
	r := &Route{Path: h.basePath, CallbackID: cid, Handler: f}
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

	serve := func(f SlackHandlerFunc, ctx interface{}) {
		if err := f(res, req, ctx); err != nil {
			h.Log("HTTP handler error: %s", err)
		}
	}

	// First check if path matches our BasePath
	// If yes then attempt to decode it to match on Command or CallbackID
	// If no then match custom paths
	if r.URL.Path == h.basePath {
		// Attempt to parse as a slash command first
		// Ignore errors here as the library always returns a non-nil struct
		sc, _ := slack.SlashCommandParse(r)
		if len(sc.Command) > 0 {
			// Loop through all our routes and attempt a match on the Command
			for _, rt := range h.Routes {
				if rt.Command == sc.Command {
					// Send the SlackCommand struct as context
					serve(rt.Handler, sc)
					return
				}
			}
			// It's a command but we have no handler for it - 404
			serve(h.DefaultRoute, nil)
			return
		}
		// Attempt to parse as an event
		event, err := h.actionEventHelper(r)
		if err != nil {
			h.Log("Error parsing Slack Event: %s", err)
		}
		if len(event.CallbackId) > 0 {
			for _, rt := range h.Routes {
				if rt.CallbackID == event.CallbackId {
					// Send the event struct as context
					serve(rt.Handler, event)
					return
				}
			}
		}
	} else {
		// Loop through all our routes and attempt a match on the path
		for _, rt := range h.Routes {
			if rt.Path == r.URL.Path {
				serve(rt.Handler, nil)
				return
			}
		}
	}
	// 404
	serve(h.DefaultRoute, nil)
}

func (h *SlackHandler) actionEventHelper(r *http.Request) (m slackevents.MessageAction, err error) {
	if e := r.ParseForm(); e != nil {
		return m, fmt.Errorf("Error parsing form data: %s", e)
	}
	m, err = slackevents.ParseActionEvent(
		r.Form.Get("payload"),
		slackevents.OptionVerifyToken(&slackevents.TokenComparator{h.appToken}),
	)
	return
}
