package diversify

import (
	"encoding/json"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/rerank"
	"jev-money-engine/internal/store"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixture(id, title string, score int) Item {
	return Item{Product: product.Product{Source: "test", SourceID: id, Name: title}, Input: llmjudge.Input{ProductName: title, ProductRole: "main_product"}, Evaluation: store.Reranked{Source: "test", SourceID: id, Scores: llmjudge.Scores{Overall: score}}}
}
func TestNormalization(t *testing.T) {
	a := Normalize("【送料無料】ＳＨＡＲＰ　ＫＣ－Ｔ５０－Ｗ P10倍 クーポン対象")
	if a != "sharp kc-t50-w" {
		t.Fatal(a)
	}
	if Normalize(a) != a {
		t.Fatal("not idempotent")
	}
}
func TestExtract(t *testing.T) {
	cases := []struct{ title, model string }{{"シャープ KC-T50-W KC-U50 の前型番", "kc-t50-w"}, {"【P5倍】Dyson Hot Cool Gen1 HP10WW", "hp10ww"}, {"RHYTHM リズム 9YYA63", "9yya63"}, {"SHARP HEPA H13 PM2.5", ""}, {"Slimac UZUKAZE3 FCE-570", "fce-570"}, {"Slimac UZUKAZE3 FCE-570matter 対応", "fce-570"}, {"Levoit Core P350", "p350"}, {"Levoit Core 300", "core300"}}
	for _, c := range cases {
		if f := Extract(c.title, "main_product"); f.Model != c.model {
			t.Errorf("%s: %s", c.title, f.Model)
		}
	}
}
func TestSameAndDifferentModels(t *testing.T) {
	rows := Cluster([]Item{fixture("a", "SHARP KC-T50-W", 70), fixture("b", "シャープ KC-T50-B", 69), fixture("c", "SHARP KC-T70-W", 68), fixture("d", "SHARP KC-U50-W", 67)})
	if rows[0].Features.ExactID != rows[1].Features.ExactID || rows[0].Features.FamilyID != rows[1].Features.FamilyID {
		t.Fatal("same model variant separated")
	}
	if rows[0].Features.ExactID == rows[2].Features.ExactID || rows[0].Features.FamilyID == rows[3].Features.FamilyID {
		t.Fatal("different model/series merged")
	}
	if rows[0].Features.FamilyID != rows[2].Features.FamilyID {
		t.Fatal("size variants family split")
	}
}
func TestCompatibleFilterNotSamePhysicalProduct(t *testing.T) {
	a := fixture("a", "SHARP 互換 フィルター FZ-Y80MF", 70)
	a.Input.ProductRole = "replacement_consumable"
	b := a
	b.Product.SourceID = "b"
	b.Evaluation.SourceID = "b"
	r := Cluster([]Item{a, b})
	if r[0].Features.ExactID == r[1].Features.ExactID || r[0].Features.FamilyID != r[1].Features.FamilyID {
		t.Fatal("fit is not manufacturer identity")
	}
}
func TestCapDeterminism(t *testing.T) {
	in := []Item{fixture("a", "Ionic Breeze MIDI white", 90), fixture("b", "Ionic Breeze MIDI black", 89), fixture("c", "Ionic Breeze GRANDE", 88), fixture("d", "Ionic Little", 87), fixture("e", "SHARP KC-T50-W", 86)}
	a := Cluster(in)
	rev := append([]Item{}, in...)
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	if !reflect.DeepEqual(a, Cluster(rev)) {
		t.Fatal("input order changed clustering")
	}
	for _, m := range []string{"none", "cap1", "cap2", "cap3", "penalty", "mmr"} {
		r, e := Rank(a, m, DefaultConfig())
		if e != nil {
			t.Fatal(e)
		}
		s, _ := Rank(Cluster(rev), m, DefaultConfig())
		if !reflect.DeepEqual(r, s) {
			t.Fatal(m, "unstable")
		}
		if m == "cap2" {
			counts := map[string]int{}
			exact := map[string]bool{}
			for _, v := range r {
				f := v.Item.Features
				counts[f.FamilyID]++
				if counts[f.FamilyID] > 2 || exact[f.ExactID] {
					t.Fatal("cap/dedup violated")
				}
				exact[f.ExactID] = true
			}
			if len(r) != 3 {
				t.Fatal(len(r))
			}
		}
	}
}
func TestHHI(t *testing.T) {
	r := []Ranked{}
	for i, f := range []string{"a", "a", "b", "c"} {
		x := fixture(string(rune('a'+i)), "", 80)
		x.Features.FamilyID = f
		r = append(r, Ranked{Item: x})
	}
	m := Measure(r, 4)
	if math.Abs(m.HHI-.375) > 1e-12 || m.LargestShare != .5 {
		t.Fatal(m)
	}
	for i := 0; i < 50; i++ {
		if !reflect.DeepEqual(m, Measure(r, 4)) {
			t.Fatal("nondeterministic float sums")
		}
	}
	if Measure(nil, 0).HHI != 0 {
		t.Fatal("empty HHI")
	}
}
func TestManifestFingerprint(t *testing.T) {
	a := map[string]any{"model": "x", "system": "rubric", "temperature": nil}
	b := map[string]any{"temperature": nil, "system": "rubric", "model": "x"}
	if Fingerprint(a) != Fingerprint(b) {
		t.Fatal("map insertion order")
	}
	b["system"] = "changed"
	if Fingerprint(a) == Fingerprint(b) {
		t.Fatal("prompt not fingerprinted")
	}
	if len(SourceFingerprint()) != 64 || SourceFingerprint() != SourceFingerprint() {
		t.Fatal("source hash")
	}
	request := []byte(`{"model":"x","system":"s","messages":[{"role":"user","content":"{\"untrusted_product_data\":{\"product_name\":\"test\"}}"}]}`)
	tmpl, in, e := RequestTemplate(request)
	if e != nil || in.ProductName != "test" || tmpl["messages"] != nil {
		t.Fatal(e, in)
	}
}
func TestStableOutput(t *testing.T) {
	items := Cluster([]Item{fixture("a", "SHARP KC-T50-W", 70), fixture("b", "Dyson Hot Cool HP10", 69), fixture("c", "Ionic Breeze MIDI", 68)})
	c := DefaultConfig()
	c.Top = 3
	rankings := map[string][]Ranked{}
	for _, m := range []string{"none", "cap1", "cap2", "cap3", "penalty", "mmr"} {
		rankings[m], _ = Rank(items, m, c)
	}
	a, b := t.TempDir(), t.TempDir()
	for _, d := range []string{a, b} {
		if _, e := WriteOutputs(d, items, rankings, c, map[string]any{"fingerprint": "test"}, nil, nil); e != nil {
			t.Fatal(e)
		}
	}
	filepath.WalkDir(a, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			t.Fatal(e)
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(a, p)
			x, _ := os.ReadFile(p)
			y, _ := os.ReadFile(filepath.Join(b, rel))
			if string(x) != string(y) {
				t.Fatal("unstable", rel)
			}
		}
		return nil
	})
	var x any
	if e := json.Unmarshal(StableJSON(c), &x); e != nil {
		t.Fatal(e)
	}
}

func TestAccessorySeparation(t *testing.T) {
	a := Extract("SHARP KC-T50-W FZ-Y80MF", "accessory")
	b := Extract("SHARP KC-T50-W", "main_product")
	if a.Model != "fz-y80mf" || a.ExactID == b.ExactID || a.FamilyID == b.FamilyID {
		t.Fatal(a, b)
	}
}
func TestSavedManifest(t *testing.T) {
	request := []byte(`{"model":"x","system":"saved rubric","max_tokens":4096,"messages":[{"role":"user","content":"{\"untrusted_product_data\":{\"product_name\":\"test\"}}"}]}`)
	e := store.RerankEvidence{Source: "s", SourceID: "a", Request: request, Response: []byte(`{"model":"x"}`), EvaluatedAt: "2026-10-04T00:00:00Z"}
	r := rerank.Run{Model: "x", InputTokens: 100, OutputTokens: 20, EstimatedCost: .0004, Prices: llmjudge.DefaultPrices()}
	rb, _ := json.Marshal(r)
	cfg := json.RawMessage(`{"Provider":"anthropic","APIVersion":"2023-06-01"}`)
	m, inputs, err := Manifest([]store.RerankEvidence{e}, cfg, []json.RawMessage{rb}, "commit")
	if err != nil || len(inputs) != 1 {
		t.Fatal(err)
	}
	if m["temperature"] != nil || m["exact_model_id"] != "x" || m["calculated_estimated_cost_usd"] != .0004 {
		t.Fatal(m)
	}
	fp := m["fingerprint_sha256"]
	delete(m, "fingerprint_sha256")
	if fp != Fingerprint(m) {
		t.Fatal("manifest hash not reproducible")
	}
	bad := e
	bad.SourceID = "b"
	bad.Request = []byte(strings.Replace(string(request), "saved rubric", "changed rubric", 1))
	if _, _, err = Manifest([]store.RerankEvidence{e, bad}, cfg, []json.RawMessage{rb}, "commit"); err == nil {
		t.Fatal("mixed prompts accepted")
	}
	bad = e
	bad.Response = []byte(`{"model":"other"}`)
	if _, _, err = Manifest([]store.RerankEvidence{bad}, cfg, []json.RawMessage{rb}, "commit"); err == nil {
		t.Fatal("model mismatch accepted")
	}
}
func TestSeriesGenerationNotSecondSKU(t *testing.T) {
	a := Extract("Slimac UZUKAZE3 FCE-570matter 対応 送料無料", "main_product")
	b := Extract("uzukaze3 ウズカゼ FCE-570 シーリングファンライト", "main_product")
	if a.ExactID != b.ExactID || a.FamilyID != b.FamilyID {
		t.Fatal(a, b)
	}
}
