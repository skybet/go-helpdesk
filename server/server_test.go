package server

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

func TestMatchSlashCommand(t *testing.T) {
	var errString string
	basePath := "/slack"
	raw := "token=TOKEN&team_id=T01ABC&team_domain=example&channel_id=D8AD0L4UB&channel_name=directmessage&user_id=UABC123&user_name=bob.smith&command=%2Fbob-test&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FABC123%2F123456%2FABC123&trigger_id=400003447986.4709815545.5c0291e01b37fc97ab64d8d7888f6cda"
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	h := func(res *Response, req *Request, ctx interface{}) error {
		c, ok := ctx.(slack.SlashCommand)
		if !ok {
			t.Fatalf("Expected slack.SlashCommand to be passed to the handler")
		}
		if c.TeamID != "T01ABC" {
			t.Fatalf("Unexpected value for TeamID: %s", c.TeamID)
		}
		return nil
	}
	s := NewSlackHandler(basePath, "TOKEN", log)
	s.HandleCommand("/bob-test", h)
	body := bytes.NewBufferString(raw)
	req := httptest.NewRequest("POST", basePath, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 200 status. Got '%d'", resp.StatusCode)
	}
}

func TestUnmatchedSlashCommand(t *testing.T) {
	var errString string
	basePath := "/slack"
	raw := "token=TOKEN&team_id=T01ABC&team_domain=example&channel_id=D8AD0L4UB&channel_name=directmessage&user_id=UABC123&user_name=bob.smith&command=%2Fbob-test&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FABC123%2F123456%2FABC123&trigger_id=400003447986.4709815545.5c0291e01b37fc97ab64d8d7888f6cda"
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	h := func(res *Response, req *Request, ctx interface{}) error {
		c, ok := ctx.(slack.SlashCommand)
		if !ok {
			t.Fatalf("Expected slack.SlashCommand to be passed to the handler")
		}
		if c.TeamID != "T01ABC" {
			t.Fatalf("Unexpected value for TeamID: %s", c.TeamID)
		}
		return nil
	}
	s := NewSlackHandler(basePath, "TOKEN", log)
	s.HandleCommand("/foobar", h)
	body := bytes.NewBufferString(raw)
	req := httptest.NewRequest("POST", basePath, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != 404 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 404 status. Got '%d'", resp.StatusCode)
	}
}

func TestMatchActionEvent(t *testing.T) {
	var errString string
	basePath := "/slack"
	raw := "payload=%7B%22type%22%3A%20%22dialog_submission%22%2C%22submission%22%3A%20%7B%22name%22%3A%20%22Sigourney%20Dreamweaver%22%2C%22email%22%3A%20%22sigdre%40example.com%22%2C%22phone%22%3A%20%22%2B1%20800-555-1212%22%2C%22meal%22%3A%20%22burrito%22%2C%22comment%22%3A%20%22No%20sour%20cream%20please%22%2C%22team_channel%22%3A%20%22C0LFFBKPB%22%2C%22who_should_sing%22%3A%20%22U0MJRG1AL%22%7D%2C%22callback_id%22%3A%20%22employee_offsite_1138b%22%2C%22team%22%3A%20%7B%22id%22%3A%20%22T1ABCD2E12%22%2C%22domain%22%3A%20%22coverbands%22%7D%2C%22user%22%3A%20%7B%22id%22%3A%20%22W12A3BCDEF%22%2C%22name%22%3A%20%22dreamweaver%22%7D%2C%22channel%22%3A%20%7B%22id%22%3A%20%22C1AB2C3DE%22%2C%22name%22%3A%20%22coverthon-1999%22%7D%2C%22action_ts%22%3A%20%22936893340.702759%22%2C%22token%22%3A%20%22TOKEN%22%2C%22response_url%22%3A%20%22https%3A%2F%2Fhooks.slack.com%2Fapp%2FT012AB0A1%2F123456789%2FJpmK0yzoZDeRiqfeduTBYXWQ%22%7D"
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	h := func(res *Response, req *Request, ctx interface{}) error {
		m, ok := ctx.(slackevents.MessageAction)
		if !ok {
			t.Fatalf("Expected slackevents.MessageAction to be passed to the handler")
		}
		if m.User.Id != "W12A3BCDEF" {
			t.Fatalf("Unexpected value for User.Id: %s", m.User.Id)
		}
		return nil
	}
	s := NewSlackHandler(basePath, "TOKEN", log)
	s.HandleCallback("employee_offsite_1138b", h)
	body := bytes.NewBufferString(raw)
	req := httptest.NewRequest("POST", basePath, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 200 status. Got '%d'", resp.StatusCode)
	}
}

func TestMalformedActionEvent(t *testing.T) {
	var errString string
	basePath := "/slack"
	raw := "payload=nonsense"
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	s := NewSlackHandler(basePath, "TOKEN", log)
	body := bytes.NewBufferString(raw)
	req := httptest.NewRequest("POST", basePath, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != 404 {
		t.Fatalf("Expected a 404 status. Got '%d'", resp.StatusCode)
	}
	if errString != "Error parsing Slack Event: MessageAction unmarshalling failed" {
		t.Fatalf("Unexpected error string: %s", errString)
	}
}

func TestMatchPath(t *testing.T) {
	var errString string
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	h := func(res *Response, req *Request, ctx interface{}) error {
		return nil
	}
	s := NewSlackHandler("/slack", "TOKEN", log)
	s.HandlePath("/foo", h)
	body := bytes.NewBufferString("foo=bar")
	req := httptest.NewRequest("POST", "/foo", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 200 status. Got '%d'", resp.StatusCode)
	}
}

func TestHandlerErrors(t *testing.T) {
	var errString string
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	h := func(res *Response, req *Request, ctx interface{}) error {
		return fmt.Errorf("Serious problem")
	}
	s := NewSlackHandler("/slack", "TOKEN", log)
	s.HandlePath("/foo", h)
	body := bytes.NewBufferString("foo=bar")
	req := httptest.NewRequest("POST", "/foo", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if errString != "HTTP handler error: Serious problem" {
		t.Fatalf("Unexpected error string: %s", errString)
	}
}
