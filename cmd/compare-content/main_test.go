package main

import (
	"encoding/json"
	cv "jev-money-engine/internal/contentvalidation"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEvidenceJoins(t *testing.T) {
	for _, bad := range []string{"valid", "orphan", "unknown", "duplicate", "missing", "unsupported"} {
		t.Run(bad, func(t *testing.T) {
			cells := []evidenceCell{{Model: "M", Field: "capacity", Status: "qualified", SourceIDs: []string{"S"}}}
			claims := []evidenceClaim{{ID: "C", Status: "verified", Used: true, SourceIDs: []string{"S"}}}
			switch bad {
			case "orphan":
				claims[0].SourceIDs = nil
			case "unknown":
				claims[0].SourceIDs = []string{"X"}
			case "duplicate":
				claims = append(claims, claims[0])
			case "missing":
				cells = nil
			case "unsupported":
				claims[0].Status = "unsupported"
			}
			err := validateEvidence([]string{"S"}, []string{"M"}, []string{"capacity"}, cells, claims)
			if (err == nil) != (bad == "valid") {
				t.Fatalf("unexpected validity: %v", err)
			}
		})
	}
}

func fixture() (map[string]string, [3][]cv.Rating) {
	keys := map[string]string{"R01": "prototype", "R02": "a", "R03": "b", "R04": "c"}
	var js [3][]cv.Rating
	for j := range js {
		for _, id := range []string{"R01", "R02", "R03", "R04"} {
			n := 4
			if id == "R01" {
				n = 5
			}
			js[j] = append(js[j], cv.Rating{ID: id, Decision: n, Evidence: n, Comparison: n, Action: n, Clarity: n, Overall: n, Notes: "fixture"})
		}
	}
	return keys, js
}
func TestAggregateValidation(t *testing.T) {
	for _, bad := range []string{"missing", "score", "duplicate", "unknown", "key"} {
		t.Run(bad, func(t *testing.T) {
			keys, js := fixture()
			switch bad {
			case "missing":
				js[0] = js[0][:3]
			case "score":
				js[1][0].Overall = 6
			case "duplicate":
				js[0][1].ID = "R01"
			case "unknown":
				js[2][3].ID = "R99"
			case "key":
				keys["R04"] = "a"
			}
			if _, _, _, _, e := aggregate(keys, js); e == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}
func TestDeterministicAggregation(t *testing.T) {
	keys, js := fixture()
	a, b, d, w, e := aggregate(keys, js)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 20; i++ {
		x, y, z, v, err := aggregate(keys, js)
		if err != nil || !reflect.DeepEqual(a, x) || b != y || d != z || w != v {
			t.Fatal("unstable aggregation")
		}
	}
	if b != 4 || d != [3]int{1, 1, 1} || w != 3 || cv.Decision(a[0].Means[5], b, a[0].Means[1], w, 0) != "Strong GO" {
		t.Fatal("unexpected scores/gate")
	}
}
func prepareFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range []string{"protocol.md", "prototype-v2.md", "claims-v2.json", "evidence-matrix-v2.json", "source-manifest-v2.json", "claim-review-2.json", "judge-prompt.md"} {
		if e := os.WriteFile(filepath.Join(dir, n), []byte("fixture"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	cs := []candidate{{"prototype", "one"}, {"a", "two"}, {"b", "three"}, {"c", "four"}}
	if e := writeNew(dir, "candidate-summaries.json", cs); e != nil {
		t.Fatal(e)
	}
	return dir
}
func TestPrepareOverwriteAndSeed(t *testing.T) {
	a, b := prepareFixture(t), prepareFixture(t)
	if e := prepare(a); e != nil {
		t.Fatal(e)
	}
	if e := prepare(b); e != nil {
		t.Fatal(e)
	}
	x, _ := os.ReadFile(filepath.Join(a, "blind-input.json"))
	y, _ := os.ReadFile(filepath.Join(b, "blind-input.json"))
	if string(x) != string(y) {
		t.Fatal("non-deterministic blinding")
	}
	if e := prepare(a); e == nil {
		t.Fatal("existing output overwritten")
	}
	z, _ := os.ReadFile(filepath.Join(a, "blind-input.json"))
	if string(x) != string(z) {
		t.Fatal("old output changed")
	}
}

func TestExplicitProtocolSeed(t *testing.T) {
	a, b := prepareFixture(t), prepareFixture(t)
	for _, dir := range []string{a, b} {
		if e := prepareWithSeed(dir, 20261017); e != nil {
			t.Fatal(e)
		}
	}
	var m manifest
	if e := read(a, "blind-manifest.json", &m); e != nil {
		t.Fatal(e)
	}
	if m.Seed != 20261017 {
		t.Fatal("protocol seed ignored")
	}
	x, _ := os.ReadFile(filepath.Join(a, "blind-input.json"))
	y, _ := os.ReadFile(filepath.Join(b, "blind-input.json"))
	if string(x) != string(y) {
		t.Fatal("same explicit seed not reproducible")
	}
}
func TestFrozenInputsAndCompleteJudges(t *testing.T) {
	dir := prepareFixture(t)
	if e := prepare(dir); e != nil {
		t.Fatal(e)
	}
	_, js := fixture()
	for i, n := range []string{"judge-1.json", "judge-2.json", "judge-3.json"} {
		if e := writeNew(dir, n, js[i]); e != nil {
			t.Fatal(e)
		}
	}
	if e := os.WriteFile(filepath.Join(dir, "prototype-v2.md"), []byte("changed"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := analyze(dir); e == nil {
		t.Fatal("changed prototype accepted")
	}
	if _, e := os.Stat(filepath.Join(dir, "score-summary.json")); !os.IsNotExist(e) {
		t.Fatal("partial result created")
	}
}
func TestAnalyzeAndOverwriteRefusal(t *testing.T) {
	dir := prepareFixture(t)
	if e := prepare(dir); e != nil {
		t.Fatal(e)
	}
	_, js := fixture()
	for i, n := range []string{"judge-1.json", "judge-2.json", "judge-3.json"} {
		if e := writeNew(dir, n, js[i]); e != nil {
			t.Fatal(e)
		}
	}
	if e := analyze(dir); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(dir, "score-summary.json"))
	if e != nil || !json.Valid(b) {
		t.Fatal("invalid output")
	}
	if e := analyze(dir); e == nil {
		t.Fatal("result overwrite accepted")
	}
}
