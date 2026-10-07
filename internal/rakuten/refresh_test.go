package rakuten

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestItemRefreshQuery(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("itemCode") != "shop:item" || r.URL.Query().Get("availability") != "0" {
			t.Error(r.URL.Query())
		}
		w.Write([]byte(`{"items":[],"pageCount":0}`))
	}))
	defer s.Close()
	c, e := New(Config{AppID: "app", AccessKey: "key", Endpoint: s.URL, Timeout: time.Second, Backoff: time.Millisecond})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.SearchPage(context.Background(), Query{ItemCode: "shop:item", AllOffers: true}, 1); e != nil {
		t.Fatal(e)
	}
}
func TestExplicitAvailability(t *testing.T) {
	for raw, want := range map[string]string{`{"availability":1}`: "api_available", `{"availability":0}`: "api_unavailable", `{}`: "unknown", `{"availability":null}`: "unknown"} {
		if got := Availability(json.RawMessage(raw)); got != want {
			t.Fatal(got, want)
		}
	}
}
