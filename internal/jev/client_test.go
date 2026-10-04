package jev

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func clientFixture(t *testing.T, h http.HandlerFunc, retries int) *Client {
	t.Helper()
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	c, err := NewClient(ClientConfig{APIKey: "very-secret-key", BaseURL: server.URL + "/v1", Timeout: time.Second, Backoff: time.Millisecond, Budget: time.Second, Retries: retries})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestClientRequestAndRetry(t *testing.T) {
	calls := 0
	c := clientFixture(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/v1/systemone" || r.Header.Get("Authorization") != "Bearer very-secret-key" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("headers/path")
		}
		var req Request
		if json.NewDecoder(r.Body).Decode(&req) != nil || len(req.Questions) != 7 || req.Model != DefaultModel {
			t.Error("request JSON")
		}
		if calls == 1 {
			w.Header().Set("retry-after-ms", "1")
			w.WriteHeader(429)
			return
		}
		if calls == 2 {
			w.WriteHeader(529)
			return
		}
		json.NewEncoder(w).Encode(fixtureResponse())
	}, 2)
	r, err := c.Evaluate(context.Background(), BuildRequest(DefaultModel, EvaluationState{}))
	if err != nil || r.Attempts != 3 || r.Response.Usage.InputTokens == nil {
		t.Fatalf("result %+v err %v", r, err)
	}
}
func TestAuthSchemaAndRetryBounds(t *testing.T) {
	for _, status := range []int{401, 402, 403, 404, 422, 500, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := clientFixture(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(status)
				fmt.Fprint(w, "very-secret-key")
			}, 2)
			_, err := c.Evaluate(context.Background(), BuildRequest(DefaultModel, EvaluationState{}))
			expected := 1
			if status >= 500 {
				expected = 3
			}
			if err == nil || calls != expected || strings.Contains(err.Error(), "very-secret-key") {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
			if status < 500 && !IsFatal(err) {
				t.Fatal("not fail-fast")
			}
		})
	}
}
func TestMalformedResponseNotRepeated(t *testing.T) {
	calls := 0
	c := clientFixture(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, `{"model":"jev-1.13.0","answers":{},"usage":{"input_tokens":11}}`)
	}, 3)
	r, err := c.Evaluate(context.Background(), BuildRequest(DefaultModel, EvaluationState{}))
	if err == nil || calls != 1 || r.Response.Usage.InputTokens == nil || *r.Response.Usage.InputTokens != 11 {
		t.Fatal("malformed response should fail once and retain reported usage")
	}
}
func TestTimeoutAndCancellation(t *testing.T) {
	c := clientFixture(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(100 * time.Millisecond):
		}
	}, 0)
	c.http.Timeout = 10 * time.Millisecond
	r, err := c.Evaluate(context.Background(), BuildRequest(DefaultModel, EvaluationState{}))
	if err == nil || r.UnknownBillingAttempts != 1 || strings.Contains(err.Error(), "very-secret-key") {
		t.Fatal("timeout")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.Evaluate(ctx, BuildRequest(DefaultModel, EvaluationState{}))
	if err != context.Canceled {
		t.Fatal(err)
	}
}
func TestRetryAfter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	h := http.Header{}
	h.Set("Retry-After", "2")
	if delayHeader(h, now) != 2*time.Second {
		t.Fatal("seconds")
	}
	h.Set("Retry-After", now.Add(3*time.Second).Format(http.TimeFormat))
	if delayHeader(h, now) != 3*time.Second {
		t.Fatal("date")
	}
	h.Set("retry-after-ms", "5")
	if delayHeader(h, now) != 5*time.Millisecond {
		t.Fatal("milliseconds")
	}
}
func TestRedirectNeverLeaksAuthorization(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("redirect followed") }))
	defer target.Close()
	c := clientFixture(t, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 302) }, 0)
	if _, err := c.Evaluate(context.Background(), BuildRequest(DefaultModel, EvaluationState{})); err == nil {
		t.Fatal("redirect accepted")
	}
}
