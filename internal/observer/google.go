package observer

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

type SearchRow struct {
	Keys        []string `json:"keys,omitempty"`
	Clicks      *float64 `json:"clicks"`
	Impressions *float64 `json:"impressions"`
	CTR         *float64 `json:"ctr"`
	Position    *float64 `json:"position"`
}
type Search struct {
	Rows        []SearchRow     `json:"rows"`
	Aggregation string          `json:"responseAggregationType"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

func ParseSearch(b []byte) (Search, error) {
	var v Search
	var raw map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || raw == nil || raw["error"] != nil {
		return v, errors.New("invalid Search Console response")
	}
	if e := json.Unmarshal(b, &v); e != nil {
		return v, errors.New("invalid Search Console rows")
	}
	for _, r := range v.Rows {
		if r.Clicks == nil || r.Impressions == nil || r.CTR == nil || r.Position == nil || *r.Clicks < 0 || *r.Impressions < 0 || *r.CTR < 0 || *r.CTR > 1 || *r.Position < 0 {
			return v, errors.New("invalid Search Console metrics")
		}
	}
	return v, nil
}

type Inspection struct {
	Index IndexStatus `json:"indexStatusResult"`
}
type IndexStatus struct {
	Verdict         string   `json:"verdict"`
	Coverage        string   `json:"coverageState"`
	Robots          string   `json:"robotsTxtState"`
	Indexing        string   `json:"indexingState"`
	LastCrawl       string   `json:"lastCrawlTime"`
	Fetch           string   `json:"pageFetchState"`
	UserCanonical   string   `json:"userCanonical"`
	GoogleCanonical string   `json:"googleCanonical"`
	Sitemaps        []string `json:"sitemap"`
	Referring       []string `json:"referringUrls"`
}

func ParseInspection(b []byte) (Inspection, error) {
	var v struct {
		Result *Inspection `json:"inspectionResult"`
	}
	if json.Unmarshal(b, &v) != nil || v.Result == nil || v.Result.Index.Verdict == "" {
		return Inspection{}, errors.New("invalid URL Inspection response")
	}
	return *v.Result, nil
}

type Sitemap struct {
	Path           string           `json:"path"`
	LastSubmitted  string           `json:"lastSubmitted"`
	LastDownloaded string           `json:"lastDownloaded"`
	Pending        *bool            `json:"isPending"`
	Errors         *int64           `json:"-"`
	Warnings       *int64           `json:"-"`
	Contents       []SitemapContent `json:"contents"`
}
type SitemapContent struct {
	Type      string      `json:"type"`
	Submitted json.Number `json:"submitted"`
}

func ParseSitemap(b []byte) (Sitemap, error) {
	var v Sitemap
	var r struct {
		Errors   *json.Number `json:"errors"`
		Warnings *json.Number `json:"warnings"`
	}
	if json.Unmarshal(b, &v) != nil || json.Unmarshal(b, &r) != nil || v.Path == "" {
		return v, errors.New("invalid sitemap response")
	}
	parse := func(n *json.Number) (*int64, error) {
		if n == nil {
			return nil, nil
		}
		x, e := strconv.ParseInt(string(*n), 10, 64)
		if e != nil || x < 0 {
			return nil, errors.New("invalid sitemap count")
		}
		return &x, nil
	}
	var e error
	v.Errors, e = parse(r.Errors)
	if e != nil {
		return v, e
	}
	v.Warnings, e = parse(r.Warnings)
	return v, e
}
func number(p *float64) Metric {
	if p == nil {
		return Unknown()
	}
	return Known(*p)
}
func (c *Client) Google(ctx context.Context, token, property, article, from, to string) map[string]Source {
	w := Window{from, to, "America/Los_Angeles", "daily"}
	sources := map[string]Source{}
	base := "https://www.googleapis.com/webmasters/v3/sites/" + url.PathEscape(property)
	s := NewSource(base+"/searchAnalytics/query", w)
	s.Notes = []string{"PT day boundaries differ from the requested JST window; query rows can be privacy-suppressed.", "dataState=all; metadata.first_incomplete_date retained; omitted daily rows are not filled with zero.", "API web may differ from manual Web (text-based); do not force equality."}
	data := map[string]Search{}
	failed := false
	for _, dim := range []string{"total", "date", "query", "page"} {
		rows := []SearchRow{}
		var agg Search
		for start := 0; start < 100000; start += 25000 {
			req := map[string]any{"startDate": from, "endDate": to, "type": "web", "dataState": "all", "rowLimit": 25000, "startRow": start, "aggregationType": "auto", "dimensionFilterGroups": []any{map[string]any{"groupType": "and", "filters": []any{map[string]string{"dimension": "page", "operator": "equals", "expression": article}}}}}
			if dim != "total" {
				req["dimensions"] = []string{dim}
			}
			b, e := c.Do(ctx, "POST", s.Endpoint, token, req)
			if e != nil {
				s.Reason = e.Error()
				failed = true
				break
			}
			v, e := ParseSearch(b)
			if e != nil {
				s.Reason = e.Error()
				failed = true
				break
			}
			agg = v
			rows = append(rows, v.Rows...)
			if len(v.Rows) < 25000 {
				break
			}
			if start == 75000 {
				s.Notes = append(s.Notes, "Row cap reached: results truncated.")
			}
		}
		if failed {
			break
		}
		agg.Rows = rows
		data[dim] = agg
	}
	if !failed {
		s.Status = "known"
		s.Data = data
		s.Metrics = map[string]Metric{"clicks": Unknown(), "impressions": Unknown(), "ctr": Unknown(), "average_position": Unknown()}
		if r := data["total"].Rows; len(r) == 1 {
			s.Metrics = map[string]Metric{"clicks": number(r[0].Clicks), "impressions": number(r[0].Impressions), "ctr": number(r[0].CTR), "average_position": number(r[0].Position)}
		} else {
			s.Notes = append(s.Notes, "No aggregate row: metrics remain unknown, not fabricated zero.")
		}
		for _, v := range data {
			var meta struct {
				Incomplete string `json:"first_incomplete_date"`
			}
			_ = json.Unmarshal(v.Metadata, &meta)
			if meta.Incomplete != "" {
				s.Delay = "incomplete from " + meta.Incomplete
				s.Status = "delayed"
			}
		}
	}
	sources["search_console"] = s
	current := Window{"current", "current", "UTC", "current indexed version; not historical"}
	ins := NewSource("https://searchconsole.googleapis.com/v1/urlInspection/index:inspect", current)
	b, e := c.Do(ctx, "POST", ins.Endpoint, token, map[string]string{"inspectionUrl": article, "siteUrl": property, "languageCode": "ja-JP"})
	if e != nil {
		ins.Reason = e.Error()
	} else {
		v, e := ParseInspection(b)
		if e != nil {
			ins.Reason = e.Error()
		} else {
			ins.Status = "known"
			ins.Data = v
		}
	}
	sources["url_inspection"] = ins
	sm := NewSource(base+"/sitemaps/"+url.PathEscape("https://choicelen.page/sitemap.xml"), current)
	b, e = c.Do(ctx, "GET", sm.Endpoint, token, nil)
	if e != nil {
		sm.Reason = e.Error()
	} else {
		v, e := ParseSitemap(b)
		if e != nil {
			sm.Reason = e.Error()
		} else {
			sm.Status = "known"
			sm.Data = v
			sm.Metrics = map[string]Metric{"errors": Unknown(), "warnings": Unknown()}
			if v.Errors != nil {
				sm.Metrics["errors"] = Known(float64(*v.Errors))
			}
			if v.Warnings != nil {
				sm.Metrics["warnings"] = Known(float64(*v.Warnings))
			}
			sm.Notes = []string{"contents.submitted counts submitted URLs; deprecated contents.indexed is deliberately ignored; not discovered/indexed proof."}
		}
	}
	sources["sitemap"] = sm
	return sources
}
