package validation

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func Report(path string, r Comparison) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Day 3.6 Diversification Quality Validation\n\n**判定: %s**。問いはfamily cap 2が独立AI judge品質を実質的に維持したか。Day 3 Strong GO、Day 3.5 NO-GOは変更しない。Day 4は実施していない。\n", r.Decision)
	fmt.Fprintf(&b, "\n## Cohort / blind解除の検証\n\nOriginal %d、Diversified %d、overlap %d、union %d。Original only %d、Diversified only %d。新規採点%d件（%d×3）。重複review ID=0、欠損=0、1〜5整数、各judge同件数、keyとの一対一joinを検証してからblind解除。過去judge結果は使用していない。\n", r.Original.N, r.Diversified.N, r.Overlap, r.Union, r.OriginalOnly.N, r.DiversifiedOnly.N, r.TotalJudgments, r.Union)
	fmt.Fprintf(&b, "\nJudge: %s。各judgeへblind CSVだけを渡し、過去会話・score・group・family・他judge回答を共有していない。同じrubricで全unionを商品ごと1回評価。説明文はDBの全文で統一し、URL/商品IDは見せていない。\n", r.JudgeIdentity)
	fmt.Fprintln(&b, "\n## Analysis A: portfolio\n\n|指標|Original|Diversified|\n|---|---:|---:|")
	a, d := r.Original, r.Diversified
	fmt.Fprintf(&b, "|n|%d|%d|\n|平均|%.6f|%.6f|\n|中央値|%.6f|%.6f|\n|標準偏差（sample）|%.6f|%.6f|\n|>=4|%d/%d (%.1f%%)|%d/%d (%.1f%%)|\n|>=4.5|%d/%d (%.1f%%)|%d/%d (%.1f%%)|\n|=4.0|%d/%d (%.1f%%)|%d/%d (%.1f%%)|\n|=5.0|%d/%d (%.1f%%)|%d/%d (%.1f%%)|\n", a.N, d.N, a.Mean, d.Mean, a.Median, d.Median, a.SD, d.SD, a.GE4, a.N, 100*a.GE4Rate, d.GE4, d.N, 100*d.GE4Rate, a.GE45, a.N, 100*a.GE45Rate, d.GE45, d.N, 100*d.GE45Rate, a.EQ4, a.N, 100*a.EQ4Rate, d.EQ4, d.N, 100*d.EQ4Rate, a.EQ5, a.N, 100*a.EQ5Rate, d.EQ5, d.N, 100*d.EQ5Rate)
	fmt.Fprintf(&b, "\n**Diversified − Original = %+.6f**。中央値差 %+.6f。Bootstrap 95%% CI [%+.6f, %+.6f]。seed=%d、%d回。\n\n方法: %s。union全商品のlisting単位を復元抽出し、同じ商品のdraw countを両portfolioに共有する。群の分母はresample内のmembership数。完全独立2群のCIは主指標に使用しない。\n", r.Difference, r.MedianDifference, r.PortfolioCI.Lower, r.PortfolioCI.Upper, r.PortfolioCI.Seed, r.PortfolioCI.Samples, r.PortfolioCI.Method)
	fmt.Fprintln(&b, "\n|judge|Original平均|Diversified平均|差|\n|---|---:|---:|---:|")
	for i, j := range r.Judges {
		fmt.Fprintf(&b, "|%d|%.4f|%.4f|%+.4f|\n", i+1, j.Original, j.Diversified, j.Difference)
	}
	fmt.Fprintln(&b, "\n## Analysis B: changed items\n\n|指標|Original only|Diversified only|\n|---|---:|---:|")
	a, d = r.OriginalOnly, r.DiversifiedOnly
	fmt.Fprintf(&b, "|n|%d|%d|\n|平均|%.6f|%.6f|\n|中央値|%.6f|%.6f|\n|SD|%.6f|%.6f|\n|>=4.5|%d/%d (%.1f%%)|%d/%d (%.1f%%)|\n\n差 %+.6f。Changed-items bootstrap 95%% CI [%+.6f, %+.6f]。disjointな7対7から群内復元抽出。\n", a.N, d.N, a.Mean, d.Mean, a.Median, d.Median, a.SD, d.SD, a.GE45, a.N, 100*a.GE45Rate, d.GE45, d.N, 100*d.GE45Rate, r.ChangedDifference, r.ChangedCI.Lower, r.ChangedCI.Upper)
	fmt.Fprintf(&b, "\n共通13商品のscoreは両portfolioに同じ値として入り、observed差から打ち消し合う。Portfolio差 = (%d/20) × changed-items差。\n", r.OriginalOnly.N)
	fmt.Fprintln(&b, "\n### 入れ替え一覧とbest/worst例\n\n対応はremovedを元rank昇順、addedをdiversified rank昇順で結ぶ説明用の対応。アルゴリズムの因果的1:1置換ではない。単一対の数値は対応方法に依存する。以下の14商品全てのClaude/AI値はchanged-items CSVでも確認できる。\n\n|removed (Claude / AI)|added (Claude / AI)|AI差|Claude差|元Topにないfamily|\n|---|---|---:|---:|---|")
	for _, x := range r.Replacements {
		fmt.Fprintf(&b, "|#%d %s (%d / %.3f)|#%d %s (%d / %.3f)|%+.3f|%+.0f|%t|\n", x.Removed.OriginalRank, clean(x.Removed.Name), x.Removed.Claude, x.Removed.Mean, x.Added.DiversifiedRank, clean(x.Added.Name), x.Added.Claude, x.Added.Mean, x.AIDifference, x.ClaudeDifference, x.NewFamily)
	}
	rr := append([]Replacement{}, r.Replacements...)
	sort.SliceStable(rr, func(i, j int) bool { return rr[i].AIDifference < rr[j].AIDifference })
	if len(rr) > 0 {
		for _, x := range []struct {
			name  string
			value Replacement
		}{{"worst", rr[0]}, {"best", rr[len(rr)-1]}} {
			q := x.value
			fmt.Fprintf(&b, "\n- %s（説明用pair）: %s → %s。AI差 %+.3f、Claude差 %+.0f。新規候補judge notes: %s\n", x.name, q.Removed.Name, q.Added.Name, q.AIDifference, q.ClaudeDifference, strings.Join(q.Added.Notes[:], " / "))
		}
	}
	fmt.Fprintf(&b, "\nCatastrophic incoming（平均<=%.1f）%d件、低品質incoming（平均<3）%d件。新規family数と全portfolioのfamily増加は下表で比較する。\n", r.Policy.CatastrophicFloor, r.Catastrophic, r.LowIncoming)
	fmt.Fprintln(&b, "\n## 固定されたdiversityとClaude評価\n\n|指標|Original|Diversified|\n|---|---:|---:|")
	x, y := r.OriginalDiversity, r.DiversifiedDiversity
	fmt.Fprintf(&b, "|family数|%d|%d|\n|最大family|%d/%d|%d/%d|\n|最大share|%.1f%%|%.1f%%|\n|IonicBreeze|%d|%d|\n|HHI|%.4f|%.4f|\n|既知brand数|%d|%d|\n|暫定buyer-theme数|%d|%d|\n|Claude平均（元0〜100点）|%.2f|%.2f|\n", x.Families, y.Families, x.LargestFamily, x.N, y.LargestFamily, y.N, 100*x.LargestShare, 100*y.LargestShare, x.IonicBreeze, y.IonicBreeze, x.HHI, y.HHI, x.Brands, y.Brands, x.Themes, y.Themes, x.ClaudeMean, y.ClaudeMean)
	fmt.Fprintf(&b, "\nunion %d商品のClaude overall ↔ 新judge平均 Spearman = %v（tieは平均rank）。Day 3の0.657は別sampleであり、数値だけから優劣は言えない。今回unionは元reranker上位の狭い範囲に限られ、相関から329件全体のランキング性能は判断しない。\n", r.Union, r.Spearman)
	fmt.Fprintln(&b, "\n## 事前ルールと判定の範囲\n\n採点前のprotocolを保存・照合した。Strong: 平均差>=−0.10、全judge差>=−0.20、>=4.5率低下<=10 percentage points、diversity維持、catastrophicなし。GO: 平均差>=−0.20、少なくとも2judge差>=−0.20、diversity維持。\n\nNO-GO triggerを優先する: 平均差<−0.20、全judgeがstrictly negative、changed差<=−0.50かつCI上限<0、平均<3の新規商品が3件以上。Catastrophicがある場合はStrongではなく、GO数値条件だけ満たしてもinconclusive。曖昧な規定の数値化は採点前に固定し、結果を見て変更していない。")
	fmt.Fprintf(&b, "\n**最終判定 %s**。%s\n", r.Decision, next(r.Decision))
	fmt.Fprintln(&b, "\n## 未解決事項\n\n- 新judgeは同じ基盤モデルの独立文脈3つ。異なるmodelや人間によるground truthを得た実験ではない。\n- 27listing/7置換という小標本。bootstrapは商品単位の有限sampleの記述的CIであり、正式な非劣性証明や市場一般化ではない。\n- 同family内の依存は追加のfamily-level bootstrapで扱っていない。物理商品familyが重なることによる評価の相関は残る。\n- 判断基準の天井効果、seller情報の真偽、暫定テーマやfamily誤分類のリスクは残る。両portfolioで平均4以上率100%という上位集合への偏りがあり、低品質商品を見分ける能力は今回検証できない。今回rubric/clusterを修正していない。\n- SEO成功、検索需要、収益、CVR、売上、他カテゴリへの一般化は証明しない。\n\n5テーマ（互換フィルター、天井設置、フィルターレス維持、小型配置、加湿運用）は固定。市場調査・記事生成・公開・新カテゴリ収集・Jev/Claude v2は実施していない。")
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.WriteString(b.String())
	return e
}
func clean(s string) string {
	s = strings.ReplaceAll(s, "|", "/")
	return strings.ReplaceAll(s, "\n", " ")
}
func next(decision string) string {
	if decision == "Strong GO" || decision == "GO" {
		return "family cap 2がAI judge品質をほぼ維持できることへの限定的なGO。次工程の5テーマ外部市場検証へ進む根拠はあるが、このタスクではDay 4を開始せず停止する。"
	}
	return "Day 4には進まず、この検証結果の確認を次工程とする。Day 3.5 NO-GOは変更しない。"
}
