# Family Resolver validation — 2026-10-07 JST

Decision: **NEEDS_MORE_VALIDATION**. Experimental v2 is a **HOLD** candidate; production remains unchanged. Reference labels are independent AI/document review, not human ground truth.

The frozen [rulebook](family-rulebook.md) predates metrics. Reused review-order listings, fixed seed 20261007, 40/category, 120 total. Labels were assigned from title, description, explicit manufacturer/model information, with official sources for boundary cases, independently of resolver output. 56 ambiguous listings were excluded, leaving 64: dehumidifiers 17, shredders 25, pressure cookers 22. Missing manufacturer, multiple offered models and uncertain accessory identity account for many exclusions. This high exclusion rate limits generalization.

| Category | n | Baseline precision | recall | F1 | false merge pairs | false split pairs | exact partition items |
|---|---:|---:|---:|---:|---:|---:|---:|
| 除湿機 | 17 | 1.0000 | .8000 | .8889 | 0 | 1 | 15 |
| 電動シュレッダー | 25 | .2432 | .5294 | .3333 | 28 | 8 | 12 |
| 電気圧力鍋 | 22 | 1.0000 | .9259 | .9615 | 0 | 2 | 19 |

Totals: 28 false-merge pairs, 11 false-split pairs, 46/64 exact partition items. Exact match compares partition membership, not ID spelling. Pair errors are not counts of defective products.

| Category | Candidate precision | recall | F1 | false merge pairs | false split pairs | exact items |
|---|---:|---:|---:|---:|---:|---:|
| 除湿機 | 1.0000 | .8000 | .8889 | 0 | 1 | 15 |
| 電動シュレッダー | 1.0000 | 1.0000 | 1.0000 | 0 | 0 | 25 |
| 電気圧力鍋 | 1.0000 | .9259 | .9615 | 0 | 2 | 19 |

Candidate totals: 0 false-merge pairs, 3 false-split pairs, 59/64 exact items. Macro F1 improves from approximately .7279 to .9501; machine-calculated macro precision/recall/F1 are in `data/family-validation/macro-metrics.csv` in private experiment outputs.

## Diagnosis and experimental changes

Baseline collapses full shredder model numbers into overly broad letter prefixes, merging P5GCX2, P2HT, HS4SC and P10GC-like distinct models. Candidate v2 prioritizes full basic models, normalizes Unicode/case/hyphens and explicit color suffixes, normalizes known manufacturer aliases, removes bundle/title noise and separates body from parts. Unknown or conflicting identities stay conservative singletons. No product IDs are hardcoded. No changes were made to `internal/diversify` production logic.

Residual splits include an IJDC-K80 bundle whose title does not identify its manufacturer and T-fal CY3511JP/CY3518JP color variants: numeric model variants cannot safely be stripped with a generic suffix rule. Independent official evidence grouped these variants in the reference, while the title-only candidate intentionally does not guess. These are real misses, not removed to improve the result.

Reference revision provenance is retained: first-pass labels had 66 confident and 54 ambiguous rows; two seller-SKU listings without identifiable manufacturer were moved to ambiguous, and official color/model references corrected boundary labels before final metrics. Rulebook did not change. Both label files are retained.

Official references checked 2026-10-07 JST:

- https://www.irisohyama.co.jp/products/support/4967576484350 — IJDC-K80 identity.
- https://www.t-fal.co.jp/consumer-services/instructions-for-use/ — shared CY3511JP/CY3518JP manual index.
- https://www.t-fal.co.jp/pressurecookers/epc/products/lakulacooker-compact-7211004725/ — product/color variation.

The linked large PDF was not fully retrieved; no claim relies on reading that PDF. Official product and manual-index identity evidence was used.

## Reproduction and limits

From repository root:

```powershell
go run ./cmd/validate-foundation -stage family -out ../../outputs/foundation-validation
```

Outputs: gold-set, baseline/v2 assignments and metrics, error-cases CSVs. Rulebook/labels stay versioned privately. Candidate improvements were developed and measured on the same reference subset; there is no independent held-out family test. Before adoption, label additional difficult and ambiguous listings independently, include verified duplicate pairs, and test held-out models/manufacturers. Family breadth reported by the unchanged discovery resolver remains a proxy, especially for shredders.
