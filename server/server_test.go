package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

func TestSlashCommands(t *testing.T) {
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	sc := &SlashCommand{
		Path:    "/foo",
		Handler: f,
	}
	if err := s.AddSlashCommand(sc); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if x, ok := s.slashCommands["/foo"]; ok {
		if x != sc {
			t.Fatal("Expected to find a matching handler in slash commands")
		}
	} else {
		t.Fatal("Expected to find /foo in slash commands")
	}
	if err := s.AddSlashCommand(sc); err == nil {
		t.Fatalf("Hmm that should have errored")
	}
	if err := s.RemoveSlashCommand(sc); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if err := s.RemoveSlashCommand(sc); err == nil {
		t.Fatal("Expected this to error")
	}
	if len(s.slashCommands) != 0 {
		t.Fatalf("Expected zero slash commands to be configured, found %d", len(s.slashCommands))
	}
}

func TestHttp(t *testing.T) {
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "Hello World")
		return nil
	}
	sc := &SlashCommand{
		Path:    "/foo",
		Handler: f,
	}
	if err := s.AddSlashCommand(sc); err != nil {
		t.Fatalf("Unexpected error adding slash command: %s", err)
	}
	go func() {
		if err := s.Start(":8080", nil); err != nil {
			t.Fatalf("Unexpected error starting server: %s", err)
		}
	}()
	resp, err := http.Get("http://localhost:8080/foo")
	if err != nil {
		t.Fatalf("Unexpected error making request: %s", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Expected a 200 got %d", resp.StatusCode)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	str := buf.String()
	if str != "Hello World" {
		t.Fatalf("Unexpected response: %s", str)
	}
}

func TestErrors(t *testing.T) {
	var errString string
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("WAH")
	}
	sc := &SlashCommand{
		Path:    "/bar",
		Handler: f,
	}
	if err := s.AddSlashCommand(sc); err != nil {
		t.Fatalf("Unexpected error adding route: %s", err)
	}
	log := func(msg string, i ...interface{}) {
		errString = fmt.Sprintf(msg, i[0])
	}
	go func() {
		if err := s.Start(":8081", log); err != nil {
			t.Fatalf("Unexpected error starting server: %s", err)
		}
	}()
	_, err := http.Get("http://localhost:8081/bar")
	if err != nil {
		t.Fatalf("Unexpected error making request: %s", err)
	}
	if errString != "HTTP Error: WAH" {
		t.Fatalf("Expecting an error 'WAH' got '%s'", errString)
	}
}
