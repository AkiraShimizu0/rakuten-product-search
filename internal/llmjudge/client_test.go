package llmjudge

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

func fixture() Scores {
	return Scores{80, 85, 70, 75, 90, 87, 84, "設置条件と手入れの比較が可能。"}
}
func response(s Scores, model string) []byte {
	scores, _ := json.Marshal(s)
	b, _ := json.Marshal(map[string]any{"model": model, "stop_reason": "end_turn", "usage": map[string]any{"input_tokens": 100, "output_tokens": 20, "cache_read_input_tokens": 10, "cache_creation_input_tokens": 5}, "content": []any{map[string]any{"type": "text", "text": string(scores)}}})
	return b
}
func testClient(t *testing.T, server *httptest.Server, retries int) *Client {
	t.Helper()
	c, e := NewClient(ClientConfig{"synthetic-secret", server.URL, time.Second, 2 * time.Second, time.Millisecond, 0, retries})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestRetryStructuredRequestUsage(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("x-api-key") != "synthetic-secret" || r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("missing auth")
		}
		var body map[string]json.RawMessage
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Fatal("request")
		}
		if _, ok := body["tools"]; ok {
			t.Error("tools enabled")
		}
		if !strings.Contains(string(body["output_config"]), `"type":"json_schema"`) {
			t.Error("not strict schema")
		}
		if calls == 1 {
			w.WriteHeader(429)
			fmt.Fprint(w, "synthetic-secret")
			return
		}
		w.Write(response(fixture(), Model))
	}))
	defer server.Close()
	call, e := testClient(t, server, 2).Evaluate(context.Background(), Model, Input{ProductName: "fixture"})
	if e != nil || call.Attempts != 2 || !call.Usage.Known || call.Scores.Overall != 84 {
		t.Fatal(call, e)
	}
	cost := DefaultPrices().Cost(call.Usage)
	if cost == nil || *cost <= 0 {
		t.Fatal("cost")
	}
}
func TestFatalAndExhausted(t *testing.T) {
	for _, status := range []int{401, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			n := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n++
				w.WriteHeader(status)
				fmt.Fprint(w, "synthetic-secret provider detail")
			}))
			defer server.Close()
			call, e := testClient(t, server, 1).Evaluate(context.Background(), Model, Input{})
			if e == nil || strings.Contains(e.Error(), "synthetic-secret") || strings.Contains(e.Error(), "provider detail") {
				t.Fatal(e)
			}
			if status == 401 && (!Fatal(e) || n != 1) {
				t.Fatal(n, e)
			}
			if status == 503 && (n != 2 || call.UnknownBillingAttempts != 2) {
				t.Fatal(n, call, e)
			}
		})
	}
}
func TestScoreValidationAndModelPin(t *testing.T) {
	b, _ := json.Marshal(fixture())
	if _, e := ParseScores(b); e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{`{}`, strings.Replace(string(b), `"comparison_depth":85`, `"comparison_depth":null`, 1), strings.Replace(string(b), `"comparison_depth":85`, `"comparison_depth":101`, 1), strings.Replace(string(b), `"comparison_depth":85`, `"comparison_depth":85.5`, 1)} {
		if _, e := ParseScores([]byte(bad)); e == nil {
			t.Fatal("accepted bad schema")
		}
	}
	var call Call
	if e := parseResponse(&call, response(fixture(), "other"), Model); e == nil || !Fatal(e) {
		t.Fatal("model pin")
	}
}

func TestClaudeUsageCostAndIncomplete(t *testing.T) {
	var call Call
	if err := parseResponse(&call, response(fixture(), Model), Model); err != nil {
		t.Fatal(err)
	}
	if call.Usage.Input != 115 || call.Usage.Cached != 10 || call.Usage.CacheWrite != 5 {
		t.Fatal(call.Usage)
	}
	cost := DefaultPrices().Cost(call.Usage)
	expected := (100.0*2 + 10*.2 + 5*2.5 + 20*10) / 1e6
	if cost == nil || *cost != expected {
		t.Fatal(cost, expected)
	}
	bad := strings.Replace(string(response(fixture(), Model)), "end_turn", "max_tokens", 1)
	if err := parseResponse(&call, []byte(bad), Model); err == nil {
		t.Fatal("accepted truncated output")
	}
	body, err := Request(Model, Input{ProductName: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	if string(request["system"]) != fmt.Sprintf("%q", Rubric) {
		var rubric string
		if json.Unmarshal(request["system"], &rubric) != nil || rubric != Rubric {
			t.Fatal("changed rubric")
		}
	}
	if strings.Contains(string(body), `"minimum"`) || strings.Contains(string(body), `"maximum"`) {
		t.Fatal("unsupported schema constraints")
	}
}
