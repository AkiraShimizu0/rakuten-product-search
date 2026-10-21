package main

import (
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/canary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// verifyRelease verifies the artifact being deployed; it never regenerates it.
func verifyRelease(dir, first string) error {
	second, e := os.ReadFile("content/published/winter-laundry-dehumidifier-conditions.md")
	if e != nil {
		return e
	}
	commit, e := validateReleaseArtifact(dir, first, string(second))
	if e != nil {
		return e
	}
	if e = exec.Command("git", "merge-base", "--is-ancestor", commit, "HEAD").Run(); e != nil {
		return fmt.Errorf("release source commit must be an ancestor of artifact checkout")
	}
	// A descendant artifact/docs commit cannot silently change the builder/content/config.
	if e = exec.Command("git", "diff", "--exit-code", commit, "HEAD", "--", "cmd", "internal", "content/published", "wrangler.jsonc").Run(); e != nil {
		return fmt.Errorf("builder/content/config diverged from release source commit")
	}
	return nil
}

func validateReleaseArtifact(dir, first, second string) (string, error) {
	var m struct {
		Version string            `json:"builder_version"`
		Commit  string            `json:"builder_commit"`
		Stage   string            `json:"stage"`
		Claims  string            `json:"claims_sha256"`
		Hashes  map[string]string `json:"output_sha256"`
		Release *releaseMetadata  `json:"release"`
	}
	raw, e := os.ReadFile(filepath.Join(dir, "build-manifest.json"))
	if e != nil {
		return "", e
	}
	if json.Unmarshal(raw, &m) != nil || m.Version != builderVersion || m.Stage != "release" || m.Claims != secondClaimsHash || canary.Hash([]byte(second)) != secondHash {
		return "", fmt.Errorf("release provenance mismatch")
	}
	o := integrationOptions{Stage: "release", BuilderCommit: m.Commit, Release: m.Release}
	if e = validateOptions(o); e != nil {
		return "", e
	}
	names := []string{"index.html", "articles/compact-air-purifier-placement/index.html", strings.TrimPrefix(secondPath, "/") + "index.html", "assets/site.css", "assets/dehumidifier.css", "404.html", "robots.txt", "sitemap.xml"}
	if len(m.Hashes) != len(names) {
		return "", fmt.Errorf("unexpected manifest file count")
	}
	files := map[string]string{}
	for _, n := range names {
		b, e := os.ReadFile(filepath.Join(dir, filepath.FromSlash(n)))
		if e != nil {
			return "", e
		}
		text := strings.ReplaceAll(string(b), "\r\n", "\n")
		if canary.Hash([]byte(text)) != m.Hashes[n] {
			return "", fmt.Errorf("release file hash mismatch: %s", n)
		}
		files[n] = text
	}
	base, e := build(first, "2026-10-06")
	if e != nil {
		return "", e
	}
	if e = validateTwo(files, base, second, o); e != nil {
		return "", e
	}
	h := files[strings.TrimPrefix(secondPath, "/")+"index.html"]
	if !strings.Contains(h, `<time datetime="`+m.Release.PublishedAt+`">`) {
		return "", fmt.Errorf("publication date missing from article")
	}
	return m.Commit, nil
}
