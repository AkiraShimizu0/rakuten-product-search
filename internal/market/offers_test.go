package market

import (
	"reflect"
	"testing"
)

func TestOfferSummary(t *testing.T) {
	in := []OfferSnapshot{{Source: "rakuten", ID: "a:1", Title: "電動シュレッダー P5GCX2", Shop: "a", Price: 1000, Reviews: 0, Family: "F"}, {Source: "rakuten", ID: "b:1", Title: "電動シュレッダー P5GCX2", Shop: "b", Price: 3000, Reviews: 10, Family: "F"}}
	x, e := SummarizeOffers(in)
	if e != nil {
		t.Fatal(e)
	}
	if x["unique_listings"] != 2 || x["unique_shops"] != 2 || x["price_median"] != float64(2000) || x["family_proxy_hhi"] != float64(1) {
		t.Fatal(x)
	}
	if !reflect.DeepEqual(x["availability"], map[string]int{"unknown": 2}) {
		t.Fatal("missing availability became stock or zero")
	}
	for i := 0; i < 10; i++ {
		y, e := SummarizeOffers(in)
		if e != nil || !reflect.DeepEqual(x, y) {
			t.Fatal("unstable aggregation")
		}
	}
	in = append(in, in[0])
	x, e = SummarizeOffers(in)
	if e != nil || x["unique_listings"] != 2 || x["duplicate_rows"] != 1 {
		t.Fatal("dedupe failure")
	}
}
func TestOfferClasses(t *testing.T) {
	for title, want := range map[string]string{"シュレッダー専用オイル": "maintenance", "明光商会 MSパック 紐付": "waste_bags", "シュレッダー 交換刃": "replacement_parts", "手動シュレッダー": "manual", "電動シュレッダー 中古 展示品 セット A3": "body_candidate_unverified", "3M ラベルシール": "other_category"} {
		got, _ := OfferClass(title)
		if got != want {
			t.Fatalf("%s: %s", title, got)
		}
	}
	_, fs := OfferClass("電動シュレッダー 中古 展示品 セット A3")
	if len(fs) != 4 {
		t.Fatal("lost overlapping condition flags")
	}
}
func TestOfferMissingMeasurements(t *testing.T) {
	if _, e := SummarizeOffers(nil); e == nil {
		t.Fatal("empty accepted")
	}
	if _, e := SummarizeOffers([]OfferSnapshot{{Source: "r", ID: "x", Shop: "s", Title: "電動シュレッダー", Price: 0}}); e == nil {
		t.Fatal("missing price became zero")
	}
}

func TestMaintenanceBonusIsNotAnAccessory(t *testing.T) {
	class, flags := OfferClass("シュレッダー 電動 6枚細断 SP-SRD01【レビュー報告でメンテナンスシート】")
	if class != "body_candidate_unverified" || !reflect.DeepEqual(flags, []string{"maintenance_bonus_marked"}) {
		t.Fatal(class, flags)
	}
}
