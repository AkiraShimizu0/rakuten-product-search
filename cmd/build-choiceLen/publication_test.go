package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommittedReleaseArtifact(t *testing.T) {
	dir := "../../site"
	if _, e := os.Stat(filepath.Join(dir, "build-manifest.json")); os.IsNotExist(e) {
		t.Skip("release artifact not staged on this branch yet")
	}
	first, e := os.ReadFile("../../content/published/compact-air-purifier-placement.md")
	if e != nil {
		t.Fatal(e)
	}
	second, e := os.ReadFile("../../content/published/winter-laundry-dehumidifier-conditions.md")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = validateReleaseArtifact(dir, string(first), string(second)); e != nil {
		t.Fatal(e)
	}
	tmp := t.TempDir()
	e = filepath.WalkDir(dir, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		dest := filepath.Join(tmp, rel)
		os.MkdirAll(filepath.Dir(dest), 0700)
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		return os.WriteFile(dest, b, 0600)
	})
	if e != nil {
		t.Fatal(e)
	}
	page := filepath.Join(tmp, "index.html")
	b, _ := os.ReadFile(page)
	os.WriteFile(page, append(b, []byte("tampered")...), 0600)
	if _, e = validateReleaseArtifact(tmp, string(first), string(second)); e == nil {
		t.Fatal("tampered artifact accepted")
	}
	if _, e = validateReleaseArtifact(dir, string(first), string(second)+"tampered"); e == nil {
		t.Fatal("modified audited draft accepted")
	}
}
func TestReleaseHeadMismatch(t *testing.T) {
	if releaseHEAD(strings.Repeat("0", 40)) == nil {
		t.Fatal("unrelated commit accepted")
	}
}
