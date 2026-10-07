package radarhistory

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHub is a private append-object repository. Scheduled jobs use only their
// own repository's ephemeral GITHUB_TOKEN; no cross-repository PAT is required.
type GitHub struct {
	Repo, Token, Base string
	HTTP              *http.Client
}

func (s *GitHub) request(ctx context.Context, method, key string, body []byte, raw bool) ([]byte, int, error) {
	base := s.Base
	if base == "" {
		base = "https://api.github.com"
	}
	if len(strings.Split(s.Repo, "/")) != 2 || s.Token == "" {
		return nil, 0, errors.New("private history repository/token required")
	}
	path := base + "/repos/" + s.Repo
	if key != "" {
		path += "/contents/" + url.PathEscape(key)
	}
	client := s.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, e := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(body))
		if e != nil {
			return nil, 0, errors.New("history request setup failed")
		}
		req.Header.Set("Authorization", "Bearer "+s.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		req.Header.Set("Accept", "application/vnd.github+json")
		if raw {
			req.Header.Set("Accept", "application/vnd.github.raw+json")
		}
		res, e := client.Do(req)
		if e == nil {
			b, re := io.ReadAll(io.LimitReader(res.Body, 50<<20))
			res.Body.Close()
			if re == nil && res.StatusCode < 500 && res.StatusCode != 429 && !(method == "PUT" && res.StatusCode == 409) {
				return b, res.StatusCode, nil
			}
		}
		if attempt == 2 {
			return nil, 0, errors.New("history bounded retries exhausted")
		}
		timer := time.NewTimer(time.Duration(1<<attempt) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, 0, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, 0, errors.New("history request failed")
}
func (s *GitHub) CheckPrivate(ctx context.Context) error {
	b, code, e := s.request(ctx, "GET", "", nil, false)
	if e != nil {
		return e
	}
	var v struct{ Private bool }
	if code != 200 || json.Unmarshal(b, &v) != nil || !v.Private {
		return fmt.Errorf("private history repository verification failed HTTP %d", code)
	}
	return nil
}
func (s *GitHub) Get(ctx context.Context, key string) ([]byte, error) {
	b, code, e := s.request(ctx, "GET", key, nil, true)
	if e != nil {
		return nil, e
	}
	if code == 404 {
		return nil, ErrNotFound
	}
	if code != 200 {
		return nil, fmt.Errorf("private history read HTTP %d", code)
	}
	return b, nil
}
func (s *GitHub) Put(ctx context.Context, key string, b []byte, immutable bool) error {
	sha := ""
	old, code, e := s.request(ctx, "GET", key, nil, false)
	if e != nil {
		return e
	}
	if code != 200 && code != 404 {
		return fmt.Errorf("private history metadata HTTP %d", code)
	}
	if code == 200 {
		var meta struct{ SHA string }
		if e = json.Unmarshal(old, &meta); e != nil || meta.SHA == "" {
			return errors.New("invalid private history metadata")
		}
		sha = meta.SHA
		original, e := s.Get(ctx, key)
		if e != nil {
			return e
		}
		if bytes.Equal(original, b) {
			return nil
		}
		if immutable {
			return errors.New("refusing immutable history overwrite")
		}
	}
	payload := map[string]string{"message": "Record private radar observation object", "content": base64.StdEncoding.EncodeToString(b)}
	if sha != "" {
		payload["sha"] = sha
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return e
	}
	_, code, e = s.request(ctx, "PUT", key, body, false)
	if e != nil {
		return e
	}
	if code != 200 && code != 201 {
		return fmt.Errorf("private history write HTTP %d", code)
	}
	return nil
}

func (s *GitHub) PutCAS(ctx context.Context, key string, b, expected []byte) error {
	meta, code, e := s.request(ctx, "GET", key, nil, false)
	if e != nil {
		return e
	}
	sha := ""
	if code == 404 {
		if expected != nil {
			return errors.New("history pointer disappeared")
		}
	} else if code == 200 {
		var v struct{ SHA string }
		if json.Unmarshal(meta, &v) != nil || v.SHA == "" {
			return errors.New("invalid CAS metadata")
		}
		sha = v.SHA
		old, e := s.Get(ctx, key)
		if e != nil {
			return e
		}
		if !bytes.Equal(old, expected) {
			return errors.New("history pointer changed; refusing stale update")
		}
	} else {
		return fmt.Errorf("history CAS metadata HTTP %d", code)
	}
	body, _ := json.Marshal(map[string]string{"message": "Advance private radar cache pointer", "content": base64.StdEncoding.EncodeToString(b), "sha": sha})
	_, code, e = s.request(ctx, "PUT", key, body, false)
	if e != nil {
		return e
	}
	if code != 200 && code != 201 {
		return fmt.Errorf("history pointer CAS HTTP %d", code)
	}
	return nil
}
