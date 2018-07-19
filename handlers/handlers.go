package handlers

import (
	"fmt"

	"github.com/nlopes/slack"
	"github.com/skybet/go-helpdesk/server"
	"github.com/skybet/go-helpdesk/wrapper"
)

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

	if err := wrapper.OpenDialog(sc.TriggerID, dialog); err != nil {
		return fmt.Errorf("Failed to open dialog: %s", err)
	}
	return nil
}
