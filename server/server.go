package server

import (
	"encoding/json"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"io/ioutil"
	"net/http"
	"strings"
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
	CallbackID, Path, Command, InteractionType, EventType string
	Handler                                               SlackHandlerFunc
}

// SlackHandler is a function executed when a route is invoked
type SlackHandler struct {
	Log          LogFunc
	Logf         LogfFunc
	ErrorLog     LogFunc
	ErrorLogf    LogfFunc
	Routes       []*Route
	DefaultRoute SlackHandlerFunc
	basePath     string
	appToken     string
	secretToken  string
	dnHeader     *string	// Used for Mutual TLS
}

// NewSlackHandler returns an initialised SlackHandler
func NewSlackHandler(basePath, appToken, secretToken string, dnHeader *string, l LogFunc, lf LogfFunc, el LogFunc, elf LogfFunc) *SlackHandler {
	return &SlackHandler{
		DefaultRoute: func(res *Response, req *Request, ctx interface{}) error {
			res.Text(http.StatusNotFound, "Not found")
			return nil
		},
		Log:         l,
		Logf:        lf,
		ErrorLog:    el,
		ErrorLogf:   elf,
		basePath:    basePath,
		appToken:    appToken,
		secretToken: secretToken,
		dnHeader:    dnHeader,
	}
}

// HandleInteractionCallback registers a handler to be executed when a specific
// InteractionType / CallbackID pair is present in the request
func (h *SlackHandler) HandleInteractionCallback(it, cid string, f SlackHandlerFunc) {
	r := &Route{Path: h.basePath, CallbackID: cid, InteractionType: it, Handler: f}
	h.handle(r)
}

// HandleEventCallback registers a handler to be executed when a specific
// EventsAPICallbackEvent type is present in the request
func (h *SlackHandler) HandleEventCallback(et string, f SlackHandlerFunc) {
	r := &Route{Path: h.basePath, EventType: et, Handler: f}
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
			h.ErrorLogf("HTTP handler error: %s", err)
			if b, bodyErr := ioutil.ReadAll(r.Body); bodyErr == nil {
				if len(b) > 0 {
					h.ErrorLogf("Request body: %s", string(b), nil)
				}
			}
		}
	}

	// If the request did not look like it came from slack, 400 and abort
	if err := req.Validate(h.secretToken, h.dnHeader); err != nil {
		h.ErrorLogf("Bad request from slack: %s", err)
		res.Text(400, "invalid slack request")
		return
	}

	// First check if path matches our BasePath and has valid form data
	// If yes then attempt to decode it to match on Command, Events challenge, or CallbackID / InteractionType
	// If no then match custom paths
	err := r.ParseForm()
	if strings.HasPrefix(r.URL.Path, h.basePath) && err == nil {
		// Is this a url challenge from Slack?
		body, err := ioutil.ReadAll(r.Body)
		// This has a body, lets do stuff with it
		if err == nil {
			// Decode the potential challenge interactionPayload
			var verificationEvent slackevents.EventsAPIURLVerificationEvent
			err := json.Unmarshal(body, &verificationEvent)
			if err == nil {
				// This seems to be a url verification request from Slack, check it is and respond accordingly
				if verificationEvent.Type == slackevents.URLVerification {
					if _, err := w.Write([]byte(verificationEvent.Challenge)); err != nil {
						h.ErrorLogf("Failed writing challenge back to verificationEvent: %w", err)
					}
					h.Logf("Successfully responded to URL verification requested from Slack")
					return
				}
			}
		}

		// Is it a slash command?
		if r.Form.Get("command") != "" {
			sc, _ := slack.SlashCommandParse(r)
			h.Logf("slack command triggered: %s", sc.Command)
			// Loop through all our routes and attempt a match on the Command
			for _, rt := range h.Routes {
				if rt.Command == sc.Command {
					// Send the SlackCommand struct as context
					serve(rt.Handler, sc)
					return
				}
			}
		}

		// Is it an interaction callback?
		if r.Form.Get("payload") != "" {
			// Does it have a valid interaction callback payload? - If so, it's an interaction callback
			interactionPayload, err := req.InteractionCallbackPayload()
			if err != nil {
				h.ErrorLogf("Error parsing interactionPayload: %s", err)
				w.WriteHeader(400)
				return
			}
			// Loop through all our routes and attempt a match on the InteractionType / CallbackID pair
			if interactionPayload != nil {
				h.Logf("slack interaction callback triggered: %s", interactionPayload.CallbackID)
				for _, rt := range h.Routes {
					if string(interactionPayload.Type) == rt.InteractionType && interactionPayload.CallbackID == rt.CallbackID {
						// Send the interactionPayload as context
						serve(rt.Handler, interactionPayload)
						return
					}
				}
			}
		}

		// Is it an event callback? If so see if we can route to it
		event, err := req.EventAPIEvent(body)
		if err == nil && event != nil {
			eventType := event.InnerEvent.Type
			h.Logf("slack event triggered: %s", eventType)
			// Loop through all our routes and attempt a match on the Event type
			for _, rt := range h.Routes {
				if eventType == rt.EventType {
					// Send the interactionPayload as context
					h.Logf("Serving request....")
					serve(rt.Handler, event)
					return
				}
			}
			// We want to exit here because it's a valid event, but we don't have a route for it
			h.Logf("no valid route found that matches [%s], returning", eventType)
			return
		}

		h.ErrorLogf("Event err:", err)
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
