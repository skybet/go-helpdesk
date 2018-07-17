package server

import (
	"fmt"
	"net/http"
)

// LogFunc is an abstraction that allows using any external logger with a Printf signature
// Set to nil to disable logging completely
type LogFunc func(string, ...interface{})

// SlashCommandHandlerFunc is an http.HandlerFunc which can return an error
type SlashCommandHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// SlashCommandHandler is a function executed when a slash command is invoked
type SlashCommandHandler struct {
	Log LogFunc
	H   SlashCommandHandlerFunc
}

// ServeHTTP satisfies http.Handler interface
func (h SlashCommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(w, r)
	if err != nil {
		h.Log("HTTP Error: %s", err)
	}
}

// SlashCommand is a handler which is invoked when a path is matched
type SlashCommand struct {
	Path    string
	Handler SlashCommandHandlerFunc
}

// SlackReceiver is a server which responds to events sent Slack in response to slash commands etc.
type SlackReceiver struct {
	slashCommands map[string]*SlashCommand
}

// NewSlackReceiver returns a new SlackReceiver
func NewSlackReceiver() *SlackReceiver {
	return &SlackReceiver{
		slashCommands: make(map[string]*SlashCommand),
	}
}

// Start the receiver, blocking
func (s *SlackReceiver) Start(addr string, log LogFunc) error {
	for _, r := range s.slashCommands {
		h := SlashCommandHandler{Log: log, H: r.Handler}
		http.Handle(r.Path, h)
	}
	return http.ListenAndServe(addr, nil)
}

// AddSlashCommand adds a new slash command
func (s *SlackReceiver) AddSlashCommand(sc *SlashCommand) error {
	// TODO: Validate path
	// Already exists?
	if _, ok := s.slashCommands[sc.Path]; ok {
		return fmt.Errorf("A slash command at this path is already configured")
	}
	s.slashCommands[sc.Path] = sc
	return nil
}

// RemoveSlashCommand removes an existing slash command
func (s *SlackReceiver) RemoveSlashCommand(sc *SlashCommand) error {
	// Exists?
	if _, ok := s.slashCommands[sc.Path]; !ok {
		return fmt.Errorf("No slash command configured with this path")
	}
	delete(s.slashCommands, sc.Path)
	return nil
}
