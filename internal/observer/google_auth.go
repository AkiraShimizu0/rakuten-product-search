package observer

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type DesktopClient struct {
	Installed struct {
		ID     string `json:"client_id"`
		Secret string `json:"client_secret"`
		Auth   string `json:"auth_uri"`
		Token  string `json:"token_uri"`
	} `json:"installed"`
}

func ParseDesktopClient(b []byte) (DesktopClient, error) {
	var c DesktopClient
	if json.Unmarshal(b, &c) != nil || c.Installed.ID == "" || c.Installed.Secret == "" || c.Installed.Token != "https://oauth2.googleapis.com/token" || (c.Installed.Auth != "https://accounts.google.com/o/oauth2/auth" && c.Installed.Auth != "https://accounts.google.com/o/oauth2/v2/auth") {
		return c, errors.New("invalid Google Desktop OAuth client; no values logged")
	}
	return c, nil
}
func nonce() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", errors.New("secure random unavailable")
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func AuthURL(id, redirect, state, verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	q := url.Values{"client_id": {id}, "redirect_uri": {redirect}, "response_type": {"code"}, "scope": {Scope}, "state": {state}, "code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"}, "access_type": {"offline"}, "prompt": {"consent"}}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode()
}

type authorizedUser struct {
	Type    string `json:"type"`
	ID      string `json:"client_id"`
	Secret  string `json:"client_secret"`
	Refresh string `json:"refresh_token"`
}
type authorizationSession struct {
	client                                           DesktopClient
	redirect, state, verifier, output, tokenEndpoint string
	http                                             *http.Client
	done                                             chan error
	used                                             atomic.Bool
}

func (s *authorizationSession) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != "GET" || r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(s.state)) != 1 {
		http.Error(w, "Invalid authorization state.", http.StatusForbidden)
		return
	}
	if !s.used.CompareAndSwap(false, true) {
		http.Error(w, "Authorization already handled.", 409)
		return
	}
	finish := func(e error) { s.done <- e }
	if r.URL.Query().Get("error") != "" {
		http.Error(w, "Google authorization was not completed. Return to the assistant.", 400)
		finish(errors.New("AUTHORIZATION_REQUIRED: user consent not completed"))
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code missing.", 400)
		finish(errors.New("authorization code missing"))
		return
	}
	form := url.Values{"client_id": {s.client.Installed.ID}, "client_secret": {s.client.Installed.Secret}, "code": {code}, "code_verifier": {s.verifier}, "redirect_uri": {s.redirect}, "grant_type": {"authorization_code"}}
	req, e := http.NewRequestWithContext(r.Context(), "POST", s.tokenEndpoint, strings.NewReader(form.Encode()))
	if e != nil {
		http.Error(w, "Authorization exchange failed.", 500)
		finish(errors.New("OAuth request setup failed"))
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := s.http.Do(req)
	if e != nil {
		http.Error(w, "Authorization exchange unavailable.", 502)
		finish(errors.New("OAuth exchange unavailable"))
		return
	}
	defer res.Body.Close()
	var token struct {
		Refresh string `json:"refresh_token"`
		Scope   string `json:"scope"`
	}
	if res.StatusCode != 200 || json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&token) != nil || token.Refresh == "" {
		http.Error(w, "Authorization exchange rejected. No credential saved.", 400)
		finish(errors.New("OAuth exchange rejected or no refresh token; upstream body suppressed"))
		return
	}
	if token.Scope != "" && token.Scope != Scope {
		http.Error(w, "Unexpected permission scope. No credential saved.", 400)
		finish(errors.New("OAuth returned unexpected scopes"))
		return
	}
	if e = WriteNew(s.output, authorizedUser{"authorized_user", s.client.Installed.ID, s.client.Installed.Secret, token.Refresh}); e != nil {
		http.Error(w, "Cannot save credential; existing files are not overwritten.", 500)
		finish(e)
		return
	}
	_, _ = io.WriteString(w, "Google read-only authorization saved. You may close this tab and return to the assistant.")
	finish(nil)
}

// AuthorizeGoogle binds only IPv4 loopback, uses PKCE/state, and never automates a browser.
func AuthorizeGoogle(ctx context.Context, clientPath, output string, out io.Writer) error {
	if _, e := os.Stat(output); !os.IsNotExist(e) {
		return errors.New("credential output already exists or unavailable; refusing overwrite")
	}
	b, e := os.ReadFile(clientPath)
	if e != nil {
		return errors.New("downloaded client file unavailable")
	}
	c, e := ParseDesktopClient(b)
	if e != nil {
		return e
	}
	state, e := nonce()
	if e != nil {
		return e
	}
	verifier, e := nonce()
	if e != nil {
		return e
	}
	listener, e := net.Listen("tcp4", "127.0.0.1:0")
	if e != nil {
		return errors.New("loopback listener unavailable")
	}
	defer listener.Close()
	base := "http://" + listener.Addr().String()
	s := &authorizationSession{client: c, redirect: base + "/", state: state, verifier: verifier, output: output, tokenEndpoint: "https://oauth2.googleapis.com/token", http: NewClient().HTTP, done: make(chan error, 1)}
	mux := http.NewServeMux()
	expectedHost := listener.Addr().String()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host, _, e := net.SplitHostPort(r.RemoteAddr)
		if e != nil || host != "127.0.0.1" || r.Host != expectedHost {
			http.Error(w, "Loopback only.", 403)
			return
		}
		if r.URL.Path == "/start" && r.Method == "GET" {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Referrer-Policy", "no-referrer")
			http.Redirect(w, r, AuthURL(c.Installed.ID, s.redirect, state, verifier), http.StatusFound)
			return
		}
		s.callback(w, r)
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	defer func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	fmt.Fprintln(out, "AUTHORIZATION_REQUIRED")
	fmt.Fprintln(out, "Open this local entry in the same PC browser; Google consent requests Search Console read-only:")
	fmt.Fprintln(out, base+"/start")
	select {
	case e := <-s.done:
		if e == nil {
			fmt.Fprintln(out, "Google read-only credential saved; no secret values logged.")
		}
		return e
	case <-ctx.Done():
		return errors.New("authorization timed out; no automatic browser retry")
	case <-serverErrors:
		return errors.New("local authorization listener stopped")
	}
}
