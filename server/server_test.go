package server

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

var (
	slackSecret = "fake_secret"
	basePath    = "/slack"
	errString   string
	logf        = func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	log = func(i ...interface{}) {
		errString = fmt.Sprintf(i[0].(string))
	}
)

func addSlackHeaders(body string, r *http.Request) {
	// Set the timestamp header
	validTime := int(time.Now().Unix())
	timestampHeader := strconv.Itoa(validTime)
	r.Header.Set("X-Slack-Request-Timestamp", timestampHeader)

	// Set the signature header
	slackBaseStr := []byte(fmt.Sprintf("v0:%d:%s", validTime, body))
	h := hmac.New(sha256.New, []byte(slackSecret))
	h.Write(slackBaseStr)
	mySig := fmt.Sprintf("v0=%s", []byte(hex.EncodeToString(h.Sum(nil))))
	r.Header.Set("X-Slack-Signature", mySig)
}

func performGenericRequest(raw, path string, s *SlackHandler) *http.Response {
	body := bytes.NewBufferString(raw)
	req := httptest.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addSlackHeaders(raw, req)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	return w.Result()
}

func TestMatchSlashCommand(t *testing.T) {
	raw := "token=TOKEN&team_id=T01ABC&team_domain=example&channel_id=D8AD0L4UB&channel_name=directmessage&user_id=UABC123&user_name=bob.smith&command=%2Fbob-test&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FABC123%2F123456%2FABC123&trigger_id=400003447986.4709815545.5c0291e01b37fc97ab64d8d7888f6cda"
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
	s := NewSlackHandler(basePath, "TOKEN", slackSecret, log, logf)
	s.HandleCommand("/bob-test", h)
	resp := performGenericRequest(raw, basePath, s)

	if resp.StatusCode != 200 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 200 status. Got '%d'", resp.StatusCode)
	}
}

func TestUnmatchedSlashCommand(t *testing.T) {
	raw := "token=TOKEN&team_id=T01ABC&team_domain=example&channel_id=D8AD0L4UB&channel_name=directmessage&user_id=UABC123&user_name=bob.smith&command=%2Fbob-test&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FABC123%2F123456%2FABC123&trigger_id=400003447986.4709815545.5c0291e01b37fc97ab64d8d7888f6cda"
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
	s := NewSlackHandler(basePath, "TOKEN", slackSecret, log, logf)
	s.HandleCommand("/foobar", h)
	resp := performGenericRequest(raw, basePath, s)

	if resp.StatusCode != 404 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 404 status. Got '%d'", resp.StatusCode)
	}
}

func TestDialogSubmissionEvent(t *testing.T) {

	raw := "payload=%7B%22type%22%3A%20%22dialog_submission%22%2C%22submission%22%3A%20%7B%22name%22%3A%20%22Sigourney%20Dreamweaver%22%2C%22email%22%3A%20%22sigdre%40example.com%22%2C%22phone%22%3A%20%22%2B1%20800-555-1212%22%2C%22meal%22%3A%20%22burrito%22%2C%22comment%22%3A%20%22No%20sour%20cream%20please%22%2C%22team_channel%22%3A%20%22C0LFFBKPB%22%2C%22who_should_sing%22%3A%20%22U0MJRG1AL%22%7D%2C%22callback_id%22%3A%20%22employee_offsite_1138b%22%2C%22team%22%3A%20%7B%22id%22%3A%20%22T1ABCD2E12%22%2C%22domain%22%3A%20%22coverbands%22%7D%2C%22user%22%3A%20%7B%22id%22%3A%20%22W12A3BCDEF%22%2C%22name%22%3A%20%22dreamweaver%22%7D%2C%22channel%22%3A%20%7B%22id%22%3A%20%22C1AB2C3DE%22%2C%22name%22%3A%20%22coverthon-1999%22%7D%2C%22action_ts%22%3A%20%22936893340.702759%22%2C%22token%22%3A%20%22TOKEN%22%2C%22response_url%22%3A%20%22https%3A%2F%2Fhooks.slack.com%2Fapp%2FT012AB0A1%2F123456789%2FJpmK0yzoZDeRiqfeduTBYXWQ%22%7D"
	h := func(res *Response, req *Request, ctx interface{}) error {
		d, ok := ctx.(*slack.DialogCallback)
		if !ok {
			t.Fatalf("Expected a *slack.DialogCallback to be passed to the handler")
		}
		if d.User.ID != "W12A3BCDEF" {
			t.Fatalf("Unexpected value for User.Id: %s", d.User.ID)
		}
		return nil
	}
	s := NewSlackHandler(basePath, "TOKEN", slackSecret, log, logf)
	s.HandleCallback("dialog_submission", "employee_offsite_1138b", h)
	resp := performGenericRequest(raw, basePath, s)

	if resp.StatusCode != 200 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 200 status. Got '%d'", resp.StatusCode)
	}
}

func TestMalformedActionEvent(t *testing.T) {
	tt := []struct {
		name  string
		raw   string
		sCode int
		err   string
	}{
		{
			"Fail on invalid JSON",
			"payload=ssion%22%3A%20%7B%22name%22%3A%20%22Sigourney%20Dreamweaver%22%2C%22email%22%3A%20%22sigdre%40example.com%22%2C%22phone%22%3A%20%22%2B1%20800-555-1212%22%2C%22meal%22%3A%20%22burrito%22%2C%22comment%22%3A%20%22No%20sour%20cream%20please%22%2C%22team_channel%22%3A%20%22C0LFFBKPB%22%2C%22who_should_sing%22%3A%20%22U0MJRG1AL%22%7D%2C%22callback_id%22%3A%20%22employee_offsite_1138b%22%2C%22team%22%3A%20%7B%22id%22%3A%20%22T1ABCD2E12%22%2C%22domain%22%3A%20%22coverbands%22%7D%2C%22user%22%3A%20%7B%22id%22%3A%20%22W12A3BCDEF%22%2C%22name%22%3A%20%22dreamweaver%22%7D%2C%22channel%22%3A%20%7B%22id%22%3A%20%22C1AB2C3DE%22%2C%22name%22%3A%20%22coverthon-1999%22%7D%2C%22action_ts%22%3A%20%22936893340.702759%22%2C%22token%22%3A%20%22TOKEN%22%2C%22response_url%22%3A%20%22https%3A%2F%2Fhooks.slack.com%2Fapp%2FT012AB0A1%2F123456789%2FJpmK0yzoZDeRiqfeduTBYXWQ%22",
			404,
			"Error parsing payload: Error parsing payload JSON: invalid character 's' looking for beginning of value",
		},
		{
			"Fail on missing value for 'type'",
			"payload=%7B%22submission%22%3A%20%7B%22name%22%3A%20%22Sigourney%20Dreamweaver%22%2C%22email%22%3A%20%22sigdre%40example.com%22%2C%22phone%22%3A%20%22%2B1%20800-555-1212%22%2C%22meal%22%3A%20%22burrito%22%2C%22comment%22%3A%20%22No%20sour%20cream%20please%22%2C%22team_channel%22%3A%20%22C0LFFBKPB%22%2C%22who_should_sing%22%3A%20%22U0MJRG1AL%22%7D%2C%22callback_id%22%3A%20%22employee_offsite_1138b%22%2C%22team%22%3A%20%7B%22id%22%3A%20%22T1ABCD2E12%22%2C%22domain%22%3A%20%22coverbands%22%7D%2C%22user%22%3A%20%7B%22id%22%3A%20%22W12A3BCDEF%22%2C%22name%22%3A%20%22dreamweaver%22%7D%2C%22channel%22%3A%20%7B%22id%22%3A%20%22C1AB2C3DE%22%2C%22name%22%3A%20%22coverthon-1999%22%7D%2C%22action_ts%22%3A%20%22936893340.702759%22%2C%22token%22%3A%20%22TOKEN%22%2C%22response_url%22%3A%20%22https%3A%2F%2Fhooks.slack.com%2Fapp%2FT012AB0A1%2F123456789%2FJpmK0yzoZDeRiqfeduTBYXWQ%22%7D",
			404,
			"Error parsing payload: Missing value for 'type' key",
		},
		{
			"Fail on missing value for 'callback_id'",
			"payload=%7B%22type%22%3A%20%22dialog_submission%22%7D",
			404,
			"Error parsing payload: Missing value for 'callback_id' key",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(T *testing.T) {
			s := NewSlackHandler(basePath, "TOKEN", slackSecret, log, logf)
			resp := performGenericRequest(tc.raw, basePath, s)
			if resp.StatusCode != tc.sCode {
				t.Errorf("Expected a %d status. Got '%d'", tc.sCode, resp.StatusCode)
			}
			if errString != tc.err {
				t.Errorf("Test Name: %s - Should result in: %s - Got: %s", tc.name, tc.err, errString)
			}
		})
	}
}

func TestMatchPath(t *testing.T) {
	h := func(res *Response, req *Request, ctx interface{}) error {
		return nil
	}
	s := NewSlackHandler("/slack", "TOKEN", slackSecret, log, logf)
	s.HandlePath("/foo", h)
	raw := "foo=bar"
	resp := performGenericRequest(raw, "/foo", s)

	if resp.StatusCode != 200 {
		t.Logf("ErrString: %s", errString)
		t.Fatalf("Expected a 200 status. Got '%d'", resp.StatusCode)
	}
}

func TestHandlerErrors(t *testing.T) {
	h := func(res *Response, req *Request, ctx interface{}) error {
		return fmt.Errorf("Serious problem")
	}
	s := NewSlackHandler("/slack", "TOKEN", slackSecret, log, logf)
	s.HandlePath("/foo", h)
	raw := "foo=bar"
	performGenericRequest(raw, "/foo", s)

	if errString != "HTTP handler error: Serious problem" {
		t.Fatalf("Unexpected error string: %s", errString)
	}
}

func TestMissingTimestamp(t *testing.T) {
	s := NewSlackHandler("/slack", "TOKEN", slackSecret, log, logf)
	s.HandlePath("/foo", nil)
	req := httptest.NewRequest("POST", "/foo", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Fatalf("Expected a 400 status. Got '%d'", resp.StatusCode)
	}
	if !strings.HasPrefix(errString, "Bad request from slack: Invalid timestamp sent from slack") {
		t.Fatalf("Unexpected error string: %s", errString)
	}
}

func TestStaleTimestamp(t *testing.T) {
	s := NewSlackHandler("/slack", "TOKEN", slackSecret, log, logf)
	s.HandlePath("/foo", nil)
	req := httptest.NewRequest("POST", "/foo", nil)

	// Set an old timestamp
	staleTime := time.Now().Add(time.Minute * -10)
	timestampHeader := strconv.Itoa(int(staleTime.Unix()))
	req.Header.Set("X-Slack-Request-Timestamp", timestampHeader)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Fatalf("Expected a 400 status. Got '%d'", resp.StatusCode)
	}
	if !strings.HasPrefix(errString, "Bad request from slack: Stale timestamp sent from slack") {
		t.Fatalf("Unexpected error string: %s", errString)
	}
}

func TestInvalidSecret(t *testing.T) {
	s := NewSlackHandler("/slack", "TOKEN", "bad_secret", log, logf)
	s.HandlePath("/foo", nil)
	raw := "text"
	body := bytes.NewBufferString(raw)
	req := httptest.NewRequest("POST", "/foo", body)
	addSlackHeaders(raw, req)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Fatalf("Expected a 400 status. Got '%d'", resp.StatusCode)
	}
	if !strings.HasPrefix(errString, "Bad request from slack: Invalid signature sent from slack") {
		t.Fatalf("Unexpected error string: %s", errString)
	}
}
