package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMainHeaders(t *testing.T) {
	original := client.Transport
	t.Cleanup(func() { client.Transport = original })
	type event = map[string]interface{}
	for _, tc := range []struct {
		name   string
		args   event
		wantIP string
	}{
		{"current", event{"http": event{"headers": event{"x-forwarded-for": " 2001:db8::1 , 192.0.2.1"}}}, "2001:db8::1"},
		{"legacy", event{"__ow_headers": event{"x-forwarded-for": "192.0.2.1"}}, "192.0.2.1"},
		{"missing", nil, ""},
		{"wrong headers type", event{"__ow_headers": true}, ""},
		{"missing header", event{"http": event{"headers": event{}}}, ""},
		{"wrong header type", event{"__ow_headers": event{"x-forwarded-for": 42}}, ""},
		{"invalid first address", event{"__ow_headers": event{"x-forwarded-for": "bad, 192.0.2.1"}}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
				if tc.wantIP == "" || r.URL.String() != "https://ipinfo.io/"+tc.wantIP+"/json" {
					t.Fatalf("unexpected lookup: %v", r.URL)
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"country":"VE"}`))}, nil
			})
			resp, err := Main(tc.args)
			if err != nil {
				t.Fatal(err)
			}
			var body IPResponse
			if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
				t.Fatal(err)
			}
			if tc.wantIP == "" {
				if resp.StatusCode != 400 || body.Error == "" {
					t.Fatalf("unexpected response: %+v", resp)
				}
			} else if resp.StatusCode != 200 || body.IP != tc.wantIP || string(body.Info) != `{"country":"VE"}` || (body.Family == "ipv6") != strings.Contains(tc.wantIP, ":") {
				t.Fatalf("unexpected response: %+v", resp)
			}
		})
	}
}

func TestLookupFailureKeepsIP(t *testing.T) {
	original := client.Transport
	t.Cleanup(func() { client.Transport = original })
	client.Transport = transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 429, Body: http.NoBody}, nil
	})
	resp, err := Main(map[string]interface{}{"__ow_headers": map[string]interface{}{"x-forwarded-for": "192.0.2.1"}})
	if err != nil {
		t.Fatal(err)
	}
	var got IPResponse
	if err := json.Unmarshal([]byte(resp.Body), &got); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || got.IP != "192.0.2.1" || got.Family != "ipv4" || string(got.Info) != "null" || got.Error == "" {
		t.Fatalf("unexpected response: %s", resp.Body)
	}
}
