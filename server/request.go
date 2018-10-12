package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/nlopes/slack"
	"regexp"
)

// Request wraps http.Request
type Request struct {
	*http.Request
	payload CallbackPayload
}

// Validate the request comes from Slack
func (r *Request) Validate(secret string, dnHeader *string) error {
	// If a dnHeader has been provided, check that the header contains the slack CN
	if dnHeader != nil {
		slackDNHeader := r.Header.Get(*dnHeader)
		dnError := fmt.Errorf("invalid CN in DN header")

		r, _ := regexp.Compile("CN=(.*?),")
		cn := r.FindStringSubmatch(slackDNHeader)
		if len(cn) != 2 {		// It should match the CN exactly one, and contain the CN value as a group
			return dnError
		}

		if cn[1] != "platform-tls-client.slack.com" {
			return dnError
		}
	}

	slackTimestampHeader := r.Header.Get("X-Slack-Request-Timestamp")
	slackTimestamp, err := strconv.ParseInt(slackTimestampHeader, 10, 64)

	// Abort if timestamp is invalid
	if err != nil {
		return fmt.Errorf("invalid timestamp sent from slack: %s", err)
	}

	// Abort if timestamp is stale (older than 5 minutes)
	now := int64(time.Now().Unix())
	if (now - slackTimestamp) > (60 * 5) {
		return fmt.Errorf("stale timestamp sent from slack: %s", err)
	}

	// Abort if request body is invalid
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("invalid request body sent from slack: %s", err)
	}
	slackBody := string(body)

	// Abort if the signature does not correspond to the signing secret
	slackBaseStr := []byte(fmt.Sprintf("v0:%d:%s", slackTimestamp, slackBody))
	slackSignature := r.Header.Get("X-Slack-Signature")
	sec := hmac.New(sha256.New, []byte(secret))
	sec.Write(slackBaseStr)
	mySig := fmt.Sprintf("v0=%s", []byte(hex.EncodeToString(sec.Sum(nil))))
	if mySig != slackSignature {
		return errors.New("invalid signature sent from slack")
	}
	// All good! The request is valid
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return nil
}

// CallbackPayload returns the parsed payload if it exists and is valid
func (r *Request) CallbackPayload() (CallbackPayload, error) {
	if r.payload == nil {
		if err := r.parsePayload(); err != nil {
			return nil, err
		}
		if err := r.payload.Validate(); err != nil {
			return nil, err
		}
	}
	return r.payload, nil
}

func (r *Request) parsePayload() error {
	var payload CallbackPayload
	j := r.Form.Get("payload")
	if j == "" {
		return errors.New("empty payload")
	}
	if err := json.Unmarshal([]byte(j), &payload); err != nil {
		return fmt.Errorf("error parsing payload JSON: %s", err)
	}
	r.payload = payload
	return nil
}

// CallbackPayload represents the data sent by Slack on a user initiated event
type CallbackPayload map[string]interface{}

// Validate the payload
func (c CallbackPayload) Validate() error {
	var errs []string
	if c["type"] == nil {
		errs = append(errs, "Missing value for 'type' key")
	}
	if c["callback_id"] == nil {
		errs = append(errs, "Missing value for 'callback_id' key")
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, ", "))
	}
	return nil
}

// MatchRoute determines if we can route this request based on the payload
func (c CallbackPayload) MatchRoute(r *Route) bool {
	if c["type"] == r.InteractionType && c["callback_id"] == r.CallbackID {
		return true
	}
	return false
}

// Mutate the payload into a go type matching it's type field
func (c CallbackPayload) Mutate() (interface{}, error) {
	switch c["type"] {
	case "dialog_submission":
		var result slack.DialogCallback
		err := mapstructure.Decode(c, &result)
		return &result, err
	case "dialog_suggestion":
		var result slack.DialogSuggestionCallback
		err := mapstructure.Decode(c, &result)
		return &result, err
	default:
		return c, nil
	}
}
