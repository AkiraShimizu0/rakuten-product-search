// Verify the public canary using normal HTTPS certificate verification.
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"jev-money-engine/internal/canary"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Check struct {
	URL       string `json:"url"`
	Status    int    `json:"http_status"`
	TLS       bool   `json:"https_verified"`
	FinalURL  string `json:"final_url"`
	SHA256    string `json:"sha256"`
	Canonical string `json:"canonical,omitempty"`
	Noindex   bool   `json:"noindex"`
	CheckedAt string `json:"checked_at"`
	Error     string `json:"error,omitempty"`
}

func verify(body string, path string, status int, header http.Header) error {
	want := 200
	if path == "/canary-missing-page/" {
		want = 404
	}
	if status != want {
		return fmt.Errorf("HTTP %d expected %d", status, want)
	}
	if path == "/" || path == "/articles/compact-air-purifier-placement/" {
		if err := canary.Check(body, "https://choicelen.page"+path); err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(header.Get("X-Robots-Tag")), "noindex") {
			return fmt.Errorf("noindex HTTP header")
		}
		if !strings.Contains(body, `href="/assets/site.css"`) {
			return fmt.Errorf("CSS link missing")
		}
	}
	if path == "/robots.txt" && !strings.Contains(body, "Allow: /\nSitemap: https://choicelen.page/sitemap.xml") {
		return fmt.Errorf("robots mismatch")
	}
	if path == "/sitemap.xml" && (!strings.Contains(body, "<loc>https://choicelen.page/</loc>") || !strings.Contains(body, "<loc>https://choicelen.page/articles/compact-air-purifier-placement/</loc>")) {
		return fmt.Errorf("sitemap mismatch")
	}
	if path == "/assets/site.css" && !strings.Contains(header.Get("Content-Type"), "text/css") {
		return fmt.Errorf("CSS MIME mismatch")
	}
	return nil
}
func main() {
	out := flag.String("out", "", "new output directory")
	flag.Parse()
	if *out == "" {
		panic("-out required")
	}
	if err := os.MkdirAll(*out, 0755); err != nil {
		panic(err)
	}
	paths := []string{"/", "/articles/compact-air-purifier-placement/", "/assets/site.css", "/robots.txt", "/sitemap.xml", "/canary-missing-page/"}
	client := http.Client{Timeout: 25 * time.Second}
	checks := []Check{}
	failed := false
	for _, path := range paths {
		c := Check{URL: "https://choicelen.page" + path, CheckedAt: time.Now().UTC().Format(time.RFC3339)}
		r, err := client.Get(c.URL)
		if err != nil {
			c.Error = err.Error()
		} else {
			c.Status = r.StatusCode
			c.FinalURL = r.Request.URL.String()
			c.TLS = r.TLS != nil && len(r.TLS.VerifiedChains) > 0
			b, e := io.ReadAll(io.LimitReader(r.Body, 2<<20))
			r.Body.Close()
			c.SHA256 = canary.Hash(b)
			c.Noindex = strings.Contains(strings.ToLower(string(b)), "noindex") || strings.Contains(strings.ToLower(r.Header.Get("X-Robots-Tag")), "noindex")
			if path == "/" || path == paths[1] {
				c.Canonical = c.URL
			}
			if e != nil {
				c.Error = e.Error()
			} else if e = verify(string(b), path, c.Status, r.Header); e != nil {
				c.Error = e.Error()
			}
			if !c.TLS {
				c.Error = "HTTPS certificate not verified"
			}
			if c.FinalURL != c.URL {
				c.Error = "unexpected redirect"
			}
		}
		if c.Error != "" {
			failed = true
		}
		checks = append(checks, c)
		fmt.Printf("%s HTTP%d TLS=%v error=%s\n", path, c.Status, c.TLS, c.Error)
	}
	b, err := json.MarshalIndent(checks, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = canary.WriteNew(filepath.Join(*out, "day6-http-check.json"), append(b, '\n')); err != nil {
		panic(err)
	}
	f, err := os.OpenFile(filepath.Join(*out, "day6-d0-check.csv"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(f)
	w.Write([]string{"checked_at", "url", "http_status", "https_verified", "canonical", "noindex", "error"})
	for _, c := range checks {
		w.Write([]string{c.CheckedAt, c.URL, fmt.Sprint(c.Status), fmt.Sprint(c.TLS), c.Canonical, fmt.Sprint(c.Noindex), c.Error})
	}
	w.Flush()
	err = w.Error()
	f.Close()
	if err != nil {
		panic(err)
	}
	if failed {
		os.Exit(1)
	}
}
