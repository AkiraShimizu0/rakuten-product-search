package editorial

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestHumanClaimSafety(t *testing.T) {
	page, err := HumanArticle(`<html><body><main><h1>配置</h1><a rel="sponsored" href="https://hb.afl.rakuten.co.jp/example">販売情報</a></main></body></html>`, "2026-10-06")
	if err != nil {
		t.Fatal(err)
	}
	mapping := regexp.MustCompile(`(?s)<script type="application/json" id="claim-source-map">(.*?)</script>`).FindStringSubmatch(page)
	var decoded struct {
		Claims  []Claim  `json:"claims"`
		Removed []string `json:"removed_claims"`
	}
	if len(mapping) != 2 || json.Unmarshal([]byte(mapping[1]), &decoded) != nil {
		t.Fatal("invalid mapping")
	}
	expected := map[string]string{"CP01": "P14", "CP02": "P14", "CP03-D6": "P14", "CP05": "P12", "CP06": "P14", "CP08": "P16", "CP09": "P14,P16", "CP10": "P16"}
	if len(decoded.Claims) != len(expected) {
		t.Fatal("missing claim")
	}
	for _, c := range decoded.Claims {
		if strings.Join(c.Sources, ",") != expected[c.ID] {
			t.Fatalf("source mismatch: %s", c.ID)
		}
	}
	for _, m := range regexp.MustCompile(`data-claim-id="([^"]+)"`).FindAllStringSubmatch(page, -1) {
		for _, id := range strings.Fields(m[1]) {
			if _, ok := expected[id]; !ok {
				t.Fatalf("unknown claim: %s", id)
			}
		}
	}
	for _, retained := range []string{"数値の余白、清掃頻度、水洗い可否、交換周期は未再確認", "1日たばこ5本相当", "水平で安定した面", "相互換算できません", "CADR60cfm", "約100m³/h", "実機を使用した性能テスト・使用感の紹介ではありません", "編集上の判断", "外径17mm・高さ40mm"} {
		if !strings.Contains(page, retained) {
			t.Fatalf("lost qualification/fact: %s", retained)
		}
	}
	for _, removed := range []string{"周囲38cm", "交換6〜8月", "清掃2〜4週", "実際に使ってみた", "おすすめNo.1", "{{"} {
		if strings.Contains(page, removed) {
			t.Fatalf("unsafe/unresolved text: %s", removed)
		}
	}
	if len(decoded.Removed) != 2 || decoded.Removed[0] != "CP04" || decoded.Removed[1] != "CP07" {
		t.Fatal("removed assertions restored")
	}
}

func TestDimensionEnvelopesKeepOfficialScale(t *testing.T) {
	diagram := DimensionDiagram()
	rects := regexp.MustCompile(`<rect[^>]*width="([0-9.]+)" height="([0-9.]+)"`).FindAllStringSubmatch(diagram, -1)
	if len(rects) != 2 {
		t.Fatal("two verified models only")
	}
	dimensions := [][2]float64{{190, 330}, {173, 290}}
	for i, r := range rects {
		width, _ := strconv.ParseFloat(r[1], 64)
		height, _ := strconv.ParseFloat(r[2], 64)
		if width/dimensions[i][0] != 0.6 || height/dimensions[i][1] != 0.6 {
			t.Fatal("distorted scale")
		}
	}
	if strings.Count(diagram, `role="img"`) != 2 || strings.Count(diagram, "<desc") != 2 || strings.Contains(diagram, "Core Mini") {
		t.Fatal("inaccessible or invented third model")
	}
	if !strings.Contains(PlacementDiagram("test"), "必要距離ではなく") {
		t.Fatal("schematic mistaken for installation instructions")
	}
}

func TestMissingFrozenInputsRefused(t *testing.T) {
	for _, s := range []string{`<body><h1>配置</h1></body>`, `<body><a rel="sponsored" href="https://example.com">link</a></body>`} {
		if _, err := HumanArticle(s, "2026-10-06"); err == nil {
			t.Fatal("missing audited input accepted")
		}
	}
}
