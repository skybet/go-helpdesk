package handlers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/nlopes/slack"
	"github.com/skybet/go-helpdesk/mocks"
	"github.com/skybet/go-helpdesk/server"
	"github.com/stretchr/testify/mock"
)

func TestHelpCallback(t *testing.T) {
	tt := []struct {
		name       string
		jsonString interface{}
		err        error
	}{
		{
			"All OK, Should Pass",
			"{\"type\": \"dialog_submission\",\"submission\": {\"name\": \"Sigourney Dreamweaver\",\"email\": \"sigdre@example.com\",\"phone\": \"+1 800-555-1212\",\"meal\": \"burrito\",\"comment\": \"No sour cream please\",\"team_channel\": \"C0LFFBKPB\",\"who_should_sing\": \"U0MJRG1AL\"},\"callback_id\": \"employee_offsite_1138b\",\"team\": {\"id\": \"T1ABCD2E12\",\"domain\": \"coverbands\"},\"user\": {\"id\": \"W12A3BCDEF\",\"name\": \"dreamweaver\"},\"channel\": {\"id\": \"C1AB2C3DE\",\"name\": \"coverthon-1999\"},\"action_ts\": \"936893340.702759\",\"token\": \"TOKEN\",\"response_url\": \"https://hooks.slack.com/app/T012AB0A1/123456789/JpmK0yzoZDeRiqfeduTBYXWQ\"}",
			nil,
		},
		{
			"Type Failure",
			42,
			errors.New("Expected a string to be passed to the handler"),
		},
		{
			"Invalid Json",
			"{\"type\":",
			errors.New("unexpected end of JSON input"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(T *testing.T) {

			r := httptest.NewRequest("POST", "/slack", nil)
			w := httptest.NewRecorder()
			req := &server.Request{Request: r}
			res := &server.Response{w}

			err := HelpCallback(res, req, tc.jsonString)
			if err != nil {
				if err.Error() == tc.err.Error() {
					// working as intended
				} else {
					t.Errorf("Test Name: %s - Should result in: %s - Got: %s", tc.name, tc.err, err)
				}
			}
		})
	}
}

func TestHelpRequest(t *testing.T) {
	mockSlack := &mocks.SlackWrapper{}
	mockSlack.On("OpenDialog", "ABC123", mock.Anything).Return(nil)
	Init(mockSlack)
	sc := slack.SlashCommand{TriggerID: "ABC123"}
	r := httptest.NewRequest("POST", "/slack", nil)
	w := httptest.NewRecorder()
	req := &server.Request{Request: r}
	res := &server.Response{w}

	err := HelpRequest(res, req, sc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestHelpRequestErrors(t *testing.T) {
	mockSlack := &mocks.SlackWrapper{}
	mockSlack.On("OpenDialog", "ABC123", mock.Anything).Return(errors.New("bad thing happen"))
	Init(mockSlack)
	sc := slack.SlashCommand{TriggerID: "ABC123"}
	r := httptest.NewRequest("POST", "/slack", nil)
	w := httptest.NewRecorder()
	req := &server.Request{Request: r}
	res := &server.Response{w}

	err := HelpRequest(res, req, "foobar")
	if err == nil {
		t.Fatal("I expected that to error")
	}

	err = HelpRequest(res, req, sc)
	if err == nil {
		t.Fatal("I expected that to error")
	}
}
