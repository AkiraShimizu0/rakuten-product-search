package canary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPage(t *testing.T) {
	md := "# Article\n" + Disclosure + "\n## Table\n|a|b|\n|---|---|\n|c|d|\n[link](https://hb.afl.rakuten.co.jp/hgc/test)"
	s, e := Render(md, "https://example.org/a/")
	if e != nil {
		t.Fatal(e)
	}
	if e = Check(s, "https://example.org/a/"); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(s, `rel="sponsored"`) {
		t.Fatal("missing sponsored")
	}
	s2, _ := Render(md, "https://example.org/a/")
	if s != s2 {
		t.Fatal("not deterministic")
	}
}
func TestInvalid(t *testing.T) {
	for _, md := range []string{"# a\nno disclosure", "# a\n# b\n" + Disclosure, "# a\n" + Disclosure + "\n絶対"} {
		if _, e := Render(md, ""); e == nil {
			t.Fatal("invalid accepted")
		}
	}
	if _, e := Render("# a\n"+Disclosure, "http://example.org"); e == nil {
		t.Fatal("non HTTPS")
	}
}
func TestChecks(t *testing.T) {
	s, _ := Render("# a\n"+Disclosure, "https://example.org/a")
	for _, bad := range []string{strings.Replace(s, Disclosure, "", 1), strings.Replace(s, "</head>", `<meta name="robots" content="noindex"></head>`, 1), strings.Replace(s, "</head>", `<link rel="canonical" href="https://example.org/a"></head>`, 1), s + `<a href="https://a.r10.to/test">x</a>`, s + `<a href="#missing">x</a>`} {
		if Check(bad, "https://example.org/a") == nil {
			t.Fatal("bad page accepted")
		}
	}
}
func TestWriteRefusesOverwrite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a")
	if e := WriteNew(p, []byte("one")); e != nil {
		t.Fatal(e)
	}
	if WriteNew(p, []byte("two")) == nil {
		t.Fatal("overwrite")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "one" {
		t.Fatal("changed")
	}
	if Hash(b) != Hash([]byte("one")) {
		t.Fatal("hash")
	}
}
