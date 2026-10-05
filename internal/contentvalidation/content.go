package contentvalidation

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/market"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Row = market.Record

var Priority = []string{"compact-placement", "humidification-area", "filter-compatibility"}

type Claim struct {
	ID             string   `json:"claim_id"`
	Theme          string   `json:"theme_id"`
	Question       string   `json:"question_id"`
	Text           string   `json:"claim"`
	Classification string   `json:"classification"`
	Sources        []string `json:"source_ids"`
	Use            bool     `json:"article_use"`
	Direct         string   `json:"direct_or_inferred"`
	Notes          string   `json:"notes"`
}
type Calculation struct {
	ID      string    `json:"calculation_id"`
	Claim   string    `json:"claim_id"`
	Op      string    `json:"operation"`
	Inputs  []float64 `json:"inputs"`
	Units   []string  `json:"input_units"`
	Unit    string    `json:"unit"`
	Formula string    `json:"formula"`
	Sources []string  `json:"source_ids"`
	Caveat  string    `json:"caveat"`
	Result  float64   `json:"result"`
}
type Candidate struct {
	ID        string `json:"candidate_id"`
	Theme     string `json:"theme_id"`
	Prototype bool   `json:"is_prototype"`
	Summary   string `json:"summary"`
	URL       string `json:"url"`
	Review    string `json:"review_id"`
}
type Input struct {
	Protocol     Row           `json:"protocol"`
	Briefs       []Row         `json:"briefs"`
	Questions    []Row         `json:"questions"`
	Sources      []Row         `json:"sources"`
	Claims       []Claim       `json:"claims"`
	Products     []Row         `json:"products"`
	Calculations []Calculation `json:"calculations"`
	Candidates   []Candidate   `json:"candidates"`
	Editorial    []Row         `json:"editorial_scores"`
}
type Rating struct {
	ID         string `json:"review_id"`
	Decision   int    `json:"decision_usefulness"`
	Evidence   int    `json:"evidence_quality"`
	Comparison int    `json:"comparison_depth"`
	Action     int    `json:"actionability"`
	Clarity    int    `json:"clarity"`
	Overall    int    `json:"overall"`
	Notes      string `json:"notes"`
}
type Audit struct {
	Theme    string `json:"theme_id"`
	Claim    string `json:"claim_id"`
	Severity string `json:"severity"`
	Issue    string `json:"issue"`
	Reason   string `json:"reason"`
}

func str(v any) string {
	if v == nil {
		return "unknown"
	}
	return fmt.Sprint(v)
}
func read(path string, v any) ([]byte, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, v); e != nil {
		return nil, e
	}
	return b, nil
}
func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func jsonBytes(v any) []byte { b, _ := json.MarshalIndent(v, "", "  "); return b }
func rowsOf(v any) []Row     { b, _ := json.Marshal(v); var r []Row; _ = json.Unmarshal(b, &r); return r }
func allowedTheme(s string) bool {
	for _, v := range Priority {
		if s == v {
			return true
		}
	}
	return false
}
func Validate(in Input) error {
	sources := map[string]Row{}
	for _, s := range in.Sources {
		id := str(s["source_id"])
		if id == "unknown" || sources[id] != nil {
			return fmt.Errorf("duplicate/missing source %s", id)
		}
		if str(s["url"]) == "unknown" || str(s["accessed_at"]) == "unknown" {
			return fmt.Errorf("source provenance %s", id)
		}
		sources[id] = s
	}
	questions := map[string]string{}
	for _, q := range in.Questions {
		id := str(q["question_id"])
		theme := str(q["theme_id"])
		if questions[id] != "" || !allowedTheme(theme) {
			return fmt.Errorf("duplicate/invalid question %s", id)
		}
		questions[id] = theme
	}
	claims := map[string]bool{}
	for _, c := range in.Claims {
		if claims[c.ID] || c.ID == "" {
			return fmt.Errorf("duplicate claim ID %s", c.ID)
		}
		claims[c.ID] = true
		if questions[c.Question] != c.Theme || !allowedTheme(c.Theme) {
			return fmt.Errorf("orphan claim %s", c.ID)
		}
		if c.Classification != "Verified" && c.Classification != "Derived" && c.Classification != "Seller-only" && c.Classification != "Unknown" {
			return fmt.Errorf("invalid claim class")
		}
		if c.Use && (c.Classification == "Seller-only" || c.Classification == "Unknown") {
			return fmt.Errorf("unsupported article claim %s", c.ID)
		}
		if c.Use && len(c.Sources) == 0 {
			return fmt.Errorf("orphan source for %s", c.ID)
		}
		for _, id := range c.Sources {
			if sources[id] == nil {
				return fmt.Errorf("unknown source %s", id)
			}
		}
	}
	counts := map[string][2]int{}
	ids := map[string]bool{}
	for _, c := range in.Candidates {
		if !allowedTheme(c.Theme) || ids[c.ID] || c.Summary == "" {
			return fmt.Errorf("invalid candidate")
		}
		ids[c.ID] = true
		n := counts[c.Theme]
		if c.Prototype {
			n[0]++
		} else {
			n[1]++
		}
		counts[c.Theme] = n
	}
	for _, t := range Priority {
		if counts[t] != [2]int{1, 3} {
			return fmt.Errorf("%s needs 1+3 candidates", t)
		}
	}
	for _, c := range in.Calculations {
		if !claims[c.Claim] || len(c.Inputs) != 2 {
			return fmt.Errorf("invalid derived input")
		}
		for _, s := range c.Sources {
			if sources[s] == nil {
				return fmt.Errorf("unknown calculation source")
			}
		}
	}
	return nil
}
func Calculate(c Calculation) (float64, error) {
	if len(c.Inputs) != 2 {
		return 0, fmt.Errorf("need two inputs")
	}
	switch c.Op {
	case "divide":
		if c.Inputs[1] == 0 {
			return 0, fmt.Errorf("zero denominator")
		}
		return c.Inputs[0] / c.Inputs[1], nil
	case "subtract":
		return c.Inputs[0] - c.Inputs[1], nil
	}
	return 0, fmt.Errorf("unknown operation")
}
func writeBatch(dir string, files map[string][]byte) error {
	if e := os.MkdirAll(dir, 0755); e != nil {
		return e
	}
	names := []string{}
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if _, e := os.Stat(filepath.Join(dir, n)); e == nil {
			return fmt.Errorf("refuse existing output %s", n)
		} else if !os.IsNotExist(e) {
			return e
		}
	}
	for _, n := range names {
		f, e := os.OpenFile(filepath.Join(dir, n), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if e != nil {
			return e
		}
		_, we := f.Write(files[n])
		ce := f.Close()
		if we != nil {
			return we
		}
		if ce != nil {
			return ce
		}
	}
	return nil
}
func csvFiles(tables map[string][]Row) (map[string][]byte, error) {
	out := map[string][]byte{}
	for n, r := range tables {
		b, e := market.CSV(r, nil)
		if e != nil {
			return nil, e
		}
		out[n] = b
	}
	return out, nil
}
func Prepare(data, content string) error {
	var in Input
	ib, e := read(filepath.Join(data, "day5-input.json"), &in)
	if e != nil {
		return e
	}
	if e = Validate(in); e != nil {
		return e
	}
	for i := range in.Calculations {
		in.Calculations[i].Result, e = Calculate(in.Calculations[i])
		if e != nil {
			return e
		}
	}
	rng := rand.New(rand.NewSource(20261005))
	sort.Slice(in.Candidates, func(i, j int) bool { return in.Candidates[i].ID < in.Candidates[j].ID })
	blind := []Row{}
	key := []Row{}
	for i := range in.Candidates {
		c := &in.Candidates[i]
		c.Review = fmt.Sprintf("r-%016x", rng.Uint64())
		key = append(key, Row{"review_id": c.Review, "candidate_id": c.ID, "theme_id": c.Theme, "is_prototype": c.Prototype, "source_url": c.URL})
		blind = append(blind, Row{"review_id": c.Review, "topic": c.Theme, "content_summary": c.Summary})
	}
	rng.Shuffle(len(blind), func(i, j int) { blind[i], blind[j] = blind[j], blind[i] })
	matrix := []Row{}
	sourceMap := map[string]Row{}
	for _, s := range in.Sources {
		sourceMap[str(s["source_id"])] = s
	}
	qmap := map[string]Row{}
	for _, q := range in.Questions {
		qmap[str(q["question_id"])] = q
	}
	for _, c := range in.Claims {
		ids := c.Sources
		if len(ids) == 0 {
			ids = []string{"unknown"}
		}
		for _, id := range ids {
			s := sourceMap[id]
			matrix = append(matrix, Row{"claim_id": c.ID, "theme": c.Theme, "decision_question": qmap[c.Question]["decision_question"], "claim": c.Text, "source": id, "source_type": s["publisher_type"], "source_url": s["url"], "accessed_at": s["accessed_at"], "evidence_strength": c.Classification, "direct_or_inferred": c.Direct, "notes": c.Notes, "article_use": c.Use})
		}
	}
	files, e := csvFiles(map[string][]Row{"day5-content-briefs.csv": in.Briefs, "day5-decision-questions.csv": in.Questions, "day5-evidence-matrix.csv": matrix, "day5-claims.csv": rowsOf(in.Claims), "day5-products.csv": in.Products, "day5-derived-data.csv": rowsOf(in.Calculations), "day5-competitor-scores.csv": in.Editorial, "day5-blind-review.csv": blind, "day5-review-key.csv": key, "day5-sources.csv": in.Sources})
	if e != nil {
		return e
	}
	hashes := map[string]string{}
	for _, n := range []string{"day5-compact-placement.md", "day5-humidification.md", "day5-filter-compatibility.md"} {
		b, e := os.ReadFile(filepath.Join(content, n))
		if e != nil {
			return e
		}
		hashes[n] = digest(b)
	}
	files["day5-prepared.json"] = jsonBytes(in)
	files["day5-blind-manifest.json"] = jsonBytes(Row{"input_hash": digest(ib), "prototype_hashes": hashes, "blind_sha256": digest(files["day5-blind-review.csv"]), "seed": 20261005, "n": 12, "representation": "anonymized_structured_summaries_not_full_articles", "protocol": in.Protocol})
	return writeBatch(data, files)
}
func ValidateRatings(r []Rating, expected map[string]bool) error {
	if len(r) != len(expected) {
		return fmt.Errorf("missing judge row")
	}
	seen := map[string]bool{}
	for _, x := range r {
		if !expected[x.ID] || seen[x.ID] {
			return fmt.Errorf("blind ID join/duplicate")
		}
		seen[x.ID] = true
		if strings.TrimSpace(x.Notes) == "" {
			return fmt.Errorf("missing judge notes")
		}
		for _, n := range []int{x.Decision, x.Evidence, x.Comparison, x.Action, x.Clarity, x.Overall} {
			if n < 1 || n > 5 {
				return fmt.Errorf("invalid score")
			}
		}
	}
	return nil
}
func Decision(p, b, e float64, wins, major int) string {
	d := p - b
	if major > 0 || p < 3.7 || d <= -0.50 {
		return "NO-GO"
	}
	if p >= 4.3 && d >= 0.30-1e-9 && wins == 3 && e >= 4.3 {
		return "Strong GO"
	}
	if p >= 4 && d >= -1e-9 && wins >= 2 {
		return "GO"
	}
	return "HOLD"
}
func Analyze(data string) error {
	var blindFile []byte
	var err error
	blindFile, err = os.ReadFile(filepath.Join(data, "day5-blind-review.csv"))
	if err != nil {
		return err
	}
	br, e := csv.NewReader(strings.NewReader(string(blindFile))).ReadAll()
	if e != nil {
		return e
	}
	if len(br) != 13 {
		return fmt.Errorf("need 12 blind candidates")
	}
	expected := map[string]bool{}
	idcol := -1
	for i, h := range br[0] {
		if h == "review_id" {
			idcol = i
		}
	}
	if idcol < 0 {
		return fmt.Errorf("blind ID header missing")
	}
	for _, r := range br[1:] {
		if r[idcol] == "" || expected[r[idcol]] {
			return fmt.Errorf("missing/duplicate blind ID")
		}
		expected[r[idcol]] = true
	}
	var judges [3][]Rating
	for i := 0; i < 3; i++ {
		_, e = read(filepath.Join(data, fmt.Sprintf("day5-judge-%d.json", i+1)), &judges[i])
		if e != nil {
			return e
		}
		if e = ValidateRatings(judges[i], expected); e != nil {
			return e
		}
	}
	// Key/source roles are read only after every judge passes completeness and range checks.
	var in Input
	_, e = read(filepath.Join(data, "day5-prepared.json"), &in)
	if e != nil {
		return e
	}
	if e = Validate(in); e != nil {
		return e
	}
	var manifest struct {
		InputHash       string            `json:"input_hash"`
		BlindHash       string            `json:"blind_sha256"`
		PrototypeHashes map[string]string `json:"prototype_hashes"`
	}
	if _, e = read(filepath.Join(data, "day5-blind-manifest.json"), &manifest); e != nil {
		return e
	}
	if digest(blindFile) != manifest.BlindHash {
		return fmt.Errorf("blind file changed after preparation")
	}
	inputBytes, e := os.ReadFile(filepath.Join(data, "day5-input.json"))
	if e != nil {
		return e
	}
	if digest(inputBytes) != manifest.InputHash {
		return fmt.Errorf("evidence input changed after preparation")
	}
	for n, h := range manifest.PrototypeHashes {
		b, e := os.ReadFile(filepath.Join(filepath.Dir(data), "content", n))
		if e != nil {
			return e
		}
		if digest(b) != h {
			return fmt.Errorf("prototype changed after preparation")
		}
	}
	keys := map[string]Candidate{}
	for _, c := range in.Candidates {
		if !expected[c.Review] || keys[c.Review].ID != "" {
			return fmt.Errorf("blind key join")
		}
		keys[c.Review] = c
	}
	if len(keys) != len(expected) {
		return fmt.Errorf("blind key count")
	}
	keyBytes, e := os.ReadFile(filepath.Join(data, "day5-review-key.csv"))
	if e != nil {
		return e
	}
	kr, e := csv.NewReader(strings.NewReader(string(keyBytes))).ReadAll()
	if e != nil {
		return e
	}
	if len(kr) != 13 {
		return fmt.Errorf("blind key row count")
	}
	ki := map[string]int{}
	for i, h := range kr[0] {
		ki[h] = i
	}
	for _, h := range []string{"review_id", "candidate_id", "theme_id", "is_prototype"} {
		if _, ok := ki[h]; !ok {
			return fmt.Errorf("key header missing")
		}
	}
	keySeen := map[string]bool{}
	for _, row := range kr[1:] {
		id := row[ki["review_id"]]
		c, ok := keys[id]
		if !ok || keySeen[id] || c.ID != row[ki["candidate_id"]] || c.Theme != row[ki["theme_id"]] || fmt.Sprint(c.Prototype) != row[ki["is_prototype"]] {
			return fmt.Errorf("blind key join mismatch")
		}
		keySeen[id] = true
	}
	var audits []Audit
	_, e = read(filepath.Join(data, "day5-audit.json"), &audits)
	if e != nil {
		return e
	}
	audited := map[string]bool{}
	claimIDs := map[string]string{}
	claimAuditSeen := map[string]bool{}
	for _, c := range in.Claims {
		claimIDs[c.ID] = c.Theme
	}
	for _, a := range audits {
		if !allowedTheme(a.Theme) || claimIDs[a.Claim] != a.Theme {
			return fmt.Errorf("invalid audit claim")
		}
		if a.Severity != "none" && a.Severity != "minor" && a.Severity != "major" && a.Severity != "critical" {
			return fmt.Errorf("invalid audit severity")
		}
		audited[a.Theme] = true
		claimAuditSeen[a.Claim] = true
	}
	for id := range claimIDs {
		if !claimAuditSeen[id] {
			return fmt.Errorf("missing claim audit %s", id)
		}
	}
	for _, t := range Priority {
		if !audited[t] {
			return fmt.Errorf("missing theme audit")
		}
	}
	avg := map[string][6]float64{}
	perJudge := map[string][3]Rating{}
	for j, rs := range judges {
		for _, x := range rs {
			v := avg[x.ID]
			for k, s := range []int{x.Decision, x.Evidence, x.Comparison, x.Action, x.Clarity, x.Overall} {
				v[k] += float64(s) / 3
			}
			avg[x.ID] = v
			pj := perJudge[x.ID]
			pj[j] = x
			perJudge[x.ID] = pj
		}
	}
	results := []Row{}
	candidateRows := []Row{}
	for _, c := range in.Candidates {
		v := avg[c.Review]
		candidateRows = append(candidateRows, Row{"candidate_id": c.ID, "theme_id": c.Theme, "is_prototype": c.Prototype, "decision_usefulness": v[0], "evidence_quality": v[1], "comparison_depth": v[2], "actionability": v[3], "clarity": v[4], "overall": v[5]})
	}
	for priority, t := range Priority {
		var p Candidate
		var comps []Candidate
		for _, c := range in.Candidates {
			if c.Theme == t {
				if c.Prototype {
					p = c
				} else {
					comps = append(comps, c)
				}
			}
		}
		sort.Slice(comps, func(i, j int) bool {
			a, b := avg[comps[i].Review][5], avg[comps[j].Review][5]
			if a != b {
				return a > b
			}
			return comps[i].ID < comps[j].ID
		})
		best := comps[0]
		a, b := avg[p.Review], avg[best.Review]
		wins := 0
		diffs := [3]float64{}
		major, minor := 0, 0
		for _, x := range audits {
			if x.Theme == t {
				if x.Severity == "major" || x.Severity == "critical" {
					major++
				}
				if x.Severity == "minor" {
					minor++
				}
			}
		}
		for j := 0; j < 3; j++ {
			mx := 0
			for _, c := range comps {
				if s := perJudge[c.Review][j].Overall; s > mx {
					mx = s
				}
			}
			diffs[j] = float64(perJudge[p.Review][j].Overall - mx)
			if diffs[j] >= 0 {
				wins++
			}
		}
		results = append(results, Row{"priority": priority + 1, "theme_id": t, "prototype_overall": a[5], "best_competitor_id": best.ID, "best_competitor_overall": b[5], "difference": a[5] - b[5], "judge_differences_to_individual_best": diffs, "judge_noninferior_count": wins, "decision_usefulness": a[0], "evidence_quality": a[1], "comparison_depth": a[2], "actionability": a[3], "clarity": a[4], "audit_major_critical": major, "audit_minor": minor, "decision": Decision(a[5], b[5], a[1], wins, major)})
	}
	files, e := csvFiles(map[string][]Row{"day5-results.csv": results, "day5-candidate-scores.csv": candidateRows, "day5-claim-audit.csv": rowsOf(audits)})
	if e != nil {
		return e
	}
	files["day5-comparison.json"] = jsonBytes(Row{"results": results, "candidate_scores": candidateRows, "n_candidates": 12, "judges": 3, "human_ground_truth": false, "full_page_superiority_tested": false, "decision_basis": "structured summary blind comparison plus independent claim audit"})
	return writeBatch(data, files)
}
