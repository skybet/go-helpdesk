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
)

// Request wraps http.Request
type Request struct {
	*http.Request
	payload CallbackPayload
}

// Validate the request comes from Slack
func (r *Request) Validate(secret string) error {
	slackTimestampHeader := r.Header.Get("X-Slack-Request-Timestamp")
	slackTimestamp, err := strconv.ParseInt(slackTimestampHeader, 10, 64)

	// Abort if timestamp is invalid
	if err != nil {
		return fmt.Errorf("Invalid timestamp sent from slack: %s", err)
	}

	// Abort if timestamp is stale (older than 5 minutes)
	now := int64(time.Now().Unix())
	if (now - slackTimestamp) > (60 * 5) {
		return fmt.Errorf("Stale timestamp sent from slack: %s", err)
	}

	// Abort if request body is invalid
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Invalid request body sent from slack: %s", err)
	}
	slackBody := string(body)

	// Abort if the signature does not correspond to the signing secret
	slackBaseStr := []byte(fmt.Sprintf("v0:%d:%s", slackTimestamp, slackBody))
	slackSignature := r.Header.Get("X-Slack-Signature")
	sec := hmac.New(sha256.New, []byte(secret))
	sec.Write(slackBaseStr)
	mySig := fmt.Sprintf("v0=%s", []byte(hex.EncodeToString(sec.Sum(nil))))
	if mySig != slackSignature {
		return errors.New("Invalid signature sent from slack")
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
		return errors.New("Empty payload")
	}
	if err := json.Unmarshal([]byte(j), &payload); err != nil {
		return fmt.Errorf("Error parsing payload JSON: %s", err)
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
		return fmt.Errorf("Error(s) with callback payload: %s", strings.Join(errs, ", "))
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
