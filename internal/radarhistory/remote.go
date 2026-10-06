package radarhistory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jev-money-engine/internal/gate"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func readGate(path string) (gate.Config, error) { return gate.Read(path) }

type R2 struct {
	Base, Account, Bucket, Token string
	HTTP                         *http.Client
}

func (s *R2) endpoint(key string) string {
	base := s.Base
	if base == "" {
		base = "https://api.cloudflare.com/client/v4"
	}
	return base + "/accounts/" + url.PathEscape(s.Account) + "/r2/buckets/" + url.PathEscape(s.Bucket) + "/objects/" + url.PathEscape(key)
}
func (s *R2) do(ctx context.Context, method, path string, b []byte) ([]byte, int, error) {
	if s.Account == "" || s.Token == "" || s.Bucket == "" {
		return nil, 0, errors.New("R2 account/token/bucket not configured")
	}
	client := s.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	for i := 0; i < 3; i++ {
		req, e := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(b))
		if e != nil {
			return nil, 0, errors.New("R2 request setup failed")
		}
		req.Header.Set("Authorization", "Bearer "+s.Token)
		req.Header.Set("Content-Type", "application/octet-stream")
		if !strings.Contains(path, "/objects/") {
			req.Header.Set("Content-Type", "application/json")
		}
		res, e := client.Do(req)
		if e != nil {
			if i == 2 {
				return nil, 0, errors.New("R2 transport failed")
			}
		} else {
			raw, re := io.ReadAll(io.LimitReader(res.Body, 50<<20))
			res.Body.Close()
			if re == nil && res.StatusCode < 500 && res.StatusCode != 429 {
				return raw, res.StatusCode, nil
			}
			if i == 2 {
				return nil, res.StatusCode, errors.New("R2 bounded retries exhausted")
			}
		}
		timer := time.NewTimer(time.Duration(1<<i) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, 0, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, 0, errors.New("R2 request failed")
}
func (s *R2) Get(ctx context.Context, key string) ([]byte, error) {
	b, code, e := s.do(ctx, "GET", s.endpoint(key), nil)
	if e != nil {
		return nil, e
	}
	if code == 404 {
		return nil, ErrNotFound
	}
	if code != 200 {
		return nil, fmt.Errorf("R2 GET HTTP %d", code)
	}
	return b, nil
}
func (s *R2) Put(ctx context.Context, key string, b []byte, immutable bool) error {
	if immutable {
		old, e := s.Get(ctx, key)
		if e == nil {
			if bytes.Equal(old, b) {
				return nil
			}
			return errors.New("refusing immutable object overwrite")
		}
		if !errors.Is(e, ErrNotFound) {
			return e
		}
	}
	_, code, e := s.do(ctx, "PUT", s.endpoint(key), b)
	if e != nil {
		return e
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("R2 PUT HTTP %d", code)
	}
	return nil
}
func (s *R2) EnsureBucket(ctx context.Context) error {
	endpoint := strings.Split(s.endpoint("x"), "/objects/")[0]
	_, code, e := s.do(ctx, "GET", endpoint, nil)
	if e != nil {
		return e
	}
	if code == 200 {
		return nil
	}
	if code != 404 {
		return fmt.Errorf("R2 bucket check HTTP %d; account may require R2 activation", code)
	}
	// No subscription, payment or public access enablement is performed.
	body, _ := json.Marshal(map[string]string{"name": s.Bucket})
	raw, code, e := s.do(ctx, "POST", strings.TrimSuffix(endpoint, "/"+url.PathEscape(s.Bucket)), body)
	if e != nil {
		return e
	}
	var result struct {
		Success bool `json:"success"`
	}
	_ = json.Unmarshal(raw, &result)
	if code < 200 || code >= 300 || !result.Success {
		return fmt.Errorf("R2 bucket create HTTP %d; account activation/authorization may be required", code)
	}
	return nil
}
