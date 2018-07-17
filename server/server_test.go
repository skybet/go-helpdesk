package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

func TestRoutes(t *testing.T) {
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) {}
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
	if err := s.RemoveRoute(r); err == nil {
		t.Fatal("Expected this to error: %s")
	}
	if len(s.routes) != 0 {
		t.Fatalf("Expected zero routes to be configured, found %d", len(s.routes))
	}
}

func TestHttp(t *testing.T) {
	s := NewSlackReceiver()
	f := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	}
	r := &Route{
		Path:    "/foo",
		Handler: f,
	}
	if err := s.AddRoute(r); err != nil {
		t.Fatalf("Unexpected error adding route: %s", err)
	}
	go func() {
		if err := s.Start(":8080"); err != nil {
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
