package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

func TestRoutes(t *testing.T) {
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	r := &Route{
		Path:    "/foo",
		Handler: f,
	}
	if err := s.AddRoute(r); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if x, ok := s.routes["/foo"]; ok {
		if x != r {
			t.Fatal("Expected to find a matching handler in routes")
		}
	} else {
		t.Fatal("Expected to find /foo in routes")
	}
	if err := s.AddRoute(r); err == nil {
		t.Fatalf("Hmm that should have errored")
	}
	if err := s.RemoveRoute(r); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if len(s.routes) != 0 {
		t.Fatalf("Expected zero routes to be configured, found %d", len(s.routes))
	}
}

func TestHttp(t *testing.T) {
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "Hello World")
		return nil
	}
	r := &Route{
		Path:    "/foo",
		Handler: f,
	}
	if err := s.AddRoute(r); err != nil {
		t.Fatalf("Unexpected error adding route: %s", err)
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
	r := &Route{
		Path:    "/bar",
		Handler: f,
	}
	if err := s.AddRoute(r); err != nil {
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
