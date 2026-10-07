# Parallel category discovery and Price Radar v1

This experiment is isolated from ChoiceLen production. It does not publish,
edit article content, change Jev/Gate/Claude/clustering, or install schedules.
The branch is stacked on `day6-1-production-publication` (PR #9).

## Category discovery

Use the current Rakuten Genre Search API to confirm category identities before
creating a UTF-8 `genre_id,name` CSV containing 10–20 categories. Keep exclusion
rules and that category CSV fixed. The experiment covers appliances, PC equipment,
furniture, tools and outdoor equipment; excluded policy areas are not sampled.

```powershell
go run ./cmd/discover-categories -stage genres -genre-ids 0 -env-file ../rakuten.env -out data/category-discovery/genres
go run ./cmd/discover-categories -stage collect -categories categories.csv -data data/category-discovery/state -out data/category-discovery/a1 -sort=-reviewCount -dry-run
go run ./cmd/discover-categories -stage collect -categories categories.csv -data data/category-discovery/state -out data/category-discovery/a1 -sort=-reviewCount -env-file ../rakuten.env
go run ./cmd/discover-categories -stage jev -data data/category-discovery/state -out data/category-discovery/a2 -cache-db path/to/frozen-day3-5.db -budget-usd 6 -dry-run
go run ./cmd/discover-categories -stage jev -data data/category-discovery/state -out data/category-discovery/a2 -cache-db path/to/frozen-day3-5.db -budget-usd 6 -env-file ../rakuten.env
go run ./cmd/discover-categories -stage claude -data data/category-discovery/state -out data/category-discovery/a3 -gate path/to/gate-v1.json -budget-usd 6 -dry-run
go run ./cmd/discover-categories -stage claude -data data/category-discovery/state -out data/category-discovery/a3 -gate path/to/gate-v1.json -budget-usd 6 -env-file ../rakuten.env
go run ./cmd/discover-categories -stage report -data data/category-discovery/state -out data/category-discovery/final -gate path/to/gate-v1.json
```

Every stage output directory must be new. State JSON files are checkpoints for
resuming unchanged cohorts; successful evaluations are reused. Jev cache reuse
requires identical v1 questions/model/state. Sample identity is stable hash order
with seed 20261004, not API result order. API pages are not a population-random
sample. Review-count sorting explicitly biases toward reviewed products.

A1 uses 90 listings/category (three pages). Fixed DROP: unique<50,
description coverage<80%, eligible<15 or rate<15%, family proxy count<10 or
largest share>50%, accessory title proxy>80%. The original Day1 deterministic
filter remains unchanged. Jev samples at most 50 eligible products/category.
Claude covers at most five categories, selected by gate pass rate, then mean
comparison value, then genre ID; at most 20 products/category with family cap2.

EXPLORE requires Claude n>=10, mean overall>=60, mean comparison>=55,
mean Jev commodity risk<0.6, gate pass rate>=40%, family count>=8, and largest
family share<=25%. Nonselected categories are HOLD, not failed categories.
No arbitrary composite category score is calculated. Manufacturer evidence
availability remains unverified until external validation.

The initial standard-sort experiment was retained when it mostly returned
unreviewed listings. A separate, common review-count sampling protocol was
registered for all 16 categories; its bias and both experiment results are kept.
The filter and semantic evaluation logic were not adjusted to rescue scores.

Costs are estimates, not bills: Jev input $0.042/million, Claude frozen snapshot
input $2/million, output $10/million, cache read $0.2/million, cache write
$2.5/million. Request bytes at one token/byte and Claude max output 4096 are
conservative planning approximations, not token counts. Paid calls stop on
unknown usage/billing attempts. Actual token usage and estimates are recorded.

Frozen Day3.5 family rules were designed for air purifiers. Brand recognition
and generic family counts outside that category are proxies; unknown brands
are not real zero-brand populations, and singletons can inflate breadth.
Do not infer category demand, SEO viability or revenue from these samples.

## Price Radar

```powershell
go run ./cmd/refresh-offers -catalog-db path/to/frozen-day3-5.db -gate path/to/gate-v1.json -db data/price-radar/radar.db -limit 75 -dry-run
go run ./cmd/refresh-offers -catalog-db path/to/frozen-day3-5.db -gate path/to/gate-v1.json -db data/price-radar/radar.db -limit 75 -env-file ../rakuten.env -out data/price-radar/run-1
```

Refresh again using a **new** `-out` directory and the same DB/cohort source.
Flags: `-source rakuten`, `-timeout 20s`, `-interval 1.2s` (minimum 1 second),
`-retries 2`, `-limit 1..100`. Initial selection: saved Jev Gate pass plus saved
Claude evaluation, Claude DESC/comparison DESC/source ID ASC, family cap2.
The existing evaluation DB is opened in SQLite `mode=ro`.

Append-only `product_offer_snapshots` and `product_offer_events` live in a
separate DB, with update/delete refusal triggers. Identical listing/content in
the same UTC-hour bucket is deduplicated. Unchanged observations in later hours
are retained. Out-of-order observations are rejected. Never mix shops or
backfill from old catalog prices. The first live observation is `NO_BASELINE`.

Only API itemPrice is used; title coupons/points are never prices. A drop of
at least 10% **and** 1000 JPY becomes `PRICE_DROP_CANDIDATE`, not proof of a deal.
Other observed events include increases, returns to a previously observed
price, explicit availability changes, and review-count jumps >=10. Historical
min/max and elapsed days use observed snapshots only.

Empty exact item-code results are `api_not_found` / `api_item_missing`; they
do not prove a 404 page, discontinued model or out-of-stock status. API explicit
availability is a purchasability signal, not a stock guarantee. Fetch errors
are never saved as missing products. Missing numerical fields remain unknown.

Priority is `drop_percent + saved_Claude_overall/10`, sorted deterministically;
the separate candidate output uses family cap2. All raw events remain in SQLite.
No new Jev/Claude calls, affiliate conversion estimates or automatic posts.
No GitHub scheduled workflow or new repository credentials are configured.

## Checks and current documentation

```powershell
go test ./...
go vet ./...
git diff --check
```

- Rakuten item API: https://webservice.rakuten.co.jp/documentation/ichiba-item-search
- Rakuten genre API: https://webservice.rakuten.co.jp/documentation/ichiba-genre-search
- Jev models/pricing: https://docs.typesafe.ai/models
- Claude prices: https://platform.claude.com/docs/en/about-claude/pricing

These were checked on 2026-10-06; rates and API behavior must be checked again
before future experiments. Experiments, DB files and local env secrets are ignored.
