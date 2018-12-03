package wrapper

import (
	"github.com/BeepBoopHQ/go-slackbot"
	"github.com/nlopes/slack"

	"fmt"
)

// SlackWrapper is a interface for Slack to enable test double injection
type SlackWrapper interface {
	OpenDialog(triggerID string, dialog slack.Dialog) error
	SendMessage(message, channel string)
}

// Slack is a wrapper around the Slack App and RTM APIs
type Slack struct {
	App *slack.Client
	Bot *slackbot.Bot
}

// New takes an app and bot token, verifies the connection and
// returns an initialised Slack struct
func New(appToken, botToken string) (*Slack, error) {
	slackApp := slack.New(appToken)
	slackBot := slackbot.New(botToken)

	// Check tokens are valid
	_, err := slackApp.AuthTest()
	if err != nil {
		return nil, err
	}

	if _, err = slackBot.Client.AuthTest(); err != nil {
		return nil, err
	}
	return &Slack{App: slackApp, Bot: slackBot}, nil
}

// OpenDialog opens a Dialog inside Slack
func (s *Slack) OpenDialog(triggerID string, dialog slack.Dialog) error {
	err := s.App.OpenDialog(triggerID, dialog)
	if err != nil {
		fmt.Printf("error opening dialog. %s\n", err)
		return err
	}
	return err
}

// SendMessage posts a message to Slack that is visible to everyone in the channel
func (s *Slack) SendMessage(message, channel string) {
	p := slack.PostMessageParameters{}
	s.Bot.Client.PostMessage(fmt.Sprintf("#%s", channel), message, p)
}
