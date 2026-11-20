package main

import (
	"encoding/json"
	"testing"
)

func TestMainHeaders(t *testing.T) {
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
			resp, err := Main(tc.args)
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]string
			if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
				t.Fatal(err)
			}
			if tc.wantIP == "" {
				if resp.StatusCode != 400 || body["error"] == "" {
					t.Fatalf("unexpected response: %+v", resp)
				}
			} else if resp.StatusCode != 200 || body["ip"] != tc.wantIP {
				t.Fatalf("unexpected response: %+v", resp)
			}
		})
	}
}
