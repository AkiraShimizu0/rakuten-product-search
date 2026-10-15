package observer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	_ "time/tzdata"
)

const Scope = "https://www.googleapis.com/auth/webmasters.readonly"
const Article = "https://choicelen.page/articles/compact-air-purifier-placement/"

type Window struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Timezone    string `json:"timezone"`
	Granularity string `json:"granularity"`
}
type Metric struct {
	Value  any    `json:"value"`
	Status string `json:"status"`
}

func Known(v float64) Metric {
	s := "known"
	if v == 0 {
		s = "zero"
	}
	return Metric{v, s}
}
func Unknown() Metric { return Metric{nil, "unknown"} }

type Source struct {
	Status          string            `json:"status"`
	Reason          string            `json:"reason,omitempty"`
	Endpoint        string            `json:"source"`
	Method          string            `json:"method"`
	Window          Window            `json:"actual_window"`
	CheckedAt       string            `json:"checked_at"`
	LatestAvailable string            `json:"latest_available_date"`
	Delay           string            `json:"reporting_delay"`
	Metrics         map[string]Metric `json:"metrics,omitempty"`
	Data            any               `json:"data,omitempty"`
	Notes           []string          `json:"notes,omitempty"`
}
type Observation struct {
	Protocol  string            `json:"protocol_version"`
	Requested Window            `json:"requested_window"`
	Sources   map[string]Source `json:"sources"`
	Decisions []string          `json:"pre_registered_decisions"`
	Pending   []string          `json:"pending"`
}

func NewSource(endpoint string, w Window) Source {
	return Source{Status: "unknown", Endpoint: endpoint, Method: "read-only API", Window: w, CheckedAt: time.Now().UTC().Format(time.RFC3339), LatestAvailable: "unknown", Delay: "unknown"}
}

type Client struct {
	HTTP    *http.Client
	Retries int
	Backoff time.Duration
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, Retries: 2, Backoff: time.Second}
}

// Do emits neither upstream bodies nor transport errors: they may contain secrets.
func (c *Client) Do(ctx context.Context, method, endpoint, token string, payload any) ([]byte, error) {
	var body []byte
	if payload != nil {
		var e error
		body, e = json.Marshal(payload)
		if e != nil {
			return nil, errors.New("request encoding failed")
		}
	}
	for attempt := 0; attempt <= c.Retries; attempt++ {
		req, e := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
		if e != nil {
			return nil, errors.New("request setup failed")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Content-Type", "application/json")
		res, e := c.HTTP.Do(req)
		if e == nil {
			b, re := io.ReadAll(io.LimitReader(res.Body, 16<<20))
			res.Body.Close()
			if re == nil && res.StatusCode >= 200 && res.StatusCode < 300 {
				return b, nil
			}
			if res.StatusCode != 429 && res.StatusCode < 500 {
				return nil, fmt.Errorf("read-only API HTTP %d", res.StatusCode)
			}
		}
		if attempt < c.Retries {
			select {
			case <-ctx.Done():
				return nil, errors.New("request cancelled")
			case <-time.After(c.Backoff * time.Duration(1<<attempt)):
			}
		}
	}
	return nil, errors.New("read-only API bounded retries exhausted")
}

// GoogleToken supports explicit process tokens and authorized_user ADC, not browser sessions.
func GoogleToken(ctx context.Context, c *Client) (string, error) {
	if v := os.Getenv("CHOICELEN_GOOGLE_ACCESS_TOKEN"); v != "" {
		return v, nil
	}
	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if path == "" {
		path = filepath.Join(os.Getenv("APPDATA"), "gcloud", "application_default_credentials.json")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return "", errors.New("AUTHORIZATION_REQUIRED: Google read-only ADC missing")
	}
	var cred struct {
		Type         string `json:"type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Refresh      string `json:"refresh_token"`
	}
	if json.Unmarshal(b, &cred) != nil || cred.Type != "authorized_user" || cred.ClientID == "" || cred.ClientSecret == "" || cred.Refresh == "" {
		return "", errors.New("AUTHORIZATION_REQUIRED: authorized_user ADC required (other types unsupported)")
	}
	values := url.Values{"client_id": {cred.ClientID}, "client_secret": {cred.ClientSecret}, "refresh_token": {cred.Refresh}, "grant_type": {"refresh_token"}, "scope": {Scope}}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	if e != nil {
		return "", errors.New("OAuth setup failed")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := c.HTTP.Do(req)
	if e != nil {
		return "", errors.New("OAuth refresh unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", errors.New("AUTHORIZATION_REQUIRED: OAuth refresh rejected")
	}
	var out struct {
		Access string `json:"access_token"`
	}
	if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&out) != nil || out.Access == "" {
		return "", errors.New("OAuth refresh invalid response")
	}
	return out.Access, nil
}
func WriteNew(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errors.New("output exists or cannot be created")
	}
	_, e = f.Write(append(b, '\n'))
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}

// Inaccessible credentials do not prove a broken production measurement pipeline.
func Decide(o Observation) ([]string, []string) {
	decisions := []string{"NO SEO PROFITABILITY DECISION"}
	pending := []string{}
	problem := false
	if s, ok := o.Sources["url_inspection"]; ok && s.Status == "known" {
		if v, ok := s.Data.(Inspection); ok {
			x := v.Index
			if x.Verdict == "FAIL" || x.Verdict == "NEUTRAL" || x.Verdict == "PARTIAL" || x.Robots == "DISALLOWED" || x.Indexing == "BLOCKED_BY_META_TAG" || x.Indexing == "BLOCKED_BY_HTTP_HEADER" || (x.GoogleCanonical != "" && x.GoogleCanonical != Article) {
				problem = true
			}
		}
	} else {
		pending = append(pending, "current URL inspection unavailable")
	}
	if s, ok := o.Sources["sitemap"]; ok && s.Status == "known" {
		if v, ok := s.Data.(Sitemap); ok && v.Errors != nil && *v.Errors > 0 {
			problem = true
		}
	} else {
		pending = append(pending, "sitemap unavailable")
	}
	operational := true
	for _, name := range []string{"search_console", "url_inspection", "sitemap", "cloudflare", "rakuten"} {
		if o.Sources[name].Status != "known" {
			operational = false
			pending = append(pending, name+" incomplete")
		}
	}
	if problem {
		decisions = append(decisions, "INVESTIGATE_INDEXING")
	}
	if v, ok := o.Sources["url_inspection"].Data.(Inspection); !ok || v.Index.Verdict == "" || v.Index.Verdict == "VERDICT_UNSPECIFIED" {
		operational = false
	}
	if v, ok := o.Sources["sitemap"].Data.(Sitemap); !ok || v.Errors == nil {
		operational = false
	}
	for _, pair := range [][2]string{{"search_console", "impressions"}, {"search_console", "clicks"}, {"cloudflare", "page_views"}, {"cloudflare", "visits"}, {"rakuten", "clicks"}, {"rakuten", "orders"}, {"rakuten", "commission_jpy"}} {
		m := o.Sources[pair[0]].Metrics[pair[1]]
		if m.Status != "known" && m.Status != "zero" {
			operational = false
		}
	}
	health, ok := o.Sources["production_health"]
	if operational && ok && health.Status == "known" {
		decisions = append(decisions, "CONTINUE")
	} else {
		pending = append(pending, "CONTINUE requires independently verified production health and operational measurement")
	}
	return decisions, pending
}
