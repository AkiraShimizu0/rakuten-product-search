# ChoiceLen / Jev Money Engine

メーカー一次資料・仕様に基づいて購入判断候補を発見し、検証した少数の記事をChoiceLenへ届けるGo基盤です。v0.1.0は統合候補で、main未merge・tag未作成です。

## What it does
Rakuten → deterministic Go filter → Jev v1 Gate → Claude reranker → family diversification → category / market validation → evidence-backed article → ChoiceLen。
Price Radar → product snapshot history → event detection → deterministic ranking。記事生成・公開には自動接続しません。

## Architecture
Collector → normalized SQLite → filter → Jev → Gate → Claude → diversification
→ category / market / evidence validation → static ChoiceLen
75-product radar cohort → paced fetch → private append-oriented history → events → family-capped candidates

Canonical Radar historyは非公開GitHub repository、SQLiteは処理/cache。Windows既存Task Schedulerが収集し、R2とcloud schedulerは現用方式ではありません。

## Setup
Go 1.26以上。SQLiteはpure-Go driverでPython不要。
go mod download / go test ./... / go vet ./...
.env.example と docs/operations.md を参照。秘密値はローカルignored env /既存secret storeに保存。取得DB、API応答、履歴、スクリーンショットはcommitしません。

## Main commands
実在するCLI。go run ./cmd/<name> -h で引数確認。API工程はconfig/fingerprint固定とdry-run費用見積もりが必要。

| CLI | 役割 |
|---|---|
| collect / inspect | 楽天収集・正規化・filter・状態確認 |
| evaluate / gate / rerank | Jev、固定Gate、Claude評価・manifest・費用 |
| diversify / validate-diversification / analyze-day3 | family cap、独立judge検証・集計 |
| discover-categories | genres / collect / jev / claude / report |
| validate-foundation | prepare / collect / evaluate / report / family |
| refresh-offers | cohort再取得・append snapshot・event/ranking |
| radar-history | private history、scheduled / status / dump 等 |
| validate-day4 / validate-content | 市場構造検証・evidence/claim/judge検証 |
| build-canary / build-choiceLen | frozen記事から静的成果物生成 |
| preview-choiceLen / verify-choiceLen | ローカルpreview・HTTP/TLS/SEO検証 |

新実験は出力先を分離。書き出しの多くは既存output上書きを拒否。CLI存在はprivateデータやcredentialの配布を意味しません。

## Current status
- ChoiceLen Canary 1 article live: https://choicelen.page/ 。UIはPR #15相当。
- Category Discovery operational。除湿機、電動シュレッダー、電気圧力鍋が候補。市場/SEO/収益性は未検証。
- Price Radar history operational: 75 products /56 families、Windows毎日09:07 JST。稼働hostが必要。
- Family Resolver v2 still HOLD / NEEDS_MORE_VALIDATION。production resolver未置換。
- Local LLM not used。
- D+7 pending: 2026-10-13。docs/day7-evaluation-protocol.md参照。

## Non-goals
mass AI article publishing、unsupported claims、Radarからの自動公開、SNS自動投稿、automatic investment executionは対象外。

## Baseline and provenance
docs/v0.1.0-baseline.md、docs/pr-consolidation-plan.md、docs/operations.md参照。Day別docsはprovenanceとして維持。統合では本番deploy、旧PR merge/close、tag作成を行いません。
