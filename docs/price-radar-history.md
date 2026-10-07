# Price Radar durable history — 2026-10-07 JST

Technical decision: **HISTORY_READY**, using private GitHub canonical history and a verified Windows scheduled task on the existing authorized collection host. Daily execution requires that host to be powered on, connected and the existing user session logged in. This is not an always-on cloud service; D+7 coverage is not guaranteed when the host is unavailable.

## Storage choice and authentication

R2 was evaluated first. Existing Cloudflare authorization returned HTTP 403 for the bucket operation; no bucket, subscription or paid plan was created. The R2 adapter/template remains optional and inactive.

Canonical history is the private repository `AkiraShimizu0/choicelen-radar-history` (privacy verified through the GitHub API). Public `rakuten-product-search` contains code/docs only. Existing GitHub CLI authorization is read into memory on the collection host, never printed or placed in configuration. Private Actions has only Rakuten app/access-key secrets and uses its own ephemeral `GITHUB_TOKEN` for repository writes; no cross-repository PAT or Cloudflare token was copied there. Temporary bootstrap secrets were removed after import. Existing public secrets were not exposed.

Private objects:

```text
cohort.json.gz
runs/YYYY/MM/DD/<run_id>.json.gz
ranked/<run_id>.json.gz
latest-success.json
latest-attempt.json
```

Run bundles are immutable at the application layer and include cumulative observed snapshots, events, run manifest, original API byte hashes and source commit. The two pointers are rebuildable caches; Git preserves their revisions. Local SQLite is disposable processing/cache. GitHub administrators can still delete/force-push: this is not WORM storage. At 75 products/day until D+7, small gzip bundles are a practical fallback; cumulative bundles grow quadratically over long retention, so design a segmented object archive before larger/long-running deployment. No history was estimated or backfilled.

## Verified real runs

| Run | Result | observations | new snapshots | duplicate | cumulative snapshots | events |
|---|---|---:|---:|---:|---:|---:|
| import-observed-20261006 | SUCCESS | existing 75 | 75 | 0 | 75 | 0 |
| 20261006T235750.485923200Z | SUCCESS | 75 | 75 | 0 | 150 | 1 |
| 20261007T000601.161221300Z | SUCCESS | 75 | 75 | 0 | 225 | 1 |
| 20261007T000956.806619400Z | SUCCESS (scheduled task manual start) | 75 | 0 | 75 | 225 | 1 |

The one event is **availability_changed**, observed 2026-10-06 23:58:33 UTC, with API price unchanged at JPY 18,800. It is an explicit API-signal change, not a guarantee of physical stock or restocking. Price-drop candidates: **0**. Same content in a different one-hour bucket is intentionally retained as another observation; within the same bucket duplicates are discarded. Latest success completed **2026-10-07 09:11:26 JST**, fetch failures 0; snapshot SHA256 `09d3541a79812bfc8f691a416f2f2c6fa8edd22666c5499234add11ca27a9c24`.

The imported existing observation had no source-commit field; its missing provenance remains unknown rather than fabricated. Later manifests record the pinned source commit.

## Scheduler and failed cloud validation

Private GitHub Actions supports `workflow_dispatch`, concurrency without cancellation, 15-minute timeout, bounded HTTP retry, tests/build and status reporting. Actual manual cloud runs **37549593470** and **37550145483** failed. The latter explicitly returned `CLIENT_IP_NOT_ALLOWED` (HTTP 403) on the first Rakuten request: 0 successes, 75 unresolved products, of which 74 were not attempted after detecting global denial. No price/availability events were invented and the previous good history remained retrievable. An earlier concurrent pointer update also hit 409; bounded retry/CAS protection was added and tested.

Do not widen the API IP allowlist automatically. The known-failing cloud cron was removed; manual workflow remains for diagnostics. Daily scheduling instead uses Windows Task Scheduler:

- Task: `ChoiceLen-PriceRadar`
- Daily **09:07 JST**, next run 2026-10-08 09:07
- Limited existing user, interactive logon; no stored password/new service account
- Start when available; ignore overlapping task starts; 15-minute execution limit
- Native Go executable in private ignored experiment outputs; no visible helper window
- Manually started through Task Scheduler on 2026-10-07 09:09:54; final task exit result **0**
- Collector pinned to `16db3a67ea81c56409489d3a554c367b96ee947e`

The task reads the existing local secret file and existing `gh` authorization. Private local task-definition XML records its precise action/settings without secret values. Changing the source does not silently replace the pinned scheduled binary. Build/register again only after validating a future collector change.

## Operation

From the repository, on the authorized host:

```powershell
go run ./cmd/radar-history -backend github -stage scheduled -gh-path "C:/Program Files/GitHub CLI/gh.exe" -env ../rakuten.env
# Read-only status: set GITHUB_TOKEN from existing secure authorization without logging it.
go run ./cmd/radar-history -backend github -stage status
# Download to a NEW private path; existing-file overwrite is refused.
go run ./cmd/radar-history -backend github -stage dump -out <new-private-path.json.gz>
go run ./cmd/radar-history -stage facts -out <private-path.json.gz>
```

`Start-ScheduledTask -TaskName ChoiceLen-PriceRadar` is the verified manual-run interface. `Get-ScheduledTaskInfo` reports last/next run and exit result. Remote `status` reports last attempt, last success, successes/failures, snapshot/event/candidate counts. No notification/publication integration exists.

## Failure safety, rules and tests

Require all 75 observations for SUCCESS. PARTIAL/FAILED manifests preserve previous snapshots/events and do not advance latest-success. Commit verifies hashes/read-back and rejects stale pointer updates. Original RawJSON bytes are base64-wrapped so JSON reformatting cannot invalidate original hashes. A failed object write cannot remove a previous good run. Event calculations reconstruct a temporary SQLite cache from canonical observations. Duplicate handling and price-drop thresholds remain unchanged: >=10% AND >=JPY 1,000; family cap 2 for candidate output. Cohort remains 75 products/56 families.

Tests cover private remote read/write, immutable overwrite refusal, raw-byte round trip, CAS/stale pointer refusal, partial failure, previous-good preservation, duplicate suppression, and reproducible event calculations. Existing radar tests continue to cover thresholds, shop separation, missing values, ranking and family cap. Real read-back, duplicate and failed-cloud retrieval also passed.

## API conditions, costs and remaining limitations

Checked 2026-10-07 JST: [current item search API](https://webservice.rakuten.co.jp/documentation/ichiba-item-search), [usage guide](https://webservice.rakuten.co.jp/guide), [terms](https://webservice.rakuten.co.jp/guide/rule), [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions), [R2 pricing](https://developers.cloudflare.com/r2/pricing/).

Current search API requires paired app ID/access key, permits item-code lookup and warns against rapid repeated identical-URL access. Fetches are paced at 1.2 seconds, timeout 20 seconds, at most two retries. Daily 75-item collection is small and passed on the configured host; actual quotas/access controls remain app-specific. No claim is made that this grants blanket terms approval: internal/private research applicability should be confirmed with Rakuten before broader deployment; the linked English terms are reference only and Japanese terms govern. No API data is published from this experiment.

Radar makes **zero Jev/Claude calls**. R2 usage is zero because provisioning failed. GitHub storage uses the existing private-repository service, not a new paid plan. Two short private Action attempts consumed existing account allowance; exact account-wide bill/remaining included minutes are unknown. Daily collection runs locally and consumes no Actions minutes. No billing limit or paid subscription was changed. Local host availability, IP changes, credential expiry and repository growth remain operational risks; scheduled task exit/status should be checked before D+7 analysis.

ChoiceLen content/settings, semantic evaluation, family assignments and event thresholds remain unchanged. D+7 Canary measurement is a separate task on 2026-10-13.
