package slack

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
)

// Added as a var so that we can change this for testing purposes
var (
	SLACK_API            string = "https://slack.com/api/"
	SLACK_WEB_API_FORMAT string = "https://%s.slack.com/api/users.admin.%s?t=%s"
)

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type AuthTestResponse struct {
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

type ClientConfig struct {
	BatchPresenceAware bool // Only deliver presence events when requested by subscription
	IncludeLocale      bool // Set this to true to receive the locale for users and channels
	MPIMAware          bool // Returns MPIMs to the client in the API response
	NoLatest           bool // Exclude latest timestamps for channels, groups, mpims, and ims
	NoUnreads          bool // Skip unread counts for each channel (improves performance)
	SimpleLatest       bool // Return timestamp only for latest message object of each channel (improves performance)
	PresenceSub        bool // Set this to true if you plan to subscribe to presence events

	token string // Authentication token bearing required scopes
}

type Client struct {
	Config *ClientConfig
	debug  bool
}

// New creates new Client.
func New(token string) *Client {
	return &Client{
		Config: &ClientConfig{
			token:              token,
			BatchPresenceAware: true,
			PresenceSub:        true,
		},
		debug: false,
	}
}

type authTestResponseFull struct {
	SlackResponse
	AuthTestResponse
}

// AuthTest tests if the user is able to do authenticated requests or not
func (api *Client) AuthTest() (response *AuthTestResponse, error error) {
	return api.AuthTestContext(context.Background())
}

// AuthTestContext tests if the user is able to do authenticated requests or not with a custom context
func (api *Client) AuthTestContext(ctx context.Context) (response *AuthTestResponse, error error) {
	api.Debugf("Challenging auth...")
	responseFull := &authTestResponseFull{}
	err := post(ctx, "auth.test", url.Values{"token": {api.Config.token}}, responseFull, api.debug)
	if err != nil {
		api.Debugf("failed to test for auth: %s", err)
		return nil, err
	}
	if !responseFull.Ok {
		api.Debugf("auth response was not Ok: %s", responseFull.Error)
		return nil, errors.New(responseFull.Error)
	}

	api.Debugf("Auth challenge was successful with response %+v", responseFull.AuthTestResponse)
	return &responseFull.AuthTestResponse, nil
}

// SetDebug switches the api into debug mode
// When in debug mode, it logs various info about what its doing
// If you ever use this in production, don't call SetDebug(true)
func (api *Client) SetDebug(debug bool) {
	api.debug = debug
	if debug && logger == nil {
		SetLogger(log.New(os.Stdout, "nlopes/slack", log.LstdFlags|log.Lshortfile))
	}
}

// Debugf print a formatted debug line.
func (api *Client) Debugf(format string, v ...interface{}) {
	if api.debug {
		logger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debugln print a debug line.
func (api *Client) Debugln(v ...interface{}) {
	if api.debug {
		logger.Output(2, fmt.Sprintln(v...))
	}
}

func (config *ClientConfig) toParams() url.Values {
	arguments := url.Values{}

	arguments.Add("token", config.token)

	if config.BatchPresenceAware {
		arguments.Add("batch_presence_aware", "1")
	}

	if config.IncludeLocale {
		arguments.Add("include_locale", "1")
	}

	if config.MPIMAware {
		arguments.Add("mpim_aware", "1")
	}

	if config.NoLatest {
		arguments.Add("no_latest", "1")
		arguments.Add("no_unreads", "1")
	}

	if config.NoUnreads {
		arguments.Add("no_unreads", "1")
	}

	if config.SimpleLatest {
		arguments.Add("simple_latest", "1")
	}

	if config.PresenceSub {
		arguments.Add("presence_sub", "1")
	}

	return arguments
}
