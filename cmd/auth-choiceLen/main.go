package main

import (
	"context"
	"flag"
	"fmt"
	"jev-money-engine/internal/observer"
	"os"
	"time"
)

func main() {
	client := flag.String("client", "", "private downloaded Desktop OAuth client JSON")
	out := flag.String("out", "../observer-google-adc.json", "new private authorized_user ADC file outside repository")
	flag.Parse()
	if *client == "" {
		fmt.Fprintln(os.Stderr, "-client is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if e := observer.AuthorizeGoogle(ctx, *client, *out, os.Stdout); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
