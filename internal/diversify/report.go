package diversify

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WriteReport(dir string, items []Item, rankings map[string][]Ranked, c Config, manifest map[string]any, summaries map[string]any) error {
	var b strings.Builder
	fmt.Fprintln(&b, "# Day 3.5 商品ファミリーと多様化\n\nDay 3のStrong GOは維持。329件の保存済みClaude評価のみ再利用、追加Jev/Claude API呼出し0。Day 4は開始していない。")
	fmt.Fprintln(&b, "\n## 固定した選択規則\n\n主方式はfamily cap 2。Level A（同一型番・色違い）の代表は最大1件、Level B（brand+series）は最大2件。元overallと全6軸は変更しない。cap1/3、family penalty=3点、MMR λ=0.08も比較。tieはDay 3と同じ固定sort。規則はAI judgeラベルを参照せず設定した。\n\nスコア単位は元Claudeの0–100。事前基準のloss≤0.20/0.30も同じ単位として適用。0–1換算は参考のみで判定には使わない。")
	fmt.Fprintf(&b, "\nAlgorithm SHA256: `%s`\n", c.AlgorithmSHA256)
	fmt.Fprintln(&b, "\n## Before / After\n\n|方式|family数|最大family|IonicBreeze|HHI|brand数（既知）|theme数|Claude平均|loss（点）|既存judge対象数|\n|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|")
	orig := Measure(rankings["none"], c.Top)
	after := Measure(rankings["cap2"], c.Top)
	for _, method := range []string{"none", "cap1", "cap2", "cap3", "penalty", "mmr"} {
		m := Measure(rankings[method], c.Top)
		s := summaries[method].(map[string]any)
		fmt.Fprintf(&b, "|%s|%d|%d/%d|%d|%.4f|%d|%d|%.3f|%.3f|%v|\n", method, m.UniqueFamilies, m.LargestFamily, m.N, m.IonicBreeze, m.HHI, m.UniqueBrands, m.UniqueThemes, m.MeanClaude, orig.MeanClaude-m.MeanClaude, s["existing_judge_covered"])
	}
	fmt.Fprintln(&b, "\n### Top5 / Top10 / Top20集中度\n\n|方式|N|family数|最大share|HHI|\n|---|---:|---:|---:|---:|")
	for _, method := range []string{"none", "cap2"} {
		for _, n := range []int{5, 10, 20} {
			m := Measure(rankings[method], n)
			fmt.Fprintf(&b, "|%s|%d|%d|%.1f%%|%.4f|\n", method, n, m.UniqueFamilies, 100*m.LargestShare, m.HHI)
		}
	}
	loss := orig.MeanClaude - after.MeanClaude
	decision := "NO-GO"
	if after.UniqueFamilies >= 12 && after.LargestFamily <= 2 && after.IonicBreeze <= 2 && loss <= .20+1e-9 {
		decision = "numerical Strong GO; cluster quality sign-off required"
	} else if after.UniqueFamilies >= 10 && after.LargestFamily <= 3 && loss <= .30+1e-9 {
		decision = "numerical GO; cluster quality sign-off required"
	}
	fmt.Fprintf(&b, "\n## 判定\n\n**%s**。cap2 loss=%.3f点。多様性の改善と、指定された厳格なスコア低下条件の充足を分離して判断する。Day 3のStrong GOを変更する判断ではない。Day 4は停止したまま。\n", decision, loss)
	fmt.Fprintln(&b, "\n## 既存AI judgeの利用限界\n\n既存Day 3の40件から被覆される商品だけの平均を下表に示す。新しく入る商品は未採点なので、diversified Top20全体の品質維持や統計的非劣性は証明できない。重複した高評価商品を除くことで被覆集合も変わるため、この平均同士は公平な比較ではない。人間ground truthではない。\n\n|方式|被覆件数|被覆部分AI平均|Day2.6既使用120件との重複|\n|---|---:|---:|---:|")
	for _, method := range []string{"none", "cap2"} {
		s := summaries[method].(map[string]any)
		fmt.Fprintf(&b, "|%s|%v|%v|%v|\n", method, s["existing_judge_covered"], s["existing_judge_covered_mean"], s["day26_holdout_overlap"])
	}

	fmt.Fprintln(&b, "\nDay2.6集合は既にjudgeで使用したデータであり、将来の未使用holdoutとは扱わない。これらのラベルをclusterルール・selectionの調整には使用していない。")
	fmt.Fprintln(&b, "\n## Family例（Level AとBを分離）")
	groups := map[string][]Item{}
	for _, x := range items {
		groups[x.Features.FamilyID] = append(groups[x.Features.FamilyID], x)
	}
	keys := []string{}
	for k, g := range groups {
		if len(g) > 1 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		a, z := groups[keys[i]], groups[keys[j]]
		if len(a) != len(z) {
			return len(a) > len(z)
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys[:min(5, len(keys))] {
		g := groups[k]
		f := g[0].Features
		fmt.Fprintf(&b, "\n### %s (%s / %s, %d listings)\n\n判定根拠: kind=%s、brand=%s、series=%sの一致。異なる型番はLevel Aでは分ける。代表: %s\n", k, f.Brand, f.Series, len(g), f.Kind, f.Brand, f.Series, g[0].Input.ProductName)
		for _, x := range g {
			fmt.Fprintf(&b, "\n- #%d `%s` / model=%s / exact=%s: %s\n", x.OriginalRank, x.Product.SourceID, x.Features.Model, x.Features.ExactID, x.Input.ProductName)
		}
	}
	fmt.Fprintln(&b, "\n## Day 4用のテーマ候補（未開始・市場未検証）\n\n詳細はdata/day3-5-day4-candidates.csv。cap2順に異なるbuyer-themeを最大5テーマ。cluster diversityをClaude評価scoreに混ぜずselectionルールとして扱う。evidence_opportunityとcontent_differentiationは既存independent_value_potentialの別名であり、新たな証拠の存在を確認した値ではない。")
	seen := map[string]bool{}
	for _, r := range rankings["cap2"][:min(c.Top, len(rankings["cap2"]))] {
		x := r.Item
		if seen[x.Features.Theme] {
			continue
		}
		seen[x.Features.Theme] = true
		fmt.Fprintf(&b, "\n- **%s**: %s（Claude %d、buyer %d、comparison %d、mistake %d、evidence proxy %d）。[%s](%s)\n", themeName(x.Features.Theme), x.Features.Theme, x.Evaluation.Scores.Overall, x.Evaluation.Scores.BuyerProblemClarity, x.Evaluation.Scores.ComparisonDepth, x.Evaluation.Scores.WrongChoiceRisk, x.Evaluation.Scores.IndependentValuePotential, x.Features.Series, SafeURL(x.Product.ItemURL))
		if len(seen) == 5 {
			break
		}
	}
	fmt.Fprintf(&b, "\n## Claude再現性\n\nprovider=%v、exact model=%v、version=llm-reranker-claude-v1、評価commit=%v。Manifestはdata/evaluation-manifests/day3-claude.json。保存済み全329 request/responseからsystem prompt・構造化rubric・adaptive thinking・effort・max tokensを照合。temperatureは未指定で、数値を捏造しない。overallは直接のholistic評価で加重和ではない。料金snapshot、入力/出力token、既存推定費用USD %vを保存。secretは含めない。\n", manifest["provider"], manifest["exact_model_id"], manifest["evaluation_git_commit"], manifest["calculated_estimated_cost_usd"])
	fmt.Fprintln(&b, "\n## 保存と再実行\n\n元DBをコピーしたday3-5-money.dbにversion付きの3 tableだけ追加。入力・設定・algorithm SHAが同じなら同一versionの再実行はno-op、違えば拒否。CSVは全329商品の原順位と各方式の順位を保持、0は未選択。元evaluationは更新しない。HHIはfamily share²の和。未知ブランドはbrand数に含めない。\n\nLevel Cはタイトルの明示キーワードに基づく暫定テーマであり、購入意図・検索需要・市場規模を確認した分類ではない。広告keywordによるテーマ誤分類と、型番がない商品のfalse splitは残る。family capはbuyer-themeの集中を完全には防止しない。")
	if review, e := os.ReadFile(filepath.Join(filepath.Dir(dir), "Day3-5-cluster-quality-review.md")); e == nil {
		b.WriteString("\n\n")
		b.Write(review)
	} else if !os.IsNotExist(e) {
		return e
	}
	return os.WriteFile(filepath.Join(filepath.Dir(dir), "Day3-5-diversification-report.md"), []byte(b.String()), 0600)
}
