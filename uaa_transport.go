package uaa

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

type UaaTransport struct {
	Transport      *http.RoundTripper
	LoggingEnabled bool
}

func (t *UaaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.logRequest(req)

	authHeader := req.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "basic") {
		req.Header.Add("X-CF-ENCODED-CREDENTIALS", "true")
	}

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp)

	return resp, err
}

func NewUaaTransport(loggingEnabled bool) *UaaTransport {
	return &UaaTransport{LoggingEnabled: loggingEnabled}
}

func (t *UaaTransport) logRequest(req *http.Request) {
	if t.LoggingEnabled {
		bytes, _ := httputil.DumpRequest(req, false)
		fmt.Printf(string(bytes))
	}
}

func (t *UaaTransport) logResponse(resp *http.Response) {
	if t.LoggingEnabled {
		bytes, _ := httputil.DumpResponse(resp, true)
		fmt.Printf(string(bytes))
	}
}

func (t *UaaTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return *t.Transport
	}

	return http.DefaultTransport
}
