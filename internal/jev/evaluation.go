package jev

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type Answer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    *float64           `json:"confidence,omitempty"`
	Noul          *float64           `json:"noul,omitempty"`
	Score         *float64           `json:"score,omitempty"`
	Legend        map[string]string  `json:"legend,omitempty"`
}
type Usage struct {
	InputTokens  *int64 `json:"input_tokens"`
	OutputTokens *int64 `json:"output_tokens"`
}
type Response struct {
	Model   string            `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage   Usage             `json:"usage"`
}

func unit(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 1 }

// Validate against the submitted questions, not guessed or fabricated defaults.
func ParseResponse(raw []byte, request Request) (Response, error) {
	var r Response
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, errors.New("malformed Jev JSON response")
	}
	if r.Model == "" || r.Answers == nil {
		return r, errors.New("missing Jev model or answers")
	}
	for _, v := range []*int64{r.Usage.InputTokens, r.Usage.OutputTokens} {
		if v != nil && *v < 0 {
			return r, errors.New("negative Jev usage")
		}
	}
	for id, q := range request.Questions {
		a, ok := r.Answers[id]
		if !ok {
			return r, fmt.Errorf("missing answer: %s", id)
		}
		if a.Type != q.Type {
			return r, fmt.Errorf("answer type mismatch: %s", id)
		}
		if a.Type == "noul" {
			if a.Noul == nil || !unit(*a.Noul) {
				return r, fmt.Errorf("invalid Noul: %s", id)
			}
			if a.Confidence != nil && !unit(*a.Confidence) {
				return r, fmt.Errorf("invalid confidence: %s", id)
			}
			continue
		}
		if a.Confidence == nil || !unit(*a.Confidence) {
			return r, fmt.Errorf("missing/invalid confidence: %s", id)
		}
		var keys []string
		switch a.Type {
		case "choice":
			criteria, ok := q.Criteria.(map[string]string)
			if !ok {
				return r, errors.New("unsupported Choice criteria")
			}
			for key := range criteria {
				keys = append(keys, key)
			}
			if _, ok := criteria[a.Choice]; !ok {
				return r, fmt.Errorf("invalid Choice option: %s", id)
			}
		case "score":
			criteria, ok := q.Criteria.([]string)
			if !ok || len(criteria) < 2 {
				return r, errors.New("unsupported Score criteria")
			}
			for i := range criteria {
				key := strconv.Itoa(i)
				keys = append(keys, key)
				if a.Legend[key] != criteria[i] {
					return r, fmt.Errorf("invalid Score legend: %s", id)
				}
			}
			if a.Score == nil || math.IsNaN(*a.Score) || *a.Score < 0 || *a.Score > float64(len(criteria)-1) {
				return r, fmt.Errorf("invalid Score: %s", id)
			}
		default:
			return r, fmt.Errorf("unsupported answer type: %s", a.Type)
		}
		if len(a.Probabilities) != len(keys) {
			return r, fmt.Errorf("probability options mismatch: %s", id)
		}
		sum, maxP, expected := 0.0, 0.0, 0.0
		for _, key := range keys {
			p, ok := a.Probabilities[key]
			if !ok || !unit(p) {
				return r, fmt.Errorf("invalid probability: %s", id)
			}
			sum += p
			maxP = math.Max(maxP, p)
			if a.Type == "score" {
				i, _ := strconv.Atoi(key)
				expected += float64(i) * p
			}
		}
		// The native API may round probabilities to two decimals.
		if math.Abs(sum-1) > float64(len(keys))*.005+1e-6 {
			return r, fmt.Errorf("probabilities do not sum to one: %s", id)
		}
		if a.Type == "choice" && a.Probabilities[a.Choice]+.011 < maxP {
			return r, fmt.Errorf("Choice is not highest probability: %s", id)
		}
		if a.Type == "score" && math.Abs(*a.Score-expected) > .01*float64(len(keys)) {
			return r, fmt.Errorf("Score expectation mismatch: %s", id)
		}
	}
	return r, nil
}

// Probability keeps Noul as a native probability and normalizes a Score rubric.
func Probability(a Answer) (float64, error) {
	switch a.Type {
	case "choice":
		p, ok := a.Probabilities["yes"]
		if !ok || !unit(p) {
			return 0, errors.New("missing yes probability")
		}
		return p, nil
	case "noul":
		if a.Noul != nil && unit(*a.Noul) {
			return *a.Noul, nil
		}
	case "score":
		if a.Score != nil && len(a.Legend) > 1 {
			return *a.Score / float64(len(a.Legend)-1), nil
		}
	}
	return 0, errors.New("invalid probability answer")
}

type Evaluation struct {
	Source             string              `json:"source"`
	SourceID           string              `json:"source_id"`
	Version            string              `json:"evaluation_version"`
	Model              string              `json:"model"`
	ProductRole        string              `json:"product_role"`
	ResearchValue      float64             `json:"research_value"`
	ProblemSpecificity float64             `json:"problem_specificity"`
	ComparisonValue    float64             `json:"comparison_value"`
	LongtailPotential  float64             `json:"longtail_potential"`
	ContentValue       float64             `json:"content_value"`
	CommodityRisk      float64             `json:"commodity_risk"`
	Confidences        map[string]*float64 `json:"confidences"`
	AverageConfidence  *float64            `json:"average_confidence"`
	MinimumConfidence  *float64            `json:"minimum_confidence"`
	OpportunityScore   float64             `json:"opportunity_score"`
	StateVersion       int                 `json:"state_version"`
	StateHash          string              `json:"state_hash"`
	RawResponse        json.RawMessage     `json:"raw_response"`
	EvaluatedAt        time.Time           `json:"evaluated_at"`
}

func NewEvaluation(source, id, version, stateHash string, r Response, raw []byte, c ScoreConfig) (Evaluation, error) {
	e := Evaluation{Source: source, SourceID: id, Version: version, Model: r.Model, ProductRole: r.Answers["product_role"].Choice, StateVersion: EvaluationStateVersion, StateHash: stateHash, RawResponse: append(json.RawMessage(nil), raw...), EvaluatedAt: time.Now().UTC(), Confidences: map[string]*float64{}}
	values := []*float64{&e.ResearchValue, &e.ProblemSpecificity, &e.ComparisonValue, &e.LongtailPotential, &e.ContentValue, &e.CommodityRisk}
	for i, id := range Axes {
		v, err := Probability(r.Answers[id])
		if err != nil {
			return e, err
		}
		*values[i] = v
	}
	count, sum, min := 0, 0.0, 1.0
	for id, a := range r.Answers {
		if id != "product_role" {
			found := false
			for _, axis := range Axes {
				found = found || id == axis
			}
			if !found {
				continue
			}
		}
		e.Confidences[id] = a.Confidence
		if a.Confidence != nil {
			count++
			sum += *a.Confidence
			min = math.Min(min, *a.Confidence)
		}
	}
	if count > 0 {
		avg := sum / float64(count)
		e.AverageConfidence = &avg
		e.MinimumConfidence = &min
	}
	e.OpportunityScore = c.Calculate(e)
	return e, nil
}
