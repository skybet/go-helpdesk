package wrapper

import (
	"github.com/adampointer/go-slackbot"
	"github.com/nlopes/slack"

	"fmt"
)

var (
	SlackApp *slack.Client
	SlackBot *slackbot.Bot
)

// Init creates connections to Slack
func Init(appToken, botToken string) error {
	SlackApp = slack.New(appToken)
	SlackBot = slackbot.New(botToken)

	// Check tokens are valid
	_, err := SlackApp.AuthTest()
	if err != nil {
		return err
	}

	_, err = SlackBot.Client.AuthTest()
	return err
}

// OpenDialog opens a Dialog inside Slack
func OpenDialog(triggerId string, dialog slack.Dialog) error {
	err := SlackApp.OpenDialog(triggerId, dialog)
	if err != nil {
		fmt.Errorf("Error opening dialog. ", err)
		return err
	}
	return err
}

// SendMessage posts a message to Slack that is visible to everyone in the channel
func SendMessage(message, channel string) {
	msg := SlackBot.RTM.NewOutgoingMessage(message, fmt.Sprintf("#%s", channel))
	SlackBot.RTM.SendMessage(msg)
}
