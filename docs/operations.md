# Operations

## ChoiceLen
Production: https://choicelen.page/
Article: /articles/compact-air-purifier-placement/
Static HTML/CSS + Workers Static Assets。既存 .github/workflows/choiceLen.yml を使用。許可されたbranchの明示deploy marker、またはworkflow_dispatchで実行。GitHub既存secretからCloudflare認証、pinned official Wranglerでdry-run→deploy。
integration-v0.1.0はpush deploy対象外。今回workflow_dispatchも行わない。mainへの将来統合だけで自動公開されるとは扱わない。
公開時はgo test ./... /go vet ./... /git diff --check、Desktop/Mobile preview、claim/SEO/affiliate確認。公開後verify-choiceLen、HTTP/TLS/canonical/robots/sitemap/404/Analytics/source/広告確認、measurement-log保存。ロールバックも別途公開認可が必要。

## Price Radar
既存Windows Task Scheduler: TaskPath=\、TaskName=ChoiceLen-PriceRadar。
毎日09:07 JST。limited user/interactive logon、15分上限、overlapはignore、start-when-available。host稼働・ネット接続・ログインが必要。
actionの場所は Get-ScheduledTask -TaskName ChoiceLen-PriceRadar のActionsをローカルで確認。private ignored outputs内のpinned Go binaryを使う。collector source 16db3a67ea81c56409489d3a554c367b96ee947e。統合checkoutは実行binaryを置換しない。

確認: Get-ScheduledTaskInfo -TaskName ChoiceLen-PriceRadar （LastRunTime/LastTaskResult/NextRunTime）。
manual: Start-ScheduledTask -TaskName ChoiceLen-PriceRadar （追加API取得になるため意図的に実行）。
remote status: go run ./cmd/radar-history -backend github -stage status。既存gh認証を安全にprocess環境のGITHUB_TOKENへ渡し、値は出力しない。statusはread-only。
scheduled CLI: go run ./cmd/radar-history -backend github -stage scheduled -gh-path "C:/Program Files/GitHub CLI/gh.exe" -env ../rakuten.env
dumpは新しいprivate出力先を指定。正史はprivate AkiraShimizu0/choicelen-radar-history、SQLiteはcache。public repoへ履歴をcommitしない。

失敗復旧: task exit/last attempt/last successを照合。403 CLIENT_IP_NOT_ALLOWEDならretryを増やさず既存allowlist/host接続を調査し、権限を自動拡大しない。previous good runを維持しpartialをSUCCESS扱いしない。認証期限/ネット障害を解消後にmanual run、read-backとevent再現を確認。欠測のbackfillは行わない。history pointerはcacheでrun objectsから復元、既存objectsを削除/上書きしない。
価格candidate条件>=10% AND >=1000JPY、family cap2。page/order UIから実在庫保証を推測しない。
R2 provisioningは未完了、cloud cronはIP拒否で停止済み。今回scheduler変更/Mac/cloud移行なし。

## API credentials
Rakuten: ignored ../rakuten.env またはprocess environment、.env.example参照。
Jev/Claude: 既存config.LoadEnvとenv templateで必要なkey名を確認。モデル/rubric/fingerprint固定。dry-run後だけ有料工程を開始。
Cloudflare: GitHub既存Actions Secrets。ローカルへコピーしない。
GitHub Radar: 既存gh credential storeからmemoryへ読み込み、private jobはephemeral GITHUB_TOKEN。secret値をdocs/log/CLI引数へ記載しない。
再現性manifestはprovider/model/prompt hash/version/weights/token costを保存し、credentialは除外。

## Limits
Windows停止時の欠測、IP/credential expiry、private cumulative bundlesの長期肥大化に注意。D+7は別protocolで観測し、0とunknownを区別。定期運用の成功を統合テストだけで保証しない。

## D+7 / D+14 / D+28 observer

CLI: go run ./cmd/observe-choiceLen -from 2026-10-06 -to 2026-10-12 -out outputs/day7-api-run-NEW
Default source=all。-source gsc/cloudflare/radar/rakuten、またはcomma-separated。-dry-runはnetwork/credential/output accessなし。再実行は必ず新しい出力directory。固定windowの-defaultはD+7、D+14等では-from/-toと-window-start/-window-endの両方を明示。-propertyは実際に認可されたexact propertyを使う。今回の既存Domain property既定はsc-domain:choicelen.page、URL-prefixならhttps://choicelen.page/。架空のpropertyへ切り替えない。

### Google authentication (AUTHORIZATION_REQUIRED if absent)

公式Search Console APIを有効化したGoogle Cloud projectのDesktop OAuth clientを用意し、本人がGoogle公式consentでread-only承認。https://console.cloud.google.com/apis/credentials
必要scopeは https://www.googleapis.com/auth/webmasters.readonly のみ。URL Inspection・Search Analytics・sitemaps getに共用。Search Console propertyへの既存アクセスも必要。書き込みscope/webmastersは不要。

Google Cloud CLIを使用する場合:
gcloud auth application-default login --client-id-file <private-desktop-client.json> --scopes=https://www.googleapis.com/auth/webmasters.readonly
本人が表示された公式認証URLを開いて承認。OAuth client/refresh tokenをチャットへ貼らない。自動ブラウザ操作・cookie抽出は行わない。
生成されたauthorized_user ADCをGOOGLE_APPLICATION_CREDENTIALSで指定、またはWindowsの標準 %APPDATA%/gcloud/application_default_credentials.json。credential fileはrepository外のprivate pathで保管。
observerはauthorized_user ADC refreshと専用process env CHOICELEN_GOOGLE_ACCESS_TOKENに対応。service_account/workload identityは現時点で未対応。tokenはこのscopeだけで発行したものを使う。scopeを指定してrefreshするが、過去に広いscopeを付与したcredentialを流用しない。

### Cloudflare authentication

https://dash.cloudflare.com/profile/api-tokens で本人が最小read-only tokenを作成し、対象accountのAnalytics Read相当だけを許可。deploy tokenは利用禁止。
専用の CHOICELEN_CF_ANALYTICS_TOKEN と CHOICELEN_CF_ACCOUNT_ID をignored observer.envまたはprocess envへ保存。既存CLOUDFLARE_API_TOKENへのfallbackは実装しない。-env-fileは明示時のみ読み込む。
公式RUM dataset rumPageloadEventsAdaptiveGroupsがWeb Analyticsに対応。認証後にschema introspectionを行い、count/requestPath/refererHostとtime/host filterを検証。sum.visitsがschemaにない場合はunknown。
schema/権限/planで取れなければCLOUDFLARE_WEB_ANALYTICS_API_UNAVAILABLEとしてmanual exportを継続。Zone HTTP requestsをpage viewsへ置換しない。schemaが確認できてもretention/query-range制限で過去日が取得不能な場合があり、empty totalを0と推定しない。
current pipelineはlive accountで未認証。実schema・dataset権限は初回接続で確認が必要。

### Read-only Radar

-gh-auth -gh-path "C:/Program Files/GitHub CLI/gh.exe" で既存認証をmemory内だけに読み込む。private repo/read-only GETのみ。既存canonical bundlesをhash検証し、run一覧とlatest-successを取得。scheduler/run/refreshは起動しない。
windowは開始inclusive、終了exclusive。expected slotsは09:07 JSTの日次機会、attemptsには手動/diagnosticも含むためscheduler成功数とは同義ではない。起動直後にremoteへ到達せず終了したrunはcanonical一覧だけでは検出不能。
import manifestはattemptから除外。snapshots/eventsはobserved_atで切り出す。candidateは閾値該当event数であり、unique商品数/ランキングcap後の数とは別。current manifestを別欄に保持。

### Manual inputs / provenance

-gsc-manual-zip <private-export.zip> で既存正式exportを保存・APIと対比。CONFIRMEDは同じwindow/値が一致した場合だけで、query suppressionやAPI webとmanual text-basedのsource差を消さない。既存manual=表示1/クリック0を上書きしない。
-rakuten-report <private-report.csv>。楽天市場アフィリエイト公式の公開reporting APIは確認できていない。Rakuten Advertising/LinkShareのAPIを別accountへ流用しない。
UTF-8/Shift-JIS CSV、recognized columns: period/期間/日付、clicks/クリック数、orders/注文件数/売上件数、commission_jpy/成果報酬/成果報酬(円)、status/状態。公式CSVが異なるheader/preambleなら匿名最小sampleでadapterを追加する。対応未確認の形式を黙って推定しない。
CSVは日別か合計の一方だけ、重複しない単位で用意。metric欠損はunknown、CSVなしもunknown、明示0だけzero。periodはrowに保持、異なる期間を固定D+7へ自動合算しない。scopeはaccount_aggregate。報酬statusや反映遅延、timezoneがCSVにない場合はunknown。
CSVの不要なaccount/personal columnsをoutputへ持ち込まない。

### Output / report / security

出力observation.jsonにrequested/actual window、timezone、granularity、checked_at、latest_available_date、reporting_delay、known/zero/unknown/delayed/unavailableを保存。取得範囲の最終日をサービスの最新反映日と同一視しない。APIのmetadata.first_incomplete_dateを保存し、欠落日の0補完は禁止。
-health <prior-day6-http-check.json> で取得済みtechnical検証を再利用。-append-report docs/day7-observed-results.md は既存docへ追記のみ。browser safety historyとmanual値は残る。判定は不足時pending。auth不足をproductionのDATA_PIPELINE_ISSUEと断定しない。
observation.jsonはignored outputsへ。public PRにはコード/匿名fixture/要約だけ。raw response/header/Authorization/OAuth credentialをlog保存しない。HTTP診断はstatusのみ。401/403はretryしない。429/5xx/transportは最大2retryのexponential backoff。redirectを追わずtoken転送を防ぐ。
WriteNewとnew-directory拒否で既存outputを保護。再取得は-sourceで限定し、異なるwindowのobservationを無条件mergeしない。

公式資料確認日2026-10-15:
- https://developers.google.com/webmaster-tools/v1/searchanalytics/query
- https://developers.google.com/webmaster-tools/v1/urlInspection.index/inspect
- https://developers.google.com/webmaster-tools/v1/sitemaps/get
- https://developers.google.com/webmaster-tools/v1/sitemaps (contents.indexed deprecated、submittedは送信URL数)
- https://cloud.google.com/docs/authentication/application-default-credentials
- https://developers.cloudflare.com/data-localization/metadata-boundary/graphql-datasets/
- https://developers.cloudflare.com/analytics/graphql-api/features/discovery/introspection/
- https://developers.cloudflare.com/analytics/graphql-api/getting-started/authentication/api-token-auth/
- https://affiliate.rakuten.co.jp/guides/report/
