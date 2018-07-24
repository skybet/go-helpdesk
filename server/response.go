package server

import (
	"fmt"
	"io"
	"net/http"
)

// Response wraps http.ResponseWriter
type Response struct {
	http.ResponseWriter
}

// Text is a convenience method for sending a response
func (r *Response) Text(code int, body string) {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(code)
	io.WriteString(r, fmt.Sprintf("%s\n", body))
}
