package handlers

import (
	"fmt"
	"net/http"

	"github.com/robiball/slack"
	"github.com/skybet/go-helpdesk/wrapper"
)

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

// DialogTest is a simple test handler to create a dialog in Slack
func DialogTest(w http.ResponseWriter, r *http.Request) error {
	//get_the_formatted_request and from it get the trigger id
	sc, err := slack.SlashCommandParse(r)
	if err != nil {
		return fmt.Errorf("Failed to parse slack slash command: %s", err)
	}

	descElement := slack.DialogTextElement{
		Type:        "text",
		Label:       "Description",
		Placeholder: "Description...",
		Name:        "FOO",
	}

	elements := []slack.DialogElement{
		descElement,
	}

	dialog := slack.Dialog{
		CallbackId:     "PETETEST",
		Title:          "Create an Incident",
		SubmitLabel:    "Create",
		NotifyOnCancel: true,
		Elements:       elements,
	}

	if err := wrapper.OpenDialog(sc.TriggerID, dialog); err != nil {
		return fmt.Errorf("Failed to open dialog: %s", err)
	}
	return nil
}
