package test

import "net/http"

func MakeMockClient(rtf http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: rtf,
	}
}

func MakeOkResponse() *http.Response {
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
