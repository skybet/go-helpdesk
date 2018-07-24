package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
	Logf         LogfFunc
	Routes       []*Route
	DefaultRoute SlackHandlerFunc
	basePath     string
	appToken     string
	secretToken  string
}

// NewSlackHandler returns an initialised SlackHandler
func NewSlackHandler(basePath, appToken, secretToken string, l LogFunc, lf LogfFunc) *SlackHandler {
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

	serve := func(f SlackHandlerFunc, ctx interface{}) {
		if err := f(res, req, ctx); err != nil {
			h.Logf("HTTP handler error: %s", err)
		}
	}

	if !h.validRequest(r) {
		http.Error(w, "invalid slack request", 400)
		return
	}

	// First check if path matches our BasePath
	// If yes then attempt to decode it to match on Command or CallbackID / InteractionType
	// If no then match custom paths
	if r.URL.Path == h.basePath {
		if err := r.ParseForm(); err != nil {
			_ = fmt.Errorf("Unable to parse request from Slack: %s", err)
			serve(h.DefaultRoute, nil)
			return
		}

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
			// It's a command but we have no handler for it - 404
			serve(h.DefaultRoute, nil)
			return
		} else {
			var payloadMap map[string]interface{}
			payloadJSON := r.Form.Get("payload")
			if err := json.Unmarshal([]byte(payloadJSON), &payloadMap); err != nil {
				h.Logf("%s", err)
				serve(h.DefaultRoute, nil)
				return
			}
			if payloadMap["type"] == nil {
				h.Log("Error parsing Slack Event: Missing value for 'type' key")
				serve(h.DefaultRoute, nil)
				return
			}
			if payloadMap["callback_id"] == nil {
				h.Log("Error parsing Slack Event: Missing value for 'callback_id' key")
				serve(h.DefaultRoute, nil)
				return
			}
			// Loop through all our routes and attempt a match on the InteractionType / CallbackID pair
			for _, rt := range h.Routes {
				if payloadMap["type"] == rt.InteractionType && payloadMap["callback_id"] == rt.CallbackID {
					// Send the payloadJSON as context
					serve(rt.Handler, payloadJSON)
					return
				}
			}
		}
		// Its path is the basepath, but we dont have a matching command or
		// action handler for it - 404
		serve(h.DefaultRoute, nil)
		return
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

func (h *SlackHandler) validRequest(r *http.Request) bool {
	slackTimestampHeader := r.Header.Get("X-Slack-Request-Timestamp")
	slackTimestamp, err := strconv.ParseInt(slackTimestampHeader, 10, 64)

	// Abort if timestamp is invalid
	if err != nil {
		h.Logf("Invalid timestamp sent from slack: %s", err)
		return false
	}

	// Abort if timestamp is stale (older than 5 minutes)
	now := int64(time.Now().Unix())
	if (now - slackTimestamp) > (60 * 5) {
		h.Logf("Stale timestamp sent from slack: %s", err)
		return false
	}

	// Abort if request body is invalid
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.Logf("Invalid request body sent from slack: %s", err)
		return false
	}
	slackBody := string(body)

	// Abort if the signature does not correspond to the signing secret
	slackBaseStr := []byte(fmt.Sprintf("v0:%d:%s", slackTimestamp, slackBody))
	slackSignature := r.Header.Get("X-Slack-Signature")
	sec := hmac.New(sha256.New, []byte(h.secretToken))
	sec.Write(slackBaseStr)
	mySig := fmt.Sprintf("v0=%s", []byte(hex.EncodeToString(sec.Sum(nil))))
	if mySig != slackSignature {
		h.Log("Invalid signature sent from slack, ignoring request.")
		return false
	}

	// All good! The request is valid
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return true
}
