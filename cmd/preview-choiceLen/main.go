package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"regexp"
	"strings"
	"time"
)

// Preview-only transformations never alter release-candidate files.
func previewHTML(b []byte) []byte {
	scripts := regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script\s*>`)
	return scripts.ReplaceAllFunc(b, func(s []byte) []byte {
		opening := s[:bytes.IndexByte(s, '>')+1]
		if regexp.MustCompile(`(?i)\btype\s*=\s*["']application/(?:json|ld\+json)["']`).Match(opening) {
			return s
		}
		return nil
	})
}
func previewHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "script-src 'none'; connect-src 'none'; object-src 'none'")
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		f, e := http.Dir(dir).Open(r.URL.Path)
		if e == nil {
			defer f.Close()
			info, e := f.Stat()
			if e == nil && info.IsDir() {
				if !strings.HasSuffix(r.URL.Path, "/") {
					http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
					return
				}
				f.Close()
				f, e = http.Dir(dir).Open(r.URL.Path + "index.html")
				if e == nil {
					defer f.Close()
				}
			}
			if e == nil && (strings.HasSuffix(r.URL.Path, "/") || strings.HasSuffix(r.URL.Path, ".html")) {
				b, e := io.ReadAll(f)
				if e == nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					if r.Method == "GET" {
						w.Write(previewHTML(b))
					}
					return
				}
			}
		}
		files.ServeHTTP(w, r)
	})
}
func main() {
	dir := flag.String("dir", "site", "site directory")
	port := flag.Int("port", 8787, "loopback preview port")
	flag.Parse()
	if *port < 1024 || *port > 65535 {
		log.Fatal("port must be 1024..65535")
	}
	s := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", *port), Handler: previewHandler(*dir), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}
