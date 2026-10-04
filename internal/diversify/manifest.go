package diversify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/rerank"
	"jev-money-engine/internal/store"
	"sort"
)

func RequestTemplate(request []byte) (map[string]json.RawMessage, llmjudge.Input, error) {
	var doc map[string]json.RawMessage
	var in llmjudge.Input
	if json.Unmarshal(request, &doc) != nil {
		return nil, in, errors.New("invalid saved request")
	}
	var messages []struct{ Role, Content string }
	if json.Unmarshal(doc["messages"], &messages) != nil || len(messages) != 1 || messages[0].Role != "user" {
		return nil, in, errors.New("saved user input missing")
	}
	var payload struct {
		Input llmjudge.Input `json:"untrusted_product_data"`
	}
	if json.Unmarshal([]byte(messages[0].Content), &payload) != nil {
		return nil, in, errors.New("invalid saved product payload")
	}
	delete(doc, "messages")
	return doc, payload.Input, nil
}
func Fingerprint(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return hash(string(b))
}
func Manifest(evidence []store.RerankEvidence, config json.RawMessage, runs []json.RawMessage, commit string) (map[string]any, map[string]llmjudge.Input, error) {
	if len(evidence) == 0 {
		return nil, nil, errors.New("no saved evaluations")
	}
	inputs := map[string]llmjudge.Input{}
	var first map[string]json.RawMessage
	times := []string{}
	responses := []string{}
	unknown := 0
	for _, e := range evidence {
		tmpl, in, err := RequestTemplate(e.Request)
		if err != nil {
			return nil, nil, err
		}
		if first == nil {
			first = tmpl
		}
		if Fingerprint(tmpl) != Fingerprint(first) {
			return nil, nil, errors.New("saved request parameters differ")
		}
		key := e.Source + "\x00" + e.SourceID
		if _, exists := inputs[key]; exists {
			return nil, nil, errors.New("duplicate evaluation")
		}
		inputs[key] = in
		times = append(times, e.EvaluatedAt)
		responses = append(responses, hash(string(e.Response)))
		var resp struct{ Model string }
		if json.Unmarshal(e.Response, &resp) != nil {
			return nil, nil, errors.New("invalid saved response")
		}
		var requested string
		json.Unmarshal(first["model"], &requested)
		if resp.Model != requested {
			return nil, nil, errors.New("response model mismatch")
		}
	}
	var prompt, model string
	json.Unmarshal(first["system"], &prompt)
	json.Unmarshal(first["model"], &model)
	if prompt == "" || model == "" {
		return nil, nil, errors.New("missing saved model/prompt")
	}
	sort.Strings(times)
	totalInput, totalOutput, cached, write := int64(0), int64(0), int64(0), int64(0)
	cost := 0.
	var price llmjudge.Prices
	for i, b := range runs {
		var r rerank.Run
		if json.Unmarshal(b, &r) != nil {
			return nil, nil, errors.New("invalid run")
		}
		if r.Model != model {
			return nil, nil, errors.New("run model mismatch")
		}
		if i == 0 {
			price = r.Prices
		}
		if price != r.Prices {
			return nil, nil, errors.New("multiple pricing snapshots")
		}
		totalInput += r.InputTokens
		totalOutput += r.OutputTokens
		cached += r.CachedInputTokens
		write += r.CacheWriteInputTokens
		cost += r.EstimatedCost
		unknown += r.UnknownUsageCalls
	}
	var cfg map[string]json.RawMessage
	if json.Unmarshal(config, &cfg) != nil {
		return nil, nil, errors.New("invalid stored config")
	}
	var provider string
	json.Unmarshal(cfg["Provider"], &provider)
	if provider != "anthropic" {
		return nil, nil, fmt.Errorf("unsupported recorded provider %s", provider)
	}
	m := map[string]any{"manifest_version": "evaluation-manifest-v1", "provider": provider, "exact_model_id": model, "endpoint": "https://api.anthropic.com/v1/messages", "endpoint_kind": "Anthropic Messages", "endpoint_evidence": "verified Day3 client code at evaluation_git_commit; endpoint not recorded in response", "api_revision": cfg["APIVersion"], "prompt_version": "llm-reranker-claude-v1", "rubric_version": "llm-reranker-claude-v1", "system_prompt": prompt, "system_prompt_sha256": hash(prompt), "request_parameters": first, "parameter_sha256": Fingerprint(first), "temperature": nil, "temperature_semantics": "omitted; provider default, not fixed/recorded numerically", "scoring_weights": map[string]any{"overall": "direct holistic LLM output; no weighted sum", "six_axes": "independent direct scores; selection tie-break only"}, "evaluation_code_version": "llm-reranker-claude-v1", "evaluation_git_commit": commit, "first_evaluated_at": times[0], "last_evaluated_at": times[len(times)-1], "pricing_snapshot": price, "input_tokens": totalInput, "output_tokens": totalOutput, "cached_tokens": cached, "cache_write_tokens": write, "calculated_estimated_cost_usd": cost, "unknown_usage_calls": unknown, "stored_config_sha256": hash(string(config)), "saved_response_hashes_sha256": Fingerprint(responses), "count": len(evidence)}
	m["fingerprint_sha256"] = Fingerprint(m)
	// Ensure no authorization/header credential was accidentally retained.
	encoded, _ := json.Marshal(first)
	if bytes.Contains(bytes.ToLower(encoded), []byte("x-api-key")) {
		return nil, nil, errors.New("credential parameter forbidden")
	}
	return m, inputs, nil
}
