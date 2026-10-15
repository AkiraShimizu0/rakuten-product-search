package observer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const CFDataset = "rumPageloadEventsAdaptiveGroups"

type typeRef struct {
	Name   string
	OfType *typeRef
}

func (t typeRef) named() string {
	if t.Name != "" {
		return t.Name
	}
	if t.OfType != nil {
		return t.OfType.named()
	}
	return ""
}

type schemaField struct {
	Name string
	Type typeRef
	Args []struct {
		Name string
		Type typeRef
	}
}
type schemaType struct {
	Name        string
	Fields      []schemaField
	InputFields []struct{ Name string }
}
type cfSchema struct {
	Data struct {
		Schema struct{ Types []schemaType } `json:"__schema"`
	}
	Errors []json.RawMessage
}
type CloudRow struct {
	Count *float64 `json:"count"`
	Sum   struct {
		Visits *float64 `json:"visits"`
	} `json:"sum"`
	Dimensions map[string]string `json:"dimensions,omitempty"`
}

func ParseCloudflare(b []byte) (map[string][]CloudRow, error) {
	var v struct {
		Data struct {
			Viewer struct {
				Accounts []map[string][]CloudRow `json:"accounts"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}
	if json.Unmarshal(b, &v) != nil || len(v.Errors) > 0 || len(v.Data.Viewer.Accounts) != 1 {
		return nil, errors.New("Cloudflare Web Analytics API unavailable or GraphQL error")
	}
	m := v.Data.Viewer.Accounts[0]
	for _, key := range []string{"total", "pages", "referrers"} {
		rows, ok := m[key]
		if !ok {
			return nil, errors.New("Cloudflare aggregation missing")
		}
		for _, r := range rows {
			if r.Count == nil || *r.Count < 0 {
				return nil, errors.New("invalid Cloudflare count")
			}
		}
	}
	return m, nil
}

// BuildCFQuery derives group/filter/dimension field names from live official introspection.
// It fails closed; HTTP/zone metrics are never substituted for RUM page loads.
func BuildCFQuery(b []byte, account, host string, w Window) (string, error) {
	var s cfSchema
	if json.Unmarshal(b, &s) != nil || len(s.Errors) > 0 {
		return "", errors.New("Cloudflare schema unavailable")
	}
	types := map[string]schemaType{}
	for _, t := range s.Data.Schema.Types {
		types[t.Name] = t
	}
	var group, filter string
	for _, t := range s.Data.Schema.Types {
		for _, f := range t.Fields {
			if f.Name == CFDataset {
				candidate := f.Type.named()
				for _, a := range f.Args {
					if a.Name == "filter" {
						filter = a.Type.named()
						group = candidate
					}
				}
			}
		}
	}
	field := func(t, name string) string {
		for _, f := range types[t].Fields {
			if f.Name == name {
				return f.Type.named()
			}
		}
		return ""
	}
	input := func(name string) bool {
		for _, f := range types[filter].InputFields {
			if f.Name == name {
				return true
			}
		}
		return false
	}
	dims := field(group, "dimensions")
	sum := field(group, "sum")
	if group == "" || field(group, "count") == "" || field(dims, "requestPath") == "" || field(dims, "refererHost") == "" || !input("requestHost") || !input("datetime_geq") || !input("datetime_lt") {
		return "", errors.New("CLOUDFLARE_WEB_ANALYTICS_API_UNAVAILABLE: required RUM schema not exposed")
	}
	q := func(v string) string { b, _ := json.Marshal(v); return string(b) }
	filt := fmt.Sprintf("{datetime_geq:%s,datetime_lt:%s,requestHost:%s}", q(w.From), q(w.To), q(host))
	metrics := "count"
	if field(sum, "visits") != "" {
		metrics += " sum { visits }"
	}
	return fmt.Sprintf("{viewer{accounts(filter:{accountTag:%s}){total:%s(limit:1,filter:%s){%s} pages:%s(limit:1000,filter:%s,orderBy:[count_DESC]){%s dimensions{requestPath}} referrers:%s(limit:1000,filter:%s,orderBy:[count_DESC]){%s dimensions{refererHost}}}}}", q(account), CFDataset, filt, metrics, CFDataset, filt, metrics, CFDataset, filt, metrics), nil
}
func (c *Client) Cloudflare(ctx context.Context, token, account string, w Window) Source {
	s := NewSource("https://api.cloudflare.com/client/v4/graphql", w)
	s.Notes = []string{"Dataset=" + CFDataset + " (official Web Analytics/RUM); adaptive sampling may apply.", "Visits is not unique people. Admin/QA cannot be separated: unclassified.", "Top pages/referrers capped at 1000; do not sum group visits into site unique visitors.", "Latest available date and reporting delay unknown unless separately established."}
	if token == "" || account == "" {
		s.Reason = "AUTHORIZATION_REQUIRED: dedicated Cloudflare Analytics Read token/account missing; deploy token not used"
		return s
	}
	introspection := "{__schema{types{name fields{name type{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name}}}}} args{name type{kind name ofType{kind name}}}} inputFields{name}}}}"
	b, e := c.Do(ctx, "POST", s.Endpoint, token, map[string]string{"query": introspection})
	if e != nil {
		s.Reason = e.Error()
		return s
	}
	query, e := BuildCFQuery(b, account, "choicelen.page", w)
	if e != nil {
		s.Status = "unavailable"
		s.Reason = e.Error()
		return s
	}
	b, e = c.Do(ctx, "POST", s.Endpoint, token, map[string]string{"query": query})
	if e != nil {
		s.Reason = e.Error()
		return s
	}
	data, e := ParseCloudflare(b)
	if e != nil {
		s.Status = "unavailable"
		s.Reason = "CLOUDFLARE_WEB_ANALYTICS_API_UNAVAILABLE: dataset query rejected; no Zone Analytics fallback"
		return s
	}
	s.Status = "known"
	s.Data = data
	s.Metrics = map[string]Metric{"page_views": Unknown(), "visits": Unknown(), "unique_visitors": Unknown()}
	if rows := data["total"]; len(rows) == 1 {
		s.Metrics["page_views"] = number(rows[0].Count)
		s.Metrics["visits"] = number(rows[0].Sum.Visits)
	} else {
		s.Notes = append(s.Notes, "No total row: no zero inferred.")
	}
	if !strings.Contains(query, "sum { visits }") {
		s.Notes = append(s.Notes, "visits field not exposed: unknown.")
	}
	return s
}
