package main

import (
	"flag"
	"fmt"
	"jev-money-engine/internal/contentvalidation"
	"os"
	"path/filepath"
)

func main() {
	mode := flag.String("mode", "prepare", "prepare/analyze")
	data := flag.String("data", "../../outputs/day5-content-validation/data", "data folder")
	content := flag.String("content", "", "prototype folder")
	flag.Parse()
	if *content == "" {
		*content = filepath.Join(filepath.Dir(*data), "content")
	}
	var e error
	switch *mode {
	case "prepare":
		e = contentvalidation.Prepare(*data, *content)
	case "analyze":
		e = contentvalidation.Analyze(*data)
	default:
		e = fmt.Errorf("unknown mode")
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
