package main

import (
	"net/http"

	"github.com/robiball/slack"
	log "github.com/sirupsen/logrus"
	"github.com/skybet/go-helpdesk/wrapper"
)

func dialogTest(w http.ResponseWriter, r *http.Request) {
	//get_the_formatted_request and from it get the trigger id
	sc, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Errorf("Failed to parse slack slash command: %s", err)
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
		log.Errorf("Failed to open dialog: %s", err)
	}
}
