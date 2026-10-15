# ChoiceLen D+7 observed evaluation

## Protocol
day7-v1、固定日2026-10-08、予定観測日2026-10-13、遅延実行日2026-10-15。
固定window: 2026-10-06 14:41:24〜2026-10-12 23:59 JST。protocol変更なし。
Status: PARTIAL / MANUAL_EXPORT_REQUIRED。未取得サービスを0扱いしない。

## Search Console
取得方法: ユーザー提供の正式Performance CSV ZIP。Web (text-based)、日付2026/10/06–2026/10/12。フィルタCSVにpage filterなし。
Googleの公式仕様では日別PerformanceはPacific Time (PT)。この期間はPDT (UTC−07:00)、実windowは2026-10-06 16:00〜2026-10-13 15:59:59 JST。固定windowと完全一致せず別集計として扱う。
latest exported day=2026-10-12 PT。サービス全体のlatest available date、preliminary/delayの有無はexportからはunknown。

| Scope | Impressions | Clicks | CTR | Average position |
|---|---:|---:|---:|---:|
| Property chart total | 1 | 0 | 0% | 6 |
| Canary article page row | 1 | 0 | 0% | 6 |

唯一の非zero日は2026-10-09 PTで、固定window内にある。国=ルーマニア、device=PC。
10/13以降の値をD+7値として加算しない。初日の公開後14:41:24〜16:00 JSTはこのexportに含まれないため、固定window全体の厳密なtotalはunknown（window内の1 impressionは確認）。
クエリCSVはheaderのみ。検索なしではなく、クエリ非開示/表示行なし。国・端末は1件の属性であり需要一般化に使わない。
URL Inspectionの現在indexed/crawl/Google canonical/user canonical、sitemap statusは未取得。10/9の検索表示実績はあるが、現在のindex状態の確認とは分離する。

一次仕様: https://support.google.com/webmasters/answer/17011259?hl=en
PT日付、privacy suppressionとchart/page aggregation差を確認。クエリ合計とpage合計の一致を要求しない。

## Cloudflare
Page views、Visits/unique相当、top pages、referrers、window/timezone/latest available date: unknown。手動export待ち。
10/6–7の管理/QAアクセスを分離不能ならunclassified。自然流入と断定しない。

## Rakuten Affiliate
Clicks/orders/commission JPY、period/timezone/reporting delay: unknown。正式report待ち。記事単位分解不能ならaccount aggregate。

## Price Radar
2026-10-15 09:08:33 JST current status（固定window値ではない）: 75 fetched、acquisition failures0、675 cumulative snapshots、59 cumulative events、candidate_count10。
固定window内のscheduled/attempted/success/failed/partial/new snapshots/events/price-drop candidatesはcanonical manifestsの別集計が必要。未集計値をcurrentから転記しない。

## Production Health
前回取得済み: top/article HTTP200、HTTPS、canonical/robots/sitemap検証通過。今回再取得・変更なし。

## Query Interpretation
開示queryなし。placement/product-specific/informational intentはunknown。平均6位という単一観測を安定順位・SEO成功と解釈しない。

## Pre-registered Decision
Final decision pending。measurement未取得はbrowser safetyによるaccess制約であり、設定不備の証明ではないためDATA_PIPELINE_ISSUEを自動適用しない。
現在未indexという証拠もないためINVESTIGATE_INDEXINGを断定しない。NO SEO PROFITABILITY DECISIONを維持。

## Next Action
URL Inspectionとsitemapのスクリーンショット、Cloudflare/Rakutenの期間付きexportを取得。残項目後にprotocol分岐を選ぶ。新記事/施策/設定変更なし。

## Caveats
前回と再開時にbrowser current URLの安全確認ができず自動操作停止。ユーザーが公式Search Consoleを開いた後も同様で、manual exportへ移行した。
D+7前のUI/editorial変更からtraffic差の因果推論をしない。0/unknown/delayedは別。ZIPはprivateで保持しpublic repoへcommitしない。
Import checked_at: 2026-10-15T20:30:39.5349180+09:00
Source ZIP SHA256: 9c86b86e8be2f02a8de7830ba48831e40e16cf359248d8a5988b037c73c464b5


## API observer acquisition (2026-10-15T13:15:41Z)

Protocol: day7-v1. Requested: 2026-10-06T14:41:24+09:00 — 2026-10-13T00:00:00+09:00 (Asia/Tokyo; upper bound exclusive).

### cloudflare

Status: unknown. Reason: AUTHORIZATION_REQUIRED: dedicated Cloudflare Analytics Read token/account missing; deploy token not used

Source: https://api.cloudflare.com/client/v4/graphql; method=read-only API; checked_at=2026-10-15T13:15:34Z.

Actual window: 2026-10-06T14:41:24+09:00 — 2026-10-13T00:00:00+09:00; timezone=Asia/Tokyo; granularity=instant; exclusive upper bound; latest_available_date=unknown; reporting_delay=unknown.
- Dataset=rumPageloadEventsAdaptiveGroups (official Web Analytics/RUM); adaptive sampling may apply.
- Visits is not unique people. Admin/QA cannot be separated: unclassified.
- Top pages/referrers capped at 1000; do not sum group visits into site unique visitors.
- Latest available date and reporting delay unknown unless separately established.

### production_health

Status: known. Reason: none

Source: previous read-only production verification; method=existing verification JSON; checked_at=2026-10-15T13:15:41Z.

Actual window: 2026-10-15T08:35:14Z — 2026-10-15T08:35:14Z; timezone=UTC; granularity=snapshot; latest_available_date=unknown; reporting_delay=unknown.
- Reused prior technical checks; no fresh production request or historical D+7 indexing inference.

### radar

Status: known. Reason: Validated 11 canonical run bundles

Source: private canonical Radar history; method=read-only API; checked_at=2026-10-15T13:15:34Z.

Actual window: 2026-10-06T14:41:24+09:00 — 2026-10-13T00:00:00+09:00; timezone=Asia/Tokyo; granularity=instant; exclusive upper bound; latest_available_date=2026-10-15T00:08:33Z; reporting_delay=unknown.
- attempted: 9 (known)
- events: 34 (known)
- failed: 1 (known)
- partial: 0 (zero)
- price_drop_candidates: 5 (known)
- snapshots: 600 (known)
- successful: 8 (known)

Expected slots=6; last successful within window=2026-10-12T00:08:32.8058251Z.
- Daily 09:07 JST expected slots; attempted includes manual and failed diagnostic runs, not proof of scheduled execution.
- New snapshots/events filtered by observation timestamp; historical import is not a new attempted run.
- Candidates count threshold-qualified price-drop events, not unique products or capped ranked list.
- Current manifest cumulative counts kept separate. Canonical manifests cannot prove attempts that failed before remote persistence.

### rakuten

Status: unknown. Reason: MANUAL_EXPORT_REQUIRED: no public Rakuten Ichiba Affiliate reporting API confirmed; reporting CSV absent

Source: official Rakuten Affiliate manual report; method=manual CSV; checked_at=2026-10-15T13:15:34Z.

Actual window: unknown — unknown; timezone=unknown; granularity=CSV explicit period rows; latest_available_date=unknown; reporting_delay=unknown.
- clicks: <nil> (unknown)
- commission_jpy: <nil> (unknown)
- orders: <nil> (unknown)
- scope=account_aggregate unless independently attributable; no article attribution inferred.
- Only recognized period/metric/status columns retained; account and personal fields omitted.
- Totals not silently combined with daily rows: supply one non-overlapping aggregation level.

### search_console

Status: unknown. Reason: AUTHORIZATION_REQUIRED: Google read-only ADC missing

Source: Google Search Console official API; method=read-only API; checked_at=2026-10-15T13:15:34Z.

Actual window: 2026-10-06 — 2026-10-12; timezone=America/Los_Angeles; granularity=daily (inspection/sitemap would be current); latest_available_date=unknown; reporting_delay=unknown.
- Manual/API reconciliation=PENDING

### search_console_manual

Status: known. Reason: none

Source: user-provided official Search Console export; method=manual CSV ZIP; checked_at=2026-10-15T13:15:34Z.

Actual window: 2026-10-06 — 2026-10-12; timezone=America/Los_Angeles; granularity=daily; latest_available_date=unknown; reporting_delay=unknown.
- average_position: 6 (known)
- clicks: 0 (zero)
- ctr: 0 (zero)
- impressions: 1 (known)
- Manual PT window retained; no JST-exact equivalence. Current index state is not supplied by Performance export.

### sitemap

Status: unknown. Reason: AUTHORIZATION_REQUIRED: Google read-only ADC missing

Source: Google Search Console official API; method=read-only API; checked_at=2026-10-15T13:15:34Z.

Actual window: 2026-10-06 — 2026-10-12; timezone=America/Los_Angeles; granularity=daily (inspection/sitemap would be current); latest_available_date=unknown; reporting_delay=unknown.

### url_inspection

Status: unknown. Reason: AUTHORIZATION_REQUIRED: Google read-only ADC missing

Source: Google Search Console official API; method=read-only API; checked_at=2026-10-15T13:15:34Z.

Actual window: 2026-10-06 — 2026-10-12; timezone=America/Los_Angeles; granularity=daily (inspection/sitemap would be current); latest_available_date=unknown; reporting_delay=unknown.

### Pre-registered decision

NO SEO PROFITABILITY DECISION

Pending: current URL inspection unavailable; sitemap unavailable; search_console incomplete; url_inspection incomplete; sitemap incomplete; cloudflare incomplete; rakuten incomplete; CONTINUE requires independently verified production health and operational measurement

No browser automation, production changes, or SEO/profitability conclusion. Prior manual export and browser safety-stop history remain in this document.

## API observer execution summary — 2026-10-15

Status remains PARTIAL / AUTHORIZATION_REQUIRED. This is delayed execution of day7-v1, not a new D+9 protocol. Parser/schema/decision fixtures were committed before this API run (29e075c); subsequent vet fix only names CLI struct fields and does not change evaluation rules.

| Source | Result |
|---|---|
| Search Analytics / URL Inspection / sitemap APIs | Google read-only ADC absent; no request made |
| Cloudflare Web Analytics API | Dedicated Analytics Read token/account absent; deploy token not used; no request made |
| Rakuten Affiliate | Public Ichiba Affiliate reporting API not confirmed; manual CSV absent; all metrics unknown |
| Manual Search Console | Preserved impressions1/clicks0/CTR0/position6; PT window, no disclosed query; API reconciliation PENDING |
| Radar canonical history | 11 run bundles retrieved/hash-validated via existing private read access; no collector run |

### Fixed JST-window Radar (separate from current)

Window: 2026-10-06 14:41:24 JST inclusive to 2026-10-13 00:00 JST exclusive.

- Expected 09:07 JST daily slots: 6.
- Canonically persisted attempts: 9; SUCCESS8 / FAILED1 / PARTIAL0. Includes manual/diagnostic runs; not 9 scheduled slots.
- Distinct canonical snapshots with observed_at in window: 600.
- New snapshots reported by successful non-import runs: 525. The remaining 75 are previously observed imported baseline observations whose timestamps are also inside the window; not fabricated historical prices.
- Events with observed_at in window: 34.
- PRICE_DROP_CANDIDATE events in window: 5. This is event count, not unique products or cap2 ranked candidates.
- Last successful completion in window: 2026-10-12 09:08:32 JST.
- Known failed run is the earlier cloud/IP-denied diagnostic, not evidence that every Windows scheduled run failed.
- Attempts that never reached canonical storage cannot be counted by remote manifests alone.

Current status remains separately labeled as of 2026-10-15 09:08:33 JST: 675 cumulative snapshots /59 events /candidate_count10, 75 fetched/failure0. Current cumulative values were not substituted for D+7.

### Pre-registered decision / next action

NO SEO PROFITABILITY DECISION. Final CONTINUE/indexing decision remains pending current inspection/sitemap and remaining measurement. Missing observer auth does not prove production DATA_PIPELINE_ISSUE; no current not-indexed evidence has been obtained. No rule thresholds changed.

Next: supply Google read-only ADC and dedicated Cloudflare Account Analytics Read credential using docs/operations.md; rerun only gsc/cloudflare into a new output directory. Rakuten remains manual CSV. No account/credential is requested in chat. Current output is in ignored outputs/day7-observation. No raw account response or screenshot is committed.

Validation: go test ./... /go vet ./... /git diff --check. site/, content/published/, canonical/affiliate settings, frozen protocol, workflows and scheduler unchanged.

## Cloudflare read-only API connected

Dedicated observer environment is configured; no credential values are included here. Acquisition checked_at: 2026-10-15T13:39:25.0000000+00:00
Fixed requested JST window: 2026-10-06 14:41:24 inclusive through 2026-10-13 00:00 exclusive. Live schema discovery and official rumPageloadEventsAdaptiveGroups query succeeded.

- Site total page views (RUM count): 80.
- Site total sum.visits: 80. This is not unique persons.
- Home path: 70 page views; article path: 10 page views.
- refererHost is empty for all 80. This is no reported referrer, not proof of organic/direct human traffic.
- Admin/QA access cannot be separated: unclassified.
- unique visitors: unknown. Reporting delay/latest available date: unknown.
- Prior Cloudflare AUTHORIZATION_REQUIRED status is superseded by this acquisition only. Google and Rakuten remain pending.

Private machine-readable source: outputs/day7-observation/cloudflare-20261015-1/observation.json. No production change.
