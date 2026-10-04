package llmjudge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ClientConfig struct {
	APIKey, Endpoint                   string
	Timeout, Budget, Backoff, Interval time.Duration
	Retries                            int
}
type Client struct {
	c    ClientConfig
	http *http.Client
	last time.Time
}
type Call struct {
	Scores                           Scores
	Model                            string
	Raw                              []byte
	Usage                            Usage
	Attempts, UnknownBillingAttempts int
}
type APIError struct {
	Status int
	Fatal  bool
	Detail string
}

func (e *APIError) Error() string { return fmt.Sprintf("LLM HTTP %d: %s", e.Status, e.Detail) }
func Fatal(err error) bool        { var e *APIError; return errors.As(err, &e) && e.Fatal }
func NewClient(c ClientConfig) (*Client, error) {
	if strings.TrimSpace(c.APIKey) == "" || strings.ContainsAny(c.APIKey, "\r\n") {
		return nil, errors.New("OPENAI_API_KEY absent/invalid")
	}
	if c.Endpoint == "" {
		c.Endpoint = "https://api.openai.com/v1/responses"
	}
	u, e := url.Parse(c.Endpoint)
	if e != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !((u.Scheme == "https" && u.Host == "api.openai.com" && u.Path == "/v1/responses") || (u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1"))) {
		return nil, errors.New("only official OpenAI Responses endpoint or loopback test server allowed")
	}
	if c.Timeout <= 0 || c.Budget <= 0 || c.Backoff <= 0 || c.Interval < 0 || c.Retries < 0 || c.Retries > 6 {
		return nil, errors.New("invalid timeout/retry configuration")
	}
	return &Client{c: c, http: &http.Client{Timeout: c.Timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func pause(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
func retryAfter(h http.Header) time.Duration {
	if n, e := strconv.ParseFloat(h.Get("Retry-After"), 64); e == nil && n > 0 {
		return time.Duration(min(n, 30) * float64(time.Second))
	}
	if t, e := http.ParseTime(h.Get("Retry-After")); e == nil {
		return min(max(time.Until(t), 0), 30*time.Second)
	}
	return 0
}
func (c *Client) Evaluate(ctx context.Context, model string, in Input) (Call, error) {
	var call Call
	body, e := Request(model, in)
	if e != nil {
		return call, e
	}
	ctx, cancel := context.WithTimeout(ctx, c.c.Budget)
	defer cancel()
	for attempt := 0; attempt <= c.c.Retries; attempt++ {
		if e = pause(ctx, time.Until(c.last.Add(c.c.Interval))); e != nil {
			return call, e
		}
		req, e := http.NewRequestWithContext(ctx, http.MethodPost, c.c.Endpoint, bytes.NewReader(body))
		if e != nil {
			return call, errors.New("request setup failed")
		}
		req.Header.Set("Authorization", "Bearer "+c.c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		c.last = time.Now()
		call.Attempts++
		resp, err := c.http.Do(req)
		wait := c.c.Backoff * time.Duration(1<<attempt)
		if err != nil {
			call.UnknownBillingAttempts++
			if ctx.Err() != nil {
				return call, ctx.Err()
			}
			e = errors.New("LLM transport failure (provider details suppressed)")
		} else {
			raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024+1))
			resp.Body.Close()
			if readErr != nil || len(raw) > 2*1024*1024 {
				call.UnknownBillingAttempts++
				return call, errors.New("LLM response read/size failure")
			}
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				raw = bytes.ReplaceAll(raw, []byte(c.c.APIKey), []byte("[REDACTED]"))
				call.Raw = raw
				e = parseResponse(&call, raw, model)
				return call, e
			}
			retry := resp.StatusCode == 429 || resp.StatusCode == 408 || resp.StatusCode >= 500 && resp.StatusCode <= 599
			e = &APIError{Status: resp.StatusCode, Fatal: !retry, Detail: "request rejected; provider body suppressed"}
			if !retry {
				return call, e
			}
			if resp.StatusCode != 429 {
				call.UnknownBillingAttempts++
			}
			wait = max(wait, retryAfter(resp.Header))
		}
		if attempt == c.c.Retries {
			return call, e
		}
		if e = pause(ctx, min(wait, 30*time.Second)); e != nil {
			return call, e
		}
	}
	return call, errors.New("retry exhausted")
}
func parseResponse(call *Call, raw []byte, model string) error {
	var r struct {
		Model, Status string
		Usage         *struct {
			Input   *int64 `json:"input_tokens"`
			Output  *int64 `json:"output_tokens"`
			Details struct {
				Cached int64 `json:"cached_tokens"`
			} `json:"input_tokens_details"`
		}
		Output []struct {
			Type    string
			Content []struct{ Type, Text string }
		}
	}
	if json.Unmarshal(raw, &r) != nil {
		return errors.New("malformed Responses JSON")
	}
	call.Model = r.Model
	if r.Usage != nil && r.Usage.Input != nil && r.Usage.Output != nil {
		i, o, k := *r.Usage.Input, *r.Usage.Output, r.Usage.Details.Cached
		if i < 0 || o < 0 || k < 0 || k > i {
			return errors.New("invalid token usage")
		}
		call.Usage = Usage{i, o, k, true}
	}
	if r.Model != model {
		return &APIError{Fatal: true, Detail: "resolved model differs from pinned snapshot"}
	}
	if r.Status != "completed" {
		return errors.New("LLM response incomplete; no evaluation saved")
	}
	text := ""
	for _, out := range r.Output {
		if out.Type != "message" {
			continue
		}
		for _, item := range out.Content {
			if item.Type == "refusal" {
				return errors.New("LLM refusal; no evaluation saved")
			}
			if item.Type == "output_text" {
				text += item.Text
			}
		}
	}
	s, e := ParseScores([]byte(text))
	if e != nil {
		return e
	}
	call.Scores = s
	return nil
}
