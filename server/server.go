package server

import (
	"fmt"
	"net/http"
)

// RouteHandler is a function executed when a route is invoked
type RouteHandler func(w http.ResponseWriter, r *http.Request)

// Route is a handler which is invoked when a path is matched
type Route struct {
	Path    string
	Handler RouteHandler
}

// SlackReceiver is a server which responds to events sent Slack in response to slash commands etc.
type SlackReceiver struct {
	routes map[string]*Route
}

// NewSlackReceiver returns a new SlackReceiver
func NewSlackReceiver() *SlackReceiver {
	return &SlackReceiver{
		routes: make(map[string]*Route),
	}
}

// Start the receiver, blocking
func (s *SlackReceiver) Start(addr string) error {
	for _, r := range s.routes {
		http.HandleFunc(r.Path, r.Handler)
	}
	return http.ListenAndServe(addr, nil)
}

// AddRoute adds a new route
func (s *SlackReceiver) AddRoute(route *Route) error {
	// TODO: Validate path
	// Already exists?
	if _, ok := s.routes[route.Path]; ok {
		return fmt.Errorf("A route at this path is already configured")
	}
	s.routes[route.Path] = route
	return nil
}

// RemoveRoute removes an existing route
func (s *SlackReceiver) RemoveRoute(route *Route) error {
	// Exists?
	if _, ok := s.routes[route.Path]; !ok {
		return fmt.Errorf("No route configured with this path")
	}
	delete(s.routes, route.Path)
	return nil
}
