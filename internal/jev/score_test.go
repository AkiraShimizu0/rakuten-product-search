package jev

import (
	"math"
	"testing"
)

func TestScore(t *testing.T) {
	c := DefaultScoreConfig()
	e := Evaluation{ResearchValue: 1, ProblemSpecificity: 1, ComparisonValue: 1, LongtailPotential: 1, ContentValue: 1, ProductRole: "replacement_consumable"}
	base := c.Calculate(e)
	if math.Abs(base-1) > 1e-9 {
		t.Fatal(base)
	}
	e.CommodityRisk = 1
	if math.Abs(c.Calculate(e)-.75) > 1e-9 {
		t.Fatal("commodity risk did not subtract")
	}
	e.ProductRole = "main_product"
	if math.Abs(c.Calculate(e)-.8) > 1e-9 {
		t.Fatal("main product adjustment")
	}
	e.ProductRole = "bundle_or_set"
	if math.Abs(c.Calculate(e)-.74) > 1e-9 {
		t.Fatal("bundle adjustment")
	}
	e = Evaluation{CommodityRisk: 1, ProductRole: "unclear"}
	if c.Calculate(e) != 0 {
		t.Fatal("lower clamp")
	}
	e = Evaluation{ResearchValue: 1, ProblemSpecificity: 1, ComparisonValue: 1, LongtailPotential: 1, ContentValue: 1, ProductRole: "main_product"}
	if c.Calculate(e) != 1 {
		t.Fatal("upper clamp")
	}
	c.RoleAdjustment["main_product"] = .5
	if c.Validate() == nil {
		t.Fatal("large role bias accepted")
	}
}
