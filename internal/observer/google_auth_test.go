package observer

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDesktopAuthReadOnlyPKCE(t *testing.T) {
	n, e := nonce()
	if e != nil || len(n) < 43 {
		t.Fatal("invalid PKCE verifier")
	}
	u, e := url.Parse(AuthURL("fake-client", "http://127.0.0.1:1234/", "fake-state", n))
	if e != nil {
		t.Fatal(e)
	}
	q := u.Query()
	if u.Host != "accounts.google.com" || q.Get("scope") != Scope || q.Get("code_challenge_method") != "S256" || q.Get("state") != "fake-state" || q.Get("code_challenge") == n {
		t.Fatal("invalid read-only authorization URL")
	}
	if _, e = ParseDesktopClient([]byte(`{"web":{"client_id":"fake"}}`)); e == nil {
		t.Fatal("non desktop client accepted")
	}
}
func TestAuthCallbackStateAndSafeSave(t *testing.T) {
	file := filepath.Join(t.TempDir(), "adc.json")
	exchange := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code_verifier") != "fake-verifier" || r.Form.Get("grant_type") != "authorization_code" {
			t.Error("missing PKCE exchange")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"refresh_token": "fake-refresh", "scope": Scope})
	}))
	defer exchange.Close()
	s := &authorizationSession{state: "expected", verifier: "fake-verifier", redirect: "http://127.0.0.1:1234/", output: file, tokenEndpoint: exchange.URL, http: exchange.Client(), done: make(chan error, 1)}
	bad := httptest.NewRecorder()
	s.callback(bad, httptest.NewRequest("GET", "http://127.0.0.1/?state=wrong&code=fake", nil))
	if bad.Code != 403 || s.used.Load() {
		t.Fatal("invalid state consumed flow")
	}
	good := httptest.NewRecorder()
	s.callback(good, httptest.NewRequest("GET", "http://127.0.0.1/?state=expected&code=fake", nil))
	if e := <-s.done; e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(file)
	if e != nil || !strings.Contains(string(b), "authorized_user") {
		t.Fatal("ADC not saved")
	}
	if strings.Contains(good.Body.String(), "fake-refresh") {
		t.Fatal("refresh token in browser response")
	}
	repeat := httptest.NewRecorder()
	s.callback(repeat, httptest.NewRequest("GET", "http://127.0.0.1/?state=expected&code=fake", nil))
	if repeat.Code != 409 {
		t.Fatal("callback replay accepted")
	}
}
func TestAuthExchangeErrorRedaction(t *testing.T) {
	exchange := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = io.WriteString(w, "do-not-log-upstream")
	}))
	defer exchange.Close()
	file := filepath.Join(t.TempDir(), "adc.json")
	s := &authorizationSession{state: "expected", output: file, tokenEndpoint: exchange.URL, http: exchange.Client(), done: make(chan error, 1)}
	rec := httptest.NewRecorder()
	s.callback(rec, httptest.NewRequest("GET", "http://127.0.0.1/?state=expected&code=fake", nil))
	e := <-s.done
	if e == nil || strings.Contains(e.Error(), "do-not-log") || strings.Contains(rec.Body.String(), "do-not-log") {
		t.Fatal("upstream response leaked")
	}
	if _, e = os.Stat(file); !os.IsNotExist(e) {
		t.Fatal("failed exchange saved credential")
	}
}
func TestAuthUnexpectedScope(t *testing.T) {
	exchange := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"refresh_token":"fake","scope":"https://www.googleapis.com/auth/webmasters"}`)
	}))
	defer exchange.Close()
	s := &authorizationSession{state: "expected", output: filepath.Join(t.TempDir(), "adc.json"), tokenEndpoint: exchange.URL, http: exchange.Client(), done: make(chan error, 1)}
	s.callback(httptest.NewRecorder(), httptest.NewRequest("GET", "http://127.0.0.1/?state=expected&code=fake", nil))
	if e := <-s.done; e == nil {
		t.Fatal("write scope accepted")
	}
}
