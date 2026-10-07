package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDiversificationImmutable(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "x.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	f := []FamilyRecord{{Source: "x", SourceID: "a", ExactID: "a", FamilyID: "f", JSON: []byte(`{}`)}}
	r := []DiversityRank{{Method: "cap2", Source: "x", SourceID: "a", Original: 1, Rank: 1}}
	save := func(ch, ih string) error { return s.SaveDiversification(ctx, "v", "c", ch, ih, []byte(`{}`), f, r) }
	if e = save("h", "i"); e != nil {
		t.Fatal(e)
	}
	if e = save("h", "i"); e != nil {
		t.Fatal(e)
	}
	if save("changed", "i") == nil || save("h", "changed") == nil {
		t.Fatal("version overwrote different data")
	}
	var n int
	s.db.QueryRow(`SELECT count(*) FROM product_family_assignments`).Scan(&n)
	if n != 1 {
		t.Fatal(n)
	}
	s.db.QueryRow(`SELECT count(*) FROM products`).Scan(&n)
	if n != 0 {
		t.Fatal("products changed")
	}
}
