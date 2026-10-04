package jev

import (
	"fmt"
	"math"
)

type ScoreConfig struct {
	ResearchValue      float64            `json:"research_value"`
	ProblemSpecificity float64            `json:"problem_specificity"`
	ComparisonValue    float64            `json:"comparison_value"`
	LongtailPotential  float64            `json:"longtail_potential"`
	ContentValue       float64            `json:"content_value"`
	CommodityRisk      float64            `json:"commodity_risk"`
	RoleAdjustment     map[string]float64 `json:"role_adjustment"`
}

func DefaultScoreConfig() ScoreConfig {
	return ScoreConfig{.22, .18, .20, .18, .22, .25, map[string]float64{"main_product": .05, "replacement_consumable": 0, "accessory": 0, "bundle_or_set": -.01, "unclear": -.03}}
}
func (c ScoreConfig) Validate() error {
	for _, v := range []float64{c.ResearchValue, c.ProblemSpecificity, c.ComparisonValue, c.LongtailPotential, c.ContentValue, c.CommodityRisk} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return fmt.Errorf("invalid score weight")
		}
	}
	for _, role := range Roles {
		v, ok := c.RoleAdjustment[role]
		if !ok || math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > .05 {
			return fmt.Errorf("role adjustments must cover each role and remain within +/-0.05")
		}
	}
	return nil
}
func (c ScoreConfig) Calculate(e Evaluation) float64 {
	v := e.ResearchValue*c.ResearchValue + e.ProblemSpecificity*c.ProblemSpecificity + e.ComparisonValue*c.ComparisonValue + e.LongtailPotential*c.LongtailPotential + e.ContentValue*c.ContentValue - e.CommodityRisk*c.CommodityRisk + c.RoleAdjustment[e.ProductRole]
	return math.Max(0, math.Min(1, v))
}
