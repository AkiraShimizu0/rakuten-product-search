package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewSafety(t *testing.T) {
	dir := t.TempDir()
	original := `<html><script src="https://static.cloudflareinsights.com/beacon.min.js"></script><script>fetch('https://example.org')</script><script type="application/json" id="claim-source-map">{"claims":["DD01"]}</script><body>article</body></html>`
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(original), 0600)
	h := previewHandler(dir)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 || strings.Contains(w.Body.String(), "beacon.min.js") || strings.Contains(w.Body.String(), "fetch(") || !strings.Contains(w.Body.String(), "claim-source-map") {
		t.Fatal("preview script safety")
	}
	if !strings.Contains(w.Header().Get("X-Robots-Tag"), "noindex") || !strings.Contains(w.Header().Get("Content-Security-Policy"), "connect-src 'none'") {
		t.Fatal("preview headers")
	}
	b, _ := os.ReadFile(filepath.Join(dir, "index.html"))
	if string(b) != original {
		t.Fatal("candidate altered")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/missing.css", nil))
	if w.Code != 404 {
		t.Fatal("missing file")
	}
}
