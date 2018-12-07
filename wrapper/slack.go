package wrapper

import (
	//"github.com/BeepBoopHQ/go-slackbot"
	"github.com/nlopes/slack"

	"fmt"
)

// SlackWrapper is a interface for Slack to enable test double injection
type SlackWrapper interface {
	OpenDialog(triggerID string, dialog slack.Dialog) error
	//SendMessage(message, channel string)
}

// Slack is a wrapper around the Slack App and RTM APIs
type Slack struct {
	App *slack.Client
	Bot *slack.Client
}

// New takes an app and bot token, verifies the connection and
// returns an initialised Slack struct
func New(appToken, botToken string) (*Slack, error) {
	slackApp := slack.New(appToken)
	slackBot := slack.New(botToken)

	// Check tokens are valid
	_, err := slackApp.AuthTest()
	if err != nil {
		return nil, err
	}

	if _, err = slackBot.AuthTest(); err != nil {
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
//
//// SendMessage posts a message to Slack that is visible to everyone in the channel
//func (c slack.Client) SendMessage(channelID, message string, params slack.PostMessageParameters) {
//	s.sendMessage(channelID, message, params, []slack.MsgOption{})
//}
//
//func (s *Slack) SendMessageWithAttachments(channelID, message string, params slack.PostMessageParameters, attachments []slack.Attachment) {
//	var opts []slack.MsgOption
//	for _, attachment := range attachments {
//		opts = append(opts, slack.MsgOptionAttachments(attachment))
//	}
//	s.sendMessage(channelID, message, params, opts)
//}
//
//func (s *Slack) sendMessage(channelID, message string, params slack.PostMessageParameters, opts []slack.MsgOption) {
//	opts = append(opts, slack.MsgOptionText(message, true))
//
//	if params.AsUser {
//		opts = append(opts, slack.MsgOptionAsUser(true))
//		opts = append(opts, slack.MsgOptionUser(params.Username))
//	}
//
//	channelId := fmt.Sprintf("#%s", channelID)
//	_, _, err := s.Bot.PostMessage(channelId, opts...)
//	if err != nil {
//		fmt.Printf("%s\n", err)
//		return
//	}
//}