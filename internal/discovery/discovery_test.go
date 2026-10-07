package discovery

import (
	"encoding/json"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/research"
	"os"
	"path/filepath"
	"testing"
)

func TestPartialCheckpointsAreNotCompleted(t *testing.T) {
	x := Evaluated{Product: product.Product{Source: "r", SourceID: "a"}}
	if complete(x) {
		t.Fatal("pending placeholder marked completed")
	}
	x.Evaluation.Model = jev.DefaultModel
	if !complete(x) {
		t.Fatal("saved success")
	}
	x.Evaluation.Model = "different-model"
	if complete(x) {
		t.Fatal("model changed")
	}
}
func TestMalformedLedgerStops(t *testing.T) {
	p := t.TempDir()
	if e := os.WriteFile(filepath.Join(p, "cost-ledger.json"), []byte("truncated"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := readLedger(p); e == nil {
		t.Fatal("silent billing reset")
	}
}
func TestDecodedQuestionRoundTrip(t *testing.T) {
	b, _ := json.Marshal(jev.Questions())
	var q map[string]jev.Question
	if e := json.Unmarshal(b, &q); e != nil {
		t.Fatal(e)
	}
	got, _ := json.Marshal(q)
	if string(got) != string(b) {
		t.Fatal("frozen question mismatch")
	}
}

func TestDeterministicSelection(t *testing.T) {
	v := []product.Product{{Source: "r", SourceID: "a"}, {Source: "r", SourceID: "b"}, {Source: "r", SourceID: "c"}}
	a := Selection(v, 2)
	v[0], v[2] = v[2], v[0]
	b := Selection(v, 2)
	if a[0].Key() != b[0].Key() || a[1].Key() != b[1].Key() {
		t.Fatal("unstable")
	}
}
func TestPreregisteredDrop(t *testing.T) {
	s := Stats{Unique: 90, DescriptionCoverage: 1, Eligible: 30, EligibleRate: 1. / 3, Families: 20, LargestFamily: .1}
	if len(Drop(s)) != 0 {
		t.Fatal(Drop(s))
	}
	s.Eligible = 14
	if len(Drop(s)) != 1 {
		t.Fatal(Drop(s))
	}
	s.Unique = 49
	if len(Drop(s)) != 2 {
		t.Fatal(Drop(s))
	}
}
func TestOutputRefusal(t *testing.T) {
	p := filepath.Join(t.TempDir(), "new")
	if e := research.NewDir(p); e != nil {
		t.Fatal(e)
	}
	if e := research.NewDir(p); e == nil {
		t.Fatal("overwrite")
	}
	_ = os.WriteFile(filepath.Join(p, "data"), []byte("x"), 0600)
}
