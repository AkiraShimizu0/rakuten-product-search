# PR consolidation plan

統合方式: 最新UI PR #15を基点にR&D PR #11をmerge commitで統合。共通祖先PR #9。旧PRはmerge/closeしない。下記close可否はintegrationがmainへmergeされ、必要fixture・履歴・運用引継ぎ確認後だけ有効。

| PR | head / base | included | unique changes | superseded by | can close after integration merge |
|---|---|---|---|---|---|
| #1 | day1-collector / main | yes | Add Go Rakuten collector and Jev state preparation for Day 1: .env.example, .gitignore, README.md, cmd/collect/main.go | integration-v0.1.0 | yes, after verification |
| #2 | day2-jev-evaluation / day1-collector | yes | Day 2: Jev評価CLI・version保存・blind review出力: .env.example, README.md, cmd/evaluate/main.go, cmd/evaluate/main_test.go | integration-v0.1.0 | yes, after verification |
| #3 | day3-llm-reranker / day2-jev-evaluation | yes | Day 3: fixed Jev Gate and Claude reranker CLI: .env.example, .gitignore, README.md, cmd/analyze-day3/main.go | integration-v0.1.0 | yes, after verification |
| #4 | day3-5-diversification / day3-llm-reranker | yes | Day 3.5: product families and deterministic diversification: README.md, cmd/diversify/main.go, go.mod, go.sum | integration-v0.1.0 | yes, after verification |
| #5 | day3-6-diversification-validation / day3-5-diversification | yes | Day 3.6: blind diversification quality validation: README.md, cmd/validate-diversification/main.go, internal/validation/analyze.go, internal/validation/cohort.go | integration-v0.1.0 | yes, after verification |
| #6 | day4-market-validation / day3-6-diversification-validation | yes | Day 4: reproducible external market validation: README.md, cmd/validate-day4/main.go, internal/market/market.go, internal/market/market_test.go | integration-v0.1.0 | yes, after verification |
| #7 | day5-content-prototype-validation / day4-market-validation | yes | Day 5: evidence-backed content prototype validation: README.md, cmd/validate-content/main.go, internal/contentvalidation/content.go, internal/contentvalidation/content_test.go | integration-v0.1.0 | yes, after verification |
| #8 | day6-canary-publication / day5-content-prototype-validation | yes | Day 6: prepare canary artifact and measurement setup: README.md, cmd/build-canary/main.go, content/published/compact-air-purifier-placement.md, internal/canary/canary.go | integration-v0.1.0 | yes, after verification |
| #9 | day6-1-production-publication / day6-canary-publication | yes | Day 6.1: deploy ChoiceLen canary to Cloudflare: .github/workflows/choiceLen.yml, .gitignore, README.md, cmd/build-choiceLen/affiliate.go | integration-v0.1.0 | yes, after verification |
| #10 | parallel-category-price-radar / day6-1-production-publication | yes | Parallel R&D: category discovery and price radar: cmd/discover-categories/main.go, cmd/refresh-offers/main.go, docs/parallel-category-price-radar.md, internal/discovery/discovery.go | integration-v0.1.0 | yes, after verification |
| #11 | family-sampling-radar-history / parallel-category-price-radar | yes | Foundation validation: family, sampling, and radar history: .github/workflows/price-radar.yml, README.md, cmd/radar-history/main.go, cmd/validate-foundation/main.go | integration-v0.1.0 | yes, after verification |
| #12 | choicelen-editorial-design-v1 / day6-1-production-publication | yes | ChoiceLen: editorial design system v1: .github/workflows/choiceLen.yml, cmd/build-choiceLen/main.go, cmd/build-choiceLen/main_test.go, docs/choicelen-editorial-design-v1.md | integration-v0.1.0 | yes, after verification |
| #13 | choicelen-human-editorial-v2 / choicelen-editorial-design-v1 | yes | ChoiceLen: human editorial v2: .github/workflows/choiceLen.yml, cmd/build-choiceLen/main.go, cmd/build-choiceLen/main_test.go, docs/choicelen-human-editorial-v2.md | integration-v0.1.0 | yes, after verification |
| #14 | choicelen-typography-layout-reset / choicelen-human-editorial-v2 | yes | ChoiceLen: typography and layout reset: .github/workflows/choiceLen.yml, docs/measurement-log.jsonl, docs/typography-layout-reset.md, internal/editorial/editorial.go | integration-v0.1.0 | yes, after verification |
| #15 | choicelen-home-brand-dedup / choicelen-typography-layout-reset | yes | ChoiceLen: remove duplicate home branding: .github/workflows/choiceLen.yml, docs/measurement-log.jsonl, internal/editorial/human.css, internal/editorial/human.go | integration-v0.1.0 | yes, after verification |

Merge preserves every PR tip as an ancestor; no cherry-picks, rebases or squashes. All unique changes are included relative to main. Draft integration PR is the single review target. Branch deletion is a separate future decision; private experiment outputs remain outside Git.
