package test

import (
	"bytes"
	"net/http"
	"os"
	"testing"
)

func NewMockClient(rtf http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: rtf,
	}
}

func NewOkResponse() *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
}

// RoundTripFunc is a helper type for creating mock HTTP transports
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func CaptureStdout(t *testing.T, fn func() error) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = fn()
	if err != nil {
		t.Fatal(err)
	}

	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatal(err)
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
