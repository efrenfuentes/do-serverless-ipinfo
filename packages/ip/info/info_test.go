package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestMainLookup(t *testing.T) {
	t.Setenv("IPINFO_ACCESS_TOKEN", "test-token")
	original := client.Transport
	t.Cleanup(func() { client.Transport = original })
	const success = `{"ip":"2001:db8::1"}`
	for _, tc := range []struct {
		name     string
		ip       string
		upstream int
		body     string
		err      error
		want     int
	}{
		{"invalid IP", "bad", 0, "", nil, 400},
		{"IPv6", " 2001:db8::1 ", 200, success, nil, 200},
		{"rate limited", "192.0.2.1", 429, "", nil, 429},
		{"provider error", "192.0.2.1", 500, "", nil, 502},
		{"invalid response", "192.0.2.1", 200, "bad", nil, 502},
		{"network error", "192.0.2.1", 0, "", errors.New("offline"), 502},
		{"timeout", "192.0.2.1", 0, "", context.DeadlineExceeded, 504},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client.Transport = transportFunc(func(req *http.Request) (*http.Response, error) {
				if tc.want == 400 {
					t.Fatal("invalid IP triggered a lookup")
				}
				if req.URL.String() != "https://ipinfo.io/"+strings.TrimSpace(tc.ip)+"/json" || req.Header.Get("Authorization") != "Bearer test-token" {
					t.Fatal("unexpected lookup URL or credentials")
				}
				if _, ok := req.Context().Deadline(); !ok {
					t.Fatal("missing lookup deadline")
				}
				if tc.err != nil {
					return nil, tc.err
				}
				return &http.Response{StatusCode: tc.upstream, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			resp, err := Main(Request{IP: tc.ip})
			if err != nil || resp.StatusCode != tc.want {
				t.Fatalf("response = %+v, err = %v", resp, err)
			}
			if tc.want == 200 && resp.Body != success {
				t.Fatalf("unexpected body: %s", resp.Body)
			}
		})
	}
}
