package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIncompleteNeverReadsPrivateKey(t *testing.T) {
	d := t.TempDir()
	rows := [][]string{}
	for i := 0; i < 40; i++ {
		rows = append(rows, []string{fmt.Sprint(i)})
	}
	blind := filepath.Join(d, "blind.csv")
	if e := writeCSV(blind, []string{"review_id"}, rows); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(d, "day3-judge-1.json"), []byte(`[]`), 0600)
	_, e := Analyze(blind, filepath.Join(d, "MUST_NOT_OPEN_KEY"), d, d, 1)
	if e == nil || !strings.Contains(e.Error(), "private key not opened") {
		t.Fatal(e)
	}
}
func TestSyntheticComparisonAndMissingJoin(t *testing.T) {
	d := t.TempDir()
	rows, keyRows := [][]string{}, [][]string{}
	judges := [3][]Judge{}
	for i := 0; i < 40; i++ {
		id := fmt.Sprint(i)
		group := "top"
		score := 5
		if i >= 20 {
			group = "random"
			score = 3
		}
		if i%3 == 0 {
			score--
		}
		rows = append(rows, []string{id, "fixture", "", "", "", ""})
		keyRows = append(keyRows, []string{id, group, "fixture", id, "true", fmt.Sprint(score * 20), "80", "80", "80", "80", "80", "80"})
		for j := 0; j < 3; j++ {
			judges[j] = append(judges[j], Judge{id, score, "fixture理由"})
		}
	}
	blind, key := filepath.Join(d, "blind.csv"), filepath.Join(d, "key.csv")
	writeCSV(blind, []string{"review_id", "product_name", "judge_1_score", "judge_2_score", "judge_3_score", "judge_notes"}, rows)
	writeCSV(key, []string{"review_id", "group", "source", "source_id", "gate_pass", "llm_overall", "buyer_problem_clarity", "comparison_depth", "wrong_choice_risk", "audience_specificity", "independent_value_potential", "investigation_value"}, keyRows)
	for j := 0; j < 3; j++ {
		b, _ := json.Marshal(judges[j])
		os.WriteFile(filepath.Join(d, fmt.Sprintf("day3-judge-%d.json", j+1)), b, 0600)
	}
	r, e := Analyze(blind, key, d, d, 20261004)
	if e != nil || r["decision"] != "Strong GO" || r["joined"] != 40 {
		t.Fatal(r, e)
	}
	if r["spearman"].(Obj)["llm_overall"] != 1. {
		t.Fatal("correlation")
	}
	keyRows[0][0] = "missing"
	writeCSV(key, []string{"review_id", "group", "source", "source_id", "gate_pass", "llm_overall", "buyer_problem_clarity", "comparison_depth", "wrong_choice_risk", "audience_specificity", "independent_value_potential", "investigation_value"}, keyRows)
	if _, e = Analyze(blind, key, d, d, 20261004); e == nil {
		t.Fatal("missing join accepted")
	}
}
