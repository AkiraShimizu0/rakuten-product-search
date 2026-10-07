package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	dir := flag.String("dir", "site", "site directory")
	flag.Parse()
	s := &http.Server{Addr: "127.0.0.1:8787", Handler: http.FileServer(http.Dir(*dir)), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}
