package observer

import (
	"encoding/json"
	"errors"
	"os"
)

func Reuse(paths []string, requested Window) (map[string]Source, error) {
	sources := map[string]Source{}
	for _, path := range paths {
		b, e := os.ReadFile(path)
		if e != nil {
			return nil, errors.New("reused observation unavailable")
		}
		var o Observation
		if json.Unmarshal(b, &o) != nil || o.Protocol != "day7-v1" || o.Requested != requested || o.Sources == nil {
			return nil, errors.New("reused observation protocol/window mismatch")
		}
		for k, s := range o.Sources {
			raw, e := json.Marshal(s.Data)
			if e != nil {
				return nil, errors.New("invalid reused source")
			}
			if s.Status == "known" {
				switch k {
				case "url_inspection":
					var v Inspection
					if json.Unmarshal(raw, &v) != nil || v.Index.Verdict == "" {
						return nil, errors.New("invalid reused inspection")
					}
					s.Data = v
				case "sitemap":
					var v Sitemap
					if json.Unmarshal(raw, &v) != nil || v.Path == "" {
						return nil, errors.New("invalid reused sitemap")
					}
					for name, m := range s.Metrics {
						if m.Status == "known" || m.Status == "zero" {
							if n, ok := m.Value.(float64); ok {
								v64 := int64(n)
								if float64(v64) != n || n < 0 {
									return nil, errors.New("invalid reused sitemap count")
								}
								if name == "errors" {
									v.Errors = &v64
								}
								if name == "warnings" {
									v.Warnings = &v64
								}
							}
						}
					}
					s.Data = v
				case "radar":
					var m map[string]json.RawMessage
					if json.Unmarshal(raw, &m) != nil {
						return nil, errors.New("invalid reused Radar")
					}
					var fixed RadarSummary
					if json.Unmarshal(m["fixed_window"], &fixed) != nil {
						return nil, errors.New("invalid reused Radar window")
					}
					var current any
					_ = json.Unmarshal(m["current_manifest"], &current)
					s.Data = map[string]any{"fixed_window": fixed, "current_manifest": current}
				}
			}
			s.Notes = append(s.Notes, "Reused saved observation; original checked_at/window retained; no API refetch.")
			sources[k] = s
		}
	}
	return sources, nil
}
