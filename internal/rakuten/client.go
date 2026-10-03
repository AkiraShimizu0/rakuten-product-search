package rakuten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jev-money-engine/internal/product"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const Endpoint = "https://openapi.rakuten.co.jp/ichibams/api/IchibaItem/Search/20260701"

type Config struct {
	AppID, AccessKey, AffiliateID, Origin string
	Endpoint                              string
	Timeout, Interval, Backoff            time.Duration
	MaxRetries                            int
}

type Client struct {
	cfg         Config
	http        *http.Client
	lastRequest time.Time
}

func New(cfg Config) (*Client, error) {
	if cfg.AppID == "" || cfg.AccessKey == "" {
		return nil, errors.New("RAKUTEN_APP_ID and RAKUTEN_ACCESS_KEY are required")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = Endpoint
	}
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, errors.New("invalid API endpoint")
	}
	if cfg.Timeout <= 0 || cfg.Interval < 0 || cfg.Backoff <= 0 || cfg.MaxRetries < 0 || cfg.MaxRetries > 8 {
		return nil, errors.New("invalid timeout, interval, backoff or retries")
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

type Query struct{ GenreID, Keyword, Sort string }
type Page struct {
	Items     []json.RawMessage `json:"items"`
	PageCount int               `json:"pageCount"`
	Count     int               `json:"count"`
}

func pause(ctx context.Context, d time.Duration) error {
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

// SearchPage is used serially: every attempt, including retries, is paced.
func (c *Client) SearchPage(ctx context.Context, q Query, page int) (Page, error) {
	if page < 1 || page > 100 || (q.GenreID == "" && strings.TrimSpace(q.Keyword) == "") {
		return Page{}, errors.New("genre or keyword required; page must be 1..100")
	}
	u, _ := url.Parse(c.cfg.Endpoint)
	params := u.Query()
	params.Set("applicationId", c.cfg.AppID)
	params.Set("format", "json")
	params.Set("formatVersion", "2")
	params.Set("hits", "30")
	params.Set("page", strconv.Itoa(page))
	if q.GenreID != "" {
		params.Set("genreId", q.GenreID)
	}
	if q.Keyword != "" {
		params.Set("keyword", q.Keyword)
	}
	if q.Sort != "" {
		params.Set("sort", q.Sort)
	}
	if c.cfg.AffiliateID != "" {
		params.Set("affiliateId", c.cfg.AffiliateID)
	}
	u.RawQuery = params.Encode()
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if err := pause(ctx, time.Until(c.lastRequest.Add(c.cfg.Interval))); err != nil {
			return Page{}, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return Page{}, errors.New("cannot construct API request")
		}
		req.Header.Set("accessKey", c.cfg.AccessKey)
		req.Header.Set("Accept", "application/json")
		if c.cfg.Origin != "" {
			req.Header.Set("Origin", c.cfg.Origin)
		}
		c.lastRequest = time.Now()
		resp, err := c.http.Do(req)
		var retryDelay time.Duration
		if err != nil {
			if ctx.Err() != nil {
				return Page{}, ctx.Err()
			}
			// url.Error includes the app ID in its URL. Do not log it.
			if attempt == c.cfg.MaxRetries {
				return Page{}, errors.New("API network/timeout failure; retries exhausted")
			}
		} else {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024+1))
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if readErr == nil {
					if len(body) > 16*1024*1024 {
						return Page{}, errors.New("API response exceeds 16 MiB")
					}
					var result Page
					if err := json.Unmarshal(body, &result); err != nil {
						return Page{}, errors.New("invalid API JSON response")
					}
					if result.Items == nil {
						return Page{}, errors.New("API response missing items array")
					}
					return result, nil
				}
				if attempt == c.cfg.MaxRetries {
					return Page{}, errors.New("API response read failed; retries exhausted")
				}
			} else {
				retryable := resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode <= 599)
				if !retryable || attempt == c.cfg.MaxRetries {
					return Page{}, fmt.Errorf("Rakuten API HTTP %d%s (page %d, attempts %d); check credentials, allowed IP, origin and query", resp.StatusCode, errorHint(body), page, attempt+1)
				}
				retryDelay = retryAfter(resp.Header.Get("Retry-After"), time.Now())
				if retryDelay > 2*time.Minute {
					return Page{}, errors.New("API Retry-After exceeds two minutes; try again later")
				}
			}
		}
		delay := c.cfg.Backoff * time.Duration(1<<attempt)
		if delay > 2*time.Minute {
			delay = 2 * time.Minute
		}
		if retryDelay > delay {
			delay = retryDelay
		}
		if err := pause(ctx, delay); err != nil {
			return Page{}, err
		}
	}
	return Page{}, errors.New("API retries exhausted")
}

// Only fixed known messages are exposed, never arbitrary response bodies.
func errorHint(body []byte) string {
	var v struct {
		Errors struct {
			Message string `json:"errorMessage"`
		} `json:"errors"`
	}
	if json.Unmarshal(body, &v) != nil {
		return ""
	}
	switch v.Errors.Message {
	case "CLIENT_IP_NOT_ALLOWED":
		return " CLIENT_IP_NOT_ALLOWED: add the current outbound IP to the Rakuten app allowed IPs"
	case "INVALID_ACCESS_KEY":
		return " INVALID_ACCESS_KEY"
	default:
		return ""
	}
}

func retryAfter(s string, now time.Time) time.Duration {
	if seconds, err := strconv.Atoi(s); err == nil && seconds > 0 {
		if seconds > 120 {
			return 3 * time.Minute
		}
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(s); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
}

// Normalize keeps the original individual item JSON, including unknown fields.
// Missing optional data is recorded; missing identity skips only this item.
func Normalize(raw json.RawMessage, now time.Time) (product.Product, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return product.Product{}, errors.New("invalid item JSON")
	}
	for _, key := range []string{"Item", "item"} {
		if nested, ok := fields[key]; ok {
			raw = nested
			if err := json.Unmarshal(raw, &fields); err != nil {
				return product.Product{}, errors.New("invalid wrapped item")
			}
			break
		}
	}
	p := product.Product{Source: "rakuten", FirstSeenAt: now.UTC(), LastSeenAt: now.UTC(), RawJSON: append(json.RawMessage(nil), raw...)}
	text := func(key string) string {
		var value string
		if err := json.Unmarshal(fields[key], &value); err != nil {
			return ""
		}
		return strings.TrimSpace(value)
	}
	number := func(key string, integer bool) float64 {
		b, ok := fields[key]
		var value float64
		if ok && string(b) != "null" {
			s := strings.Trim(string(b), "\"")
			v, err := strconv.ParseFloat(s, 64)
			if err == nil && !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && (!integer || (v < math.Exp2(63) && v == math.Trunc(v))) {
				value = v
				return value
			}
		}
		p.MissingFields = append(p.MissingFields, key)
		return 0
	}
	p.SourceID = text("itemCode")
	if p.SourceID == "" {
		return product.Product{}, errors.New("item missing itemCode")
	}
	p.Name = product.CleanText(text("itemName"))
	p.Caption = product.CleanText(text("itemCaption"))
	p.Price = int64(number("itemPrice", true))
	p.ReviewCount = int64(number("reviewCount", true))
	p.ReviewAverage = number("reviewAverage", false)
	if p.ReviewAverage > 5 {
		p.MissingFields = append(p.MissingFields, "reviewAverage")
		p.ReviewAverage = 0
	}
	p.GenreID = text("genreId")
	if p.GenreID == "" && len(fields["genreId"]) > 0 && string(fields["genreId"]) != "null" {
		var n json.Number
		if json.Unmarshal(fields["genreId"], &n) == nil {
			p.GenreID = n.String()
		}
	}
	p.ShopName = product.CleanText(text("shopName"))
	p.ItemURL = text("itemUrl")
	p.AffiliateURL = text("affiliateUrl")
	for _, f := range []struct{ key, value string }{{"itemName", p.Name}, {"itemCaption", p.Caption}, {"genreId", p.GenreID}, {"shopName", p.ShopName}, {"itemUrl", p.ItemURL}} {
		if f.value == "" {
			p.MissingFields = append(p.MissingFields, f.key)
		}
	}
	return p, nil
}
