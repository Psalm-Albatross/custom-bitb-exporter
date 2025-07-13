package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBoolToString(t *testing.T) {
	if boolToString(true) != "true" || boolToString(false) != "false" {
		t.Errorf("boolToString failed")
	}
}

func TestParseRFC3339ToUnix(t *testing.T) {
	ts, err := parseRFC3339ToUnix("2024-07-13T12:34:56+00:00")
	if err != nil || ts == 0 {
		t.Errorf("Failed to parse valid RFC3339 timestamp")
	}
}

func TestCollector_HandlesAPILimit(t *testing.T) {
	// Mock Bitbucket API rate limit endpoint
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/2.0/workspaces/testws/rate-limits/" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"limits": map[string]interface{}{
					"core": map[string]interface{}{
						"remaining": 5,
						"reset":     123,
					},
				},
			})
			return
		}
		w.WriteHeader(404)
	})
	ts := httptest.NewServer(h)
	defer ts.Close()
	// You would inject ts.URL into your BitbucketClient for testing
	// and call the collector.Collect method, then check the metrics output.
}
