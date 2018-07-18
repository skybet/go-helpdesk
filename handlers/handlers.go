package handlers

import (
	"fmt"
	"net/http"

	"github.com/robiball/slack"
	"github.com/skybet/go-helpdesk/wrapper"
)

// HelpRequest is a handler that creates a dialog in Slack to capture a
// customers help request
func HelpRequest(w http.ResponseWriter, r *http.Request) error {
	sc, err := slack.SlashCommandParse(r)
	if err != nil {
		return fmt.Errorf("Failed to parse slack slash command: %s", err)
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
