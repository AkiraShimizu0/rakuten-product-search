# Sampling bias validation — 2026-10-07 JST

All three categories are **ROBUST under the frozen exploratory protocol**. Exploration priority remains 除湿機 → 電動シュレッダー → 電気圧力鍋. Existing Track A GO remains unchanged; this is not market, SEO or revenue validation.

The [protocol](sampling-validation-protocol.md) was saved before new collection. Seed 20261007; available listings, price JPY 3,000–50,000, review presence; six pages under each of standard, ascending price and descending price. Eligibility remains >=5 reviews and >=80 description runes. Three rank-price bands, 30 fixed-hash products/band/category, 270 unique sampled listings total. No sample quota adjustment was made after seeing data.

| Category | raw fetched | unique pool | eligible pool | eligible rate | sample |
|---|---:|---:|---:|---:|---:|
| 除湿機 | 540 | 410 | 169 | 41.22% | 90 |
| 電動シュレッダー | 540 | 412 | 159 | 38.59% | 90 |
| 電気圧力鍋 | 540 | 304 | 111 | 36.51% | 90 |

Sample eligibility is 100% by construction, not pool eligibility. Review-order pool eligibility was 90.00%, 91.11%, 93.33%, respectively. The new acquisition has much lower eligible density. Pool duplicate ratios are 24.07%, 23.70%, 43.70%; all selected source/source_id pairs are unique within category.

| Category | Jev gate review → price | raw mean review → price | raw median review → price | Claude review → price | delta | rank |
|---|---|---|---|---|---:|---|
| 除湿機 | 100% → 97.78% | .988218 → .987520 | .998050 → .993300 | 64.35 → 65.00 | +.65 | 1 → 1 |
| 電動シュレッダー | 98% → 95.56% | .972560 → .960909 | .975150 → .965400 | 62.75 → 60.20 | -2.55 | 2 → 2 |
| 電気圧力鍋 | 100% → 100% | .974848 → .978609 | .975650 → .977400 | 60.90 → 59.70 | -1.20 | 3 → 3 |

Old Jev n=50/category, new n=45/category. Old Claude n=20/category, new n=10/category (3 low/3 mid/4 high). Gate shortages were not backfilled. Models, prompts, rubric, thresholds and diversification are fingerprint-checked and unchanged.

| Category | buyer problem review → price | comparison review → price | wrong-choice review → price | families scored review → price | largest share | HHI |
|---|---|---|---|---|---|---|
| 除湿機 | 70.25 → 71.40 | 64.85 → 65.90 | 68.90 → 69.60 | 19/20 → 10/10 | .10 → .10 | .055 → .100 |
| 電動シュレッダー | 70.20 → 68.80 | 67.70 → 65.20 | 62.10 → 60.90 | 17/20 → 8/10 | .10 → .30 | .065 → .160 |
| 電気圧力鍋 | 64.60 → 62.50 | 62.50 → 63.40 | 59.20 → 58.20 | 16/20 → 8/10 | .10 → .20 | .070 → .140 |

These breadth values use the unchanged discovery proxy. Different n makes raw family counts/HHI imperfect comparisons; family validation also finds baseline identification errors. Do not interpret those counts as established product diversity.

## Standard-order failure analysis

The preserved original standard sample covers 16×90=1,440 unique listings. Only three pass the unchanged filter (one each in fan, display and cooler categories). Every category fails minimum eligible breadth. 89–90 listings/category fail review-count eligibility; title/accessory, missing description and low-price counts overlap and are reported separately. Low reviews are the dominant observable bottleneck. An empty eligible sample does not prove the category itself lacks useful products, nor does it establish accessory contamination as the cause.

## Required questions

1. Dehumidifiers remain the strongest internal candidate: new Claude mean 65.00, rank 1.
2. Shredders/pressure cookers retain ranks 2/3; mean decreases are 2.55/1.20 points.
3. Review ordering does not uniformly increase Claude scores: dehumidifiers score slightly higher under price stratification.
4. Review ordering strongly increases eligible acquisition density, mainly by satisfying the review threshold. This experiment cannot isolate accessory-quality cleanup from popularity/model/semantic-selection effects.
5. The three EXPLORE choices have exploratory replication sufficient to prioritize later market validation, subject to family/selection caveats. No market research was performed here.

## Cost and limitations

Preflight estimate: **$1.672052**, below $2. New Jev: 109 successes, 0 failures, input 392,269/output 26,932 tokens, estimated $0.016475298; 26 reused, 135 scored total. New Claude: 29 successes, 0 failures, input 84,173/output 7,114, estimated $0.239486; 1 reused, 30 scored total. Total additional estimate **$0.255961298**. No new paid service was subscribed.

This is a bounded eligible search pool, still conditioned on reviews and price, not an unbiased census. Old semantic selection used score ordering while the new Claude subset uses a hash; n is small and no causal attribution/significance claim is made. ROBUST thresholds (>5-point mean shift, >.20 gate-rate shift, rank movement >=2; incomplete/<8 Claude yields INCONCLUSIVE) are exploratory and unchanged. Missing billing remains unknown and prevents further paid work.

Private outputs include pool/sample, eval checkpoint, acquisition comparison, standard-order reasons and cost ledger. `cmd/validate-foundation` supports prepare/collect/evaluate/report/family; use a new output directory for a new experiment. `-dry-run` estimates only pending paid evaluations; after completion it correctly reports none pending, not the original preflight cost.
