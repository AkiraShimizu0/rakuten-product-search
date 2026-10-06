package foundation

import (
	"jev-money-engine/internal/product"
	"testing"
)

func TestCandidateFamilies(t *testing.T) {
	cases := []struct {
		a, b string
		same bool
	}{
		{"アイリスオーヤマ ＰＣ－ＭＡ４－Ｂ ブラック", "IRIS OHYAMA PC - MA4 - W ホワイト", true},
		{"アイリスオーヤマ IJD-P20", "アイリスオーヤマ IJD-P20 STMX-920 物干しセット", true},
		{"アイリスオーヤマ PC-MA2", "アイリスオーヤマ PC-MA4", false},
		{"シャープ CV-T71-W ホワイト", "シャープ CV-R71-W ホワイト", false},
		{"アイリスオーヤマ SMS06 メンテナンスシート", "アイリスオーヤマ SMS06 本体", false},
		{"アイリスオーヤマ シュレッダー P5GCX2 A4", "アイリスオーヤマ シュレッダー P2HT A4", false},
		{"ナカバヤシ NSE-DTM01LG A4", "ナカバヤシ NSE-DTM01LG", true},
		{"コロナ CD-P63A3", "コロナ CD-P63A3 送料無料", true},
	}
	for _, c := range cases {
		same := CandidateFamily(c.a, "a") == CandidateFamily(c.b, "b")
		if same != c.same {
			t.Errorf("%q / %q same=%v", c.a, c.b, same)
		}
	}
}
func TestSamplingStable(t *testing.T) {
	p := []product.Product{}
	for i := 0; i < 99; i++ {
		p = append(p, product.Product{Source: "r", SourceID: string(rune(1000 + i)), Price: int64(i)})
	}
	a, e := Stratified(p, 30)
	if e != nil || len(a) != 90 {
		t.Fatal(e)
	}
	b, e := Stratified(p, 30)
	if e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	for i, x := range a {
		if seen[x.Key()] || x.Key() != b[i].Key() {
			t.Fatal("unstable or duplicate sample")
		}
		seen[x.Key()] = true
	}
	if _, e := Stratified(p[:89], 30); e == nil {
		t.Fatal("undersized pool accepted")
	}
}
func TestPairMetrics(t *testing.T) {
	m := Pairwise([]string{"a", "a", "b", "b"}, []string{"x", "x", "x", "y"})
	if m.TP != 1 || m.FP != 2 || m.FN != 1 {
		t.Fatal(m)
	}
	if Sensitivity(64, 60, 1, 1, 0, 10) != "ROBUST" || Sensitivity(64, 58, 1, 1, 0, 10) != "SAMPLE_SENSITIVE" || Sensitivity(64, 64, 1, 1, 0, 7) != "INCONCLUSIVE" {
		t.Fatal("sensitivity")
	}
}
