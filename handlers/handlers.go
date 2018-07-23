package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"github.com/skybet/go-helpdesk/server"
	"github.com/skybet/go-helpdesk/wrapper"
)

var slackWrapper wrapper.SlackWrapper

// Init initialises any external dependencies
func Init(sw wrapper.SlackWrapper) {
	slackWrapper = sw
}

func HelpCallback(res *server.Response, req *server.Request, ctx interface{}) error {
	s, ok := ctx.(string)
	if !ok {
		return fmt.Errorf("Expected a string to be passed to the handler")
	}
	var d *slack.DialogCallback
	err := json.Unmarshal([]byte(s), &d)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	log.Printf("User: '%s' Requested Help: '%s'", d.User.Name, d.Submission["HelpRequestDescription"])
	return nil
}

// HelpRequest is a handler that creates a dialog in Slack to capture a
// customers help request
func HelpRequest(res *server.Response, req *server.Request, ctx interface{}) error {
	sc, ok := ctx.(slack.SlashCommand)
	if !ok {
		return fmt.Errorf("Expected a slack.SlashCommand to be passed to the handler")
	}
	descriptionElement := slack.DialogTextElement{
		Type:        "text",
		Label:       "Help Request Description",
		Placeholder: "Describe what you would like help with ...",
		Name:        "HelpRequestDescription",
	}

	elements := []slack.DialogElement{
		descriptionElement,
	}

	dialog := slack.Dialog{
		CallbackId:     "HelpRequest",
		Title:          "Request Help",
		SubmitLabel:    "Create",
		NotifyOnCancel: true,
		Elements:       elements,
	}

	if err := slackWrapper.OpenDialog(sc.TriggerID, dialog); err != nil {
		return fmt.Errorf("Failed to open dialog: %s", err)
	}
	return nil
}
