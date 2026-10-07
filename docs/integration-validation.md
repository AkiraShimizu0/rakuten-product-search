# Integration validation

Merge commit 67aaa3ef2af05166bdc6c8f2febc246aa40400ef combines PR #15 and PR #11. No conflicts. Every PR #1–#15 remote tip was checked with git merge-base --is-ancestor and is included.

## Regression checks
- Full site tree equals origin/choicelen-home-brand-dedup: top HTML, article HTML, CSS, robots/sitemap, affiliate URL, canonical and claim JSON preserved.
- content/published and wrangler.jsonc equal latest UI source.
- internal/jev, internal/gate, internal/rerank, internal/diversify and cmd/build-choiceLen equal latest UI source.
- internal/radar, internal/radarhistory, internal/foundation and Radar workflow equal R&D tip.
- Radar DefaultRule remains 10 percent AND 1000 JPY; cap remains 2.
- Experimental resolver v2 is isolated under foundation; production diversify remains unchanged.
- Full go test ./... and go vet ./... passed. Collector/Rakuten mocked fetch and storage pipeline, family fixtures, event thresholds/history/failure safety, frozen article/source integrity are covered without live paid API calls.
- No deploy: integration branch is absent from both workflows' push branch allowlists; no manual workflow dispatched.

## Repository audit
No newly tracked SQLite/env/executable/credential/response dump/history/screenshot. Existing official publication manifests and fixtures remain. Old editorial/site.css and inactive R2/diagnostic adapters were detected but retained for provenance; no risky deletion/refactor.
Some CLI defaults refer to private output paths outside the checkout. Fresh checkout cannot reconstruct historical experiments without those inputs. Scheduled binary remains pinned, not updated from integration.
Site tree SHA: 339f56ffbed6d5daa0eb6d1ba67656b140e10cb9
