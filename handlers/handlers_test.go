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
