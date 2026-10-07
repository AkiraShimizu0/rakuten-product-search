package foundation

import (
	"bytes"
	"encoding/json"
	"errors"
	"jev-money-engine/internal/diversify"
	"jev-money-engine/internal/evaluate"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/research"
)

func canonical(v any) []byte {
	b, _ := json.Marshal(v)
	var x any
	_ = json.Unmarshal(b, &x)
	b, _ = json.Marshal(x)
	return b
}
func Frozen(path string) error {
	var old struct {
		ClaudeModel string          `json:"claude_model"`
		Rubric      string          `json:"claude_rubric_sha256"`
		Cluster     string          `json:"cluster_code_sha256"`
		Jev         json.RawMessage `json:"jev"`
		Questions   json.RawMessage `json:"questions"`
	}
	if e := Load(path, &old); e != nil {
		return e
	}
	if old.ClaudeModel != llmjudge.Model || old.Rubric != research.Hash([]byte(llmjudge.Rubric)) || old.Cluster != diversify.SourceFingerprint() || !bytes.Equal(canonical(old.Jev), canonical(evaluate.DefaultConfig())) || !bytes.Equal(canonical(old.Questions), canonical(jev.Questions())) {
		return errors.New("semantic/family configuration differs from frozen baseline")
	}
	return nil
}
