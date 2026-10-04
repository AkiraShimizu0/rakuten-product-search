package jev

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
	APIKey, BaseURL                    string
	Timeout, Interval, Backoff, Budget time.Duration
	Retries                            int
}
type Client struct {
	config      ClientConfig
	http        *http.Client
	lastRequest time.Time
}
type APIError struct {
	Status  int
	Message string
	Fatal   bool
}

func (e *APIError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("Jev HTTP %d: %s", e.Status, e.Message)
	}
	return "Jev: " + e.Message
}
func IsFatal(err error) bool { var e *APIError; return errors.As(err, &e) && e.Fatal }

type CallResult struct {
	Response               Response
	Raw                    []byte
	Attempts               int
	UnknownBillingAttempts int
}

func NewClient(c ClientConfig) (*Client, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return nil, errors.New("JEV_API_KEY is not configured")
	}
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1"))) {
		return nil, errors.New("JEV_BASE_URL must be an HTTPS API base URL (HTTP allowed only on loopback for tests)")
	}
	if c.Timeout <= 0 || c.Budget <= 0 || c.Backoff <= 0 || c.Interval < 0 || c.Retries < 0 || c.Retries > 6 {
		return nil, errors.New("invalid Jev timeout, retry, interval or budget")
	}
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")
	return &Client{config: c, http: &http.Client{Timeout: c.Timeout, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func delayHeader(h http.Header, now time.Time) time.Duration {
	if s, err := strconv.ParseFloat(h.Get("retry-after-ms"), 64); err == nil && s > 0 {
		if s > 120000 {
			return 3 * time.Minute
		}
		return time.Duration(s * float64(time.Millisecond))
	}
	if s, err := strconv.ParseFloat(h.Get("Retry-After"), 64); err == nil && s > 0 {
		if s > 120 {
			return 3 * time.Minute
		}
		return time.Duration(s * float64(time.Second))
	}
	if t, err := http.ParseTime(h.Get("Retry-After")); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
}
func retryStatus(s int) bool {
	return s == 408 || s == 429 || s == 500 || s == 502 || s == 503 || s == 504 || s == 529
}

// Evaluate is serial and paced. A successful but malformed response is not
// retried automatically because it may already have incurred a charge.
func (c *Client) Evaluate(ctx context.Context, request Request) (CallResult, error) {
	var result CallResult
	body, err := RequestJSON(request)
	if err != nil {
		return result, err
	}
	state, _ := json.Marshal(request.State)
	if len(state) > 24000 || len(body) > 60000 {
		return result, &APIError{Message: "request exceeds conservative byte budget", Fatal: true}
	}
	ctx, cancel := context.WithTimeout(ctx, c.config.Budget)
	defer cancel()
	for attempt := 0; attempt <= c.config.Retries; attempt++ {
		if err := wait(ctx, time.Until(c.lastRequest.Add(c.config.Interval))); err != nil {
			return result, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/systemone", bytes.NewReader(body))
		if err != nil {
			return result, &APIError{Message: "cannot construct request", Fatal: true}
		}
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		req.Header.Set("Content-Type", "application/json")
		c.lastRequest = time.Now()
		result.Attempts++
		resp, err := c.http.Do(req)
		delay := c.config.Backoff * time.Duration(1<<attempt)
		if err != nil {
			result.UnknownBillingAttempts++
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if attempt == c.config.Retries {
				return result, &APIError{Message: "network/timeout retries exhausted"}
			}
		} else {
			raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024+1))
			resp.Body.Close()
			if resp.StatusCode == 200 {
				result.Raw = raw
				if readErr != nil || len(raw) > 4*1024*1024 {
					result.UnknownBillingAttempts++
					return result, &APIError{Message: "unreadable or oversized response"}
				}
				result.Response, err = ParseResponse(raw, request)
				if result.Response.Usage.InputTokens == nil {
					result.UnknownBillingAttempts++
				}
				if err != nil {
					return result, err
				}
				return result, nil
			}
			fatal := resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 402 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 422
			if !retryStatus(resp.StatusCode) || attempt == c.config.Retries {
				return result, &APIError{Status: resp.StatusCode, Message: "request rejected or retry limit reached (response body withheld)", Fatal: fatal}
			}
			retryAfter := delayHeader(resp.Header, time.Now())
			if retryAfter > 2*time.Minute {
				return result, &APIError{Status: resp.StatusCode, Message: "Retry-After exceeds retry budget; resume later"}
			}
			if retryAfter > delay {
				delay = retryAfter
			}
		}
		if delay > 2*time.Minute {
			delay = 2 * time.Minute
		}
		if err = wait(ctx, delay); err != nil {
			return result, err
		}
	}
	return result, &APIError{Message: "retry limit reached"}
}
