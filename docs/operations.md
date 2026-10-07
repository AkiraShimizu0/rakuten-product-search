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
