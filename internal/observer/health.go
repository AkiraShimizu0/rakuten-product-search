package observer

import (
	"encoding/json"
	"errors"
	"os"
)

func ProductionHealth(path string) Source {
	s := NewSource("previous read-only production verification", Window{"previous check", "previous check", "UTC", "snapshot"})
	s.Method = "existing verification JSON"
	if path == "" {
		s.Reason = "prior verification not supplied"
		return s
	}
	b, e := os.ReadFile(path)
	if e != nil {
		s.Reason = "health evidence unavailable"
		return s
	}
	var rows []struct {
		URL       string `json:"url"`
		Status    int    `json:"http_status"`
		TLS       bool   `json:"https_verified"`
		Canonical string `json:"canonical"`
		Noindex   bool   `json:"noindex"`
		Checked   string `json:"checked_at"`
		Error     string `json:"error"`
	}
	if json.Unmarshal(b, &rows) != nil {
		s.Reason = "invalid health evidence"
		return s
	}
	required := map[string]bool{"https://choicelen.page/": false, Article: false, "https://choicelen.page/robots.txt": false, "https://choicelen.page/sitemap.xml": false}
	for _, r := range rows {
		if _, ok := required[r.URL]; ok {
			if r.Status != 200 || !r.TLS || r.Noindex || r.Error != "" {
				s.Reason = "production verification failure"
				return s
			}
			if (r.URL == Article || r.URL == "https://choicelen.page/") && r.Canonical != r.URL {
				s.Reason = "canonical verification failure"
				return s
			}
			required[r.URL] = true
			if s.Window.From == "previous check" || r.Checked < s.Window.From {
				s.Window.From = r.Checked
			}
			if s.Window.To == "previous check" || r.Checked > s.Window.To {
				s.Window.To = r.Checked
			}
		}
	}
	for _, v := range required {
		if !v {
			s.Reason = errors.New("health evidence incomplete").Error()
			return s
		}
	}
	s.Status = "known"
	s.Notes = []string{"Reused prior technical checks; no fresh production request or historical D+7 indexing inference."}
	return s
}
