package diversify

import (
	"embed"
	"sort"
	"strings"
)

// Algorithm source is embedded so version guards detect unversioned rule changes.
//
//go:embed family.go rank.go output.go manifest.go provenance.go report.go
var sources embed.FS

func SourceFingerprint() string {
	names := []string{"family.go", "rank.go", "output.go", "manifest.go", "provenance.go", "report.go"}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		v, e := sources.ReadFile(n)
		if e != nil {
			panic(e)
		}
		b.WriteString(n)
		b.WriteByte(0)
		b.WriteString(strings.ReplaceAll(string(v), "\r\n", "\n"))
		b.WriteByte(0)
	}
	return hash(b.String())
}
