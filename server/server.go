package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/robiball/slack"
	"github.com/robiball/slack/slackevents"
)

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

// LogFunc is an abstraction that allows using any external logger with a Printf signature
// Set to nil to disable logging completely
type LogFunc func(string, ...interface{})

// SlackHandlerFunc is an http.HandlerFunc which can return an error and has a context
// Context varies depending on the request type and is for injecting arbitrary data in
// at routing time
type SlackHandlerFunc func(res *Response, req *Request, ctx interface{}) error

// SlackHandler is a function executed when a route is invoked
type SlackHandler struct {
	Log          LogFunc
	Routes       []*Route
	DefaultRoute SlackHandlerFunc
	BasePath     string
	Context      interface{}
}

// NewSlackHandler returns an initialised SlackHandler
func NewSlackHandler(basePath string, l LogFunc) *SlackHandlerFunc {
	return &SlackHandlerFunc{
		DefaultRoute: func(res *Response, req *Request, ctx interface{}) error {
			res.Text(http.StatusNotFound, "Not found")
		},
		Log:      l,
		BasePath: string,
	}
}

// HandleCallback registers a handler to be executed when a specific
// CallbackID is present in the request payload sent to the BasePath
func (h SlackHandler) HandleCallback(cid string, f SlackHandlerFunc) {
	r := &Route{Path: h.BasePath, CallbackID: cid, Handler: f}
	h.handle(r)
}

// HandleCommand registers a handler to be executed when a slash command
// request is sent to the BasePath
func (h SlackHandler) HandleCommand(c string, f SlackHandlerFunc) {
	r := &Route{Path: h.BasePath, Command: c, Handler: f}
	h.handle(r)
}

// HandlePath registers handlers for specific paths
func (h SlackHandler) HandlePath(p string, f SlackHandlerFunc) {
	r := &Route{Path: p, Handler: f}
	h.handle(r)
}

func (h SlackHandler) handle(r *Route) {
	// TODO: validate no duplicates
	h.Routes = append(h.Routes, r)
}

// ServeHTTP satisfies http.Handler interface
func (h SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{Request: r}
	res := &Response{w}

	// First check if path matches our BasePath
	// If yes then attempt to decode it to match on Command or CallbackID
	// If no then match custom paths
	if r.URL.Path == h.BasePath {
		// Attempt to parse as a slash command first
		// Ignore errors here as the library always returns a non-nil struct
		sc, _ := slack.SlashCommandParse(r)
		if len(sc.Command) > 0 {
			// Loop through all our routes and attempt a match on the Command
			for _, rt := range h.Routes {
				if rt.Command == sc.Command {
					// Send the SlackCommand struct as context
					err := rt.Handler(res, req, sc)
					if err != nil {
						h.Log("HTTP serve error: %s", err)
					}
					return
				}
			}
			if err := h.DefaultRoute(res, req); err != nil {
				h.Log("HTTP serve error: %s", err)
			}
			return
		}
		// Attempt to parse as an event
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()
		event, err := slackevents.ParseEvent(
			json.RawMessage(body),
			// TODO: This needs to be our real App Token
			slackevents.OptionVerifyToken(&slackevents.TokenComparator{"TOKEN"}),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if len(event.CallbackId) > 0 {
			for _, rt := range h.Routes {
				if rt.CallbackID == event.CallbackId {
					// Send the event struct as context
					err := rt.Handler(res, req, event)
					if err != nil {
						h.Log("HTTP serve error: %s", err)
					}
					return
				}
			}
		}
	} else {
		// Loop through all our routes and attempt a match on the path
		for _, rt := range h.Routes {
			if rt.Path == r.URL.Path {
				err := rt.Handler(res, req, nil)
				if err != nil {
					h.Log("HTTP serve error: %s", err)
				}
				return
			}
		}
	}
	// DefaultRoute is executed if nothing matches
	// Handle the error in case a custom DefaultRoute is supplied
	if err := h.DefaultRoute(res, req); err != nil {
		h.Log("HTTP serve error: %s", err)
	}
}
