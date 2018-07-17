package server

import (
	"fmt"
	"github.com/robiball/slack"
	"github.com/skybet/go-helpdesk/wrapper"
	"net/http"
)

func LoadDefaultCommands(s *SlackReceiver) {
	// create a table of commands to itterate over?
	if err := LoadHelpCommand(s); err != nil {
		fmt.Errorf("Failed to load `help` command into server: %s", err)
	}
}

func LoadHelpCommand(s *SlackReceiver) error {
	f := func(w http.ResponseWriter, r *http.Request) {
		sc, err := slack.SlashCommandParse(r)
		if err != nil {
			fmt.Errorf("foo")
			return
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
			fmt.Errorf("foo")
			return
		}
	}

	r := Route{
		Path:    "/slack/command/help",
		Handler: f,
	}

	if err := s.AddRoute(&r); err != nil {
		return err
	}

	return nil
}
