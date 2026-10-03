package rakuten

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

func TestNormalize(t *testing.T) {
	raw := json.RawMessage(`{"itemCode":"shop:123","itemName":" 商品 &amp; A ","itemCaption":"<p>日本語</p><br>説明","itemPrice":3500,"reviewCount":"8","reviewAverage":"4.5","genreId":123,"shopName":" 店 ","itemUrl":"https://example.com/item","affiliateUrl":"https://example.com/aff","unknown":true}`)
	now := time.Date(2026, 10, 4, 1, 2, 3, 0, time.UTC)
	p, err := Normalize(raw, now)
	if err != nil {
		t.Fatal(err)
	}
	if p.Source != "rakuten" || p.SourceID != "shop:123" || p.Name != "商品 & A" || p.Caption != "日本語 説明" || p.Price != 3500 || p.ReviewCount != 8 || p.ReviewAverage != 4.5 || p.GenreID != "123" || p.ShopName != "店" || p.ItemURL != "https://example.com/item" || p.AffiliateURL != "https://example.com/aff" || !p.FirstSeenAt.Equal(now) || !p.LastSeenAt.Equal(now) || string(p.RawJSON) != string(raw) || len(p.MissingFields) != 0 {
		t.Fatalf("unexpected product: %+v", p)
	}
	wrapped, err := Normalize(json.RawMessage(`{"Item":`+string(raw)+`}`), now)
	if err != nil || wrapped.SourceID != p.SourceID {
		t.Fatalf("wrapper: %v", err)
	}
	missing, err := Normalize(json.RawMessage(`{"itemCode":"x","itemPrice":null,"reviewCount":"bad","reviewAverage":8}`), now)
	if err != nil || missing.Price != 0 || len(missing.MissingFields) < 3 {
		t.Fatalf("missing fields: %+v %v", missing, err)
	}
	for _, raw := range []string{`{}`, `null`, `{bad`, `{"itemCode":""}`} {
		if _, err := Normalize(json.RawMessage(raw), now); err == nil {
			t.Errorf("expected identity error for %s", raw)
		}
	}
}

func testClient(t *testing.T, handler http.HandlerFunc, retries int) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c, err := New(Config{AppID: "secret-app", AccessKey: "secret-key", AffiliateID: "aff", Endpoint: server.URL, Timeout: time.Second, Backoff: time.Millisecond, MaxRetries: retries})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRetryAndParameters(t *testing.T) {
	calls := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("accessKey") != "secret-key" || r.URL.Query().Get("accessKey") != "" || r.URL.Query().Get("applicationId") != "secret-app" || r.URL.Query().Get("affiliateId") != "aff" || r.URL.Query().Get("keyword") != "空気 清浄" || r.URL.Query().Get("formatVersion") != "2" || r.URL.Query().Get("page") != "2" || r.URL.Query().Get("hits") != "30" {
			t.Error("incorrect authentication/parameters")
		}
		if calls == 1 {
			w.WriteHeader(429)
			return
		}
		if calls == 2 {
			w.WriteHeader(503)
			return
		}
		fmt.Fprint(w, `{"Items":[],"pageCount":3}`)
	}, 2)
	page, err := c.SearchPage(context.Background(), Query{Keyword: "空気 清浄"}, 2)
	if err != nil || calls != 3 || page.PageCount != 3 {
		t.Fatalf("retry result: calls=%d err=%v page=%+v", calls, err, page)
	}
}

func TestRetryBoundAndNoCredentialLeaks(t *testing.T) {
	for _, status := range []int{400, 401, 403, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(status)
				fmt.Fprint(w, "secret-app secret-key")
			}, 2)
			_, err := c.SearchPage(context.Background(), Query{GenreID: "123"}, 1)
			expected := 1
			if status == 429 || status == 500 {
				expected = 3
			}
			if err == nil || calls != expected || strings.Contains(err.Error(), "secret-") {
				t.Fatalf("calls=%d error=%v", calls, err)
			}
		})
	}
}

func TestTimeoutCancellationAndMalformedResponse(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }, 0)
	c.http.Timeout = 10 * time.Millisecond
	if _, err := c.SearchPage(context.Background(), Query{GenreID: "1"}, 1); err == nil || strings.Contains(err.Error(), "secret-app") {
		t.Fatalf("timeout error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.SearchPage(ctx, Query{GenreID: "1"}, 1); err != context.Canceled {
		t.Fatalf("cancel: %v", err)
	}
	for _, body := range []string{`{bad`, `{"error":"bad"}`} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }, 0)
		if _, err := c.SearchPage(context.Background(), Query{GenreID: "1"}, 1); err == nil {
			t.Errorf("accepted %s", body)
		}
	}
}

func TestRetryAfter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	if retryAfter("3", now) != 3*time.Second || retryAfter(now.Add(5*time.Second).Format(http.TimeFormat), now) != 5*time.Second || retryAfter("bad", now) != 0 {
		t.Fatal("Retry-After parsing")
	}
}
