# Dehumidifier Production Release

## Approval

今回のユーザー依頼「Final Release / Production Publish」を、PR #25でREADY_TO_PUBLISHとなった広告リンクなし完成稿を、検証済み2記事構成で公開する明示承認として記録した。新claim／新機種／広告／デザイン変更は承認範囲に含めず、追加していない。

**Final Status: NOT_PUBLISHED。AUTHORIZATION_REQUIRED。** 公開前ゲートは通過したが、既存GitHub Actions SecretによるCloudflare認証が拒否され、本番deployは完了していない。認証復旧前に同じtokenで再実行していない。

親は`dehumidifier-publication-integration` / PR #25 / `5201126f279a9c774db2b0b50a6652bfde35f14a`。新branchは`dehumidifier-production-release`。既存PRをmerge／closeせず、main／integration-v0.1.0へもmergeしていない。既存判定・評点・protocolは保持した。

## Release Metadata

|Field|Value|
|---|---|
|Build-time published_at|`2026-10-21T04:30:01+09:00`|
|source_checked|`2026-10-20`|
|deployment_commit（builder／content状態）|`7c096481895625b96140f4ef073ef0dc82c4d754`|
|Artifact checkout commit|`0d5d86fe36f8e2fe3bf6053e760b4bd9e20a8f65`|
|canonical_url|`https://choicelen.page/articles/winter-laundry-dehumidifier-conditions/`|
|article_slug|`articles/winter-laundry-dehumidifier-conditions`|
|builder_version|`two-article-v1`|
|updated_at|unset|

公開日時はbuild直前に実時刻をTokyo Standard Timeへ変換して取得した。**deploy不成立のため、この時刻を実際に公開された日時とは扱わない。** 公開再開時には実時刻のmetadataから新しいreleaseを生成する。

原本確認日とのcalendar差は1日。PR #24／#25から連続した検証のため全面再調査は行わず、資料確認日を公開日に書き換えていない。完成稿のidentityとhashは同一。

## Build

監査済みfinal Markdown SHA256：`b15b4b3f0648a965d63452a32a7a2087057174b5e65b7acfa36fe445a2533e47`。

claims-v2 SHA256：`f958bcb5d543169e9aff7923173be9cb512566f85f0ecbb03d0b3107190799ce`。22claim／9公式参照／4表／4機種のscopeを維持。内容再評価・新たなjudge実行なし。DD参照は属性とJSON source mapへ維持する。

`-stage release`で新規private `outputs/dehumidifier-production-release/release/`へ9ファイルを生成。RCをそのままdeployせず、公開metadataを反映したhome・新記事・sitemap・manifestを生成した。ファイル構成はRCと同じ9件。

release CLIへcanonical URL／slug／versionの完全照合と、metadata commit＝実際のbuild HEAD、builder／content／configのclean状態の検査を追加した。既存legacy frozen hashは維持する。監査済みMarkdownは内容不変でpublished sourceとして追加し、LFを固定した。

## Pre-deploy Validation

全PASS後にのみ、検証した9ファイルをそのまま`site/`へ配置した。deploy直前の再buildなし。

- expected files9、各出力hash一致、内部file／fragment／CSS参照正常。
- 3ページのself-canonical各1、sitemap3URL、URL重複／旧記事消失0。
- localhost／placeholder／candidate noindex0。robotsは従来のAllowと正式sitemapを維持。
- 1本目のHTML／CSS／affiliate／disclosure／source／公開metadataは従来生成と一致。
- deploy前に現在のproductionとも比較。Cloudflare自動beaconとその挿入改行、CRLFを除いて1本目HTML一致、CSS完全一致。
- Markdown全行の可視text／表cell／数値／型番の順序照合、22claim join、9source URL保持。
- 27℃／60%／50・60Hz、cadoナイト送風、Panasonic音ひかえめ除湿、IJC全温度・Panasonic連続上限unknown、排水／停止／購入保留を保持。
- 除湿機の価格・購入・affiliateリンク0。1本目のsponsored1件と正式destinationは変更なし。
- Desktop1440×1000／Mobile390×844でhome・1本目・除湿機を短く再確認。6画面保存、横page overflow0、資料確認と公開日の区別正常、preview tracking script0。
- `go test ./...`／`go vet ./...`／`git diff --check` PASS。

9ファイルの集合hash：`5f9fb8bfef41b4716307d918f903c43e7736285c09b4b7316277474e1e7ce5a0`。

計算方法：相対path順に`path TAB file_SHA256 LF`をUTF-8で連結し、そのSHA256。build-manifest.json自身も集合hashに含める。これは**検証済み・deployを試みたartifact**のhashであり、deploy済みhashではない。

artifact commitは生成済みsiteだけを追加した後続commitで、builder／content commitとは分けた。Actionsはfull historyを取得し、`-verify-release site`でmanifest・全hash・claim／source・リンク・metadataを再検査し、manifestのsource commitがancestorであることと、そこからcheckoutまでのbuilder／content／Wrangler差分0を要求する。未検証buildの混入を防ぎ、commitを自己参照するmanifestも作らない。

## Deployment

既存Workers Static Assets設定、account／project／domain／route／Wrangler設定は変更なし。既存GitHub ActionsのCloudflare Secretsを使い、workflow_dispatchのみを実行した。新secret設定・token作成・移動・ローカルコピーなし。

[Actions run 39484227719](https://github.com/AkiraShimizu0/rakuten-product-search/actions/runs/39484227719)

- Checkout：`0d5d86fe36f8e2fe3bf6053e760b4bd9e20a8f65`。
- GitHub側の全tests／vet／diff check：SUCCESS。
- 検証済みartifactの再検査（再生成なし）：SUCCESS。
- pinned Wrangler install／dry-run：SUCCESS。
- production deploy：FAILURE。2026-10-21 04:32:53 JST、Cloudflare APIの`/workers/services/choicelen-canary`でAuthentication error `10000`、続くaccount確認でInvalid access token `9109`。
- 原因は認証拒否まで確認。期限切れ／失効／設定誤りのどれかは断定していない。
- 新version ID／成功deployment timestamp／deployed artifact hash：**なし**。

既存Secretの値は表示・保存せず、失敗ログもerror codeと非機密のreceiptに限定した。tokenが無効なため、同じ認証でdeployを繰り返していない。

## Production HTTP

失敗後のread-only HTTPS取得：

|URL / path|HTTP|結果|
|---|---:|---|
|`https://choicelen.page/`|200|既存1記事homeのまま|
|`/articles/compact-air-purifier-placement/`|200|本文／canonical／広告／CSS／source保持|
|`/articles/winter-laundry-dehumidifier-conditions/`|404|未公開|
|`/assets/site.css`|200|既存CSSと一致|
|`/sitemap.xml`|200|既存home＋1本目のまま|
|`/robots.txt`|200|既存値と一致|

TLS検証を無効にしていない。最初のPython HTTP clientで403となった取得は本番成功確認に使わず、既存PowerShell HTTPS clientで取得し直して比較した。

## Canonical

release artifactはhome／1本目／除湿機がそれぞれ正式HTTPS self-canonical。productionは既存home／1本目が正常なまま。除湿機は404のため、新記事canonicalが本番に反映されたとは扱わない。

## Sitemap

release artifactはhome＋既存＋除湿機のvalid3URL。1本目のlastmod2026-10-06維持、新記事／homeは明示release日時。

productionは変更前の2URL sitemapと同一。まだ新URLは反映していない。public sourceやprivate releaseのsitemapを、実配信されたsitemapと混同しない。

## Robots

release artifactのindexabilityは正常、preview専用HTTP noindexはartifactに混入していない。production robotsは変更前と一致。新記事のindexabilityの本番確認は未公開のため未完了。

## Analytics

既存方式はCloudflare配信時の自動beacon挿入。manual beaconを追加せず、token／provider／設定も変更していない。

失敗後のhome・1本目の配信HTMLは各beacon1件。新URLの404 HTMLにもbeacon1件あったが、**新記事が配信・計測できた証拠とは扱わない**。成功公開後に新記事HTMLでbeacon1件を改めて確認する必要がある。即時集計値は取得・判定していない。

## Search Console

**NOT_RUN**。ユーザー指定の順序に従い、新記事のproduction HTTP／canonical／sitemap／robotsが通るまでGSC操作を行わない。既存sitemapの重複再送信、未公開URLのinspection／index requestなし。

既存observerは`webmasters.readonly`認証。復旧後は公式APIを優先してsitemap／URL Inspectionを取得する。[URL Inspection API](https://developers.google.com/webmaster-tools/v1/urlInspection.index/inspect)はindex状態の検査であり、live testやindex requestの送信APIではない。[個別URLのクロール要求](https://developers.google.com/search/docs/crawling-indexing/ask-google-to-recrawl)は公式UIの機能が利用できる場合だけ行う。以前blockedだったbrowser操作を繰り返していない。

公開直後の未index／未反映をSEO failureとしない。今回は公開自体が未成立であり、SEO／収益性の評価はしない。

## First Article Regression

deploy前・失敗後のsanitized snapshotsでhome／1本目／CSS／sitemap／robotsが完全一致。1本目のURL／title／description／canonical／本文／公開日／affiliate／disclosure／outgoing source／CSSを保持。デザイン・claim・Analytics挙動に今回の変更は配信されていない。

## Final Status

**NOT_PUBLISHED — AUTHORIZATION_REQUIRED。**

公開前ゲートはローカルとGitHub双方でPASS。production blocking issueは「既存記事破損」ではなく、deploy認証が成立しないこと。除湿機は未公開なのでPUBLISHED_OK／PUBLISHED_WITH_ISSUEとはしない。

必要なユーザー操作：GitHub repositoryの既存Actions Secret `CLOUDFLARE_API_TOKEN`を、既存ChoiceLen運用に必要な同等権限の有効な認証へ復旧する。失効理由やaccount対応も本人のCloudflare管理画面で確認する。password／tokenをチャットに貼らない。今回はtoken作成・変更を代理実行していない。

復旧後は公開承認範囲を維持して再開できる。実時刻で新metadata／release artifactを生成し直し、必要な時間差確認、pre-deploy gate、同一artifactのdeploy、production smoke、GSC、記録の順に進める。古い4:30のbuildを公開済み扱い・日時流用しない。

## Rollback Status

rollbackなし。deployが認証段階で失敗し、比較した既存production5ファイルも不変なので、rollbackの必要はなかった。新しいversionを配信済みとするreceiptは存在しない。

## Reproducibility

private `outputs/dehumidifier-production-release/`にapproval、release metadata、9file release、hash／pre-deploy gate、before／post-attempt HTTP、6画面QA／DOM、失敗receipt、GSC未実行理由、final manifestを保存。credential／token値は保存していない。準備したGSC private utilityは実行していない。

tracked差分：release metadata／HEAD guard、既存artifact検査、tests、既存workflowの検査step、内容不変の公開用source、承認済み生成site、当report。新しいscheduler／workflow／CF project、main merge、旧PR close／merge、release tagなし。siteを変更したため、legacyのhome／sitemapは旧hashを固定するfixtureとして保持し、加えて新release全体のmanifest／意味検査を実行する。1本目のfrozen保証を弱めていない。

External LLM APIs0／$0、楽天API0、GSC API0。Cloudflareは1回の失敗deploy workflow（service・account認証確認2件）で成功version0。GitHub Actions runは約70秒、実際の課金額は未確認で0と断定しない。新しい有料planは契約していない。

最終検査：`go test ./...` PASS、`go vet ./...` PASS、`git diff --check` PASS。GitHub runでも同検査とartifact validationがSUCCESS。生成物の再buildなしにdeployを試みたが、認証復旧待ちで停止した。


## Retry after credential update — 2026-10-21（最新結果）

**Latest Final Status: NOT_PUBLISHED。** 以下はユーザーが既存GitHub Actions Secretを更新したと連絡した後の再開記録。上記の旧失敗履歴・旧metadataは変更せず保持する。

### Fresh release metadata / artifact

- 新しい実時刻（Asia/Tokyo RFC3339）：`2026-10-21T21:50:47+09:00`。旧`2026-10-21T04:30:01+09:00`は再利用していない。
- 新release生成時の実HEAD／builder-content commit：`5b72394a29c6536578acb633e7185c9324dc230d`。
- source_checked：`2026-10-20`のまま。公開日とのcalendar差1日。検証から連続した再開であり、原本確認日を今日へ書き換えず、全面的な再評価もしていない。
- 監査済み本文・claims・4機種・2記事builder・共通CSS・1本目・広告方針の変更0。除湿機の価格／購入／affiliateリンク0を維持。
- private artifact：`outputs/dehumidifier-production-release/retry-20261021T215046/release/`。新規directoryへstage releaseで9ファイル生成し、既存RC／旧releaseをそのままdeployしていない。
- 検証済み集合SHA256：`9d9b98f6fcf36ea7dc334439f3e219604d0a913720f8b1e8dac07ec5823158cf`（同じsorted path TAB file hash LF方式、manifest自身も含む）。**実deploy済みhashではない。**
- artifact commit：`dcf8a5062aeff244031baaef373f58baecf03cbe`。前回attemptからのtracked変更はhome／新記事／sitemap／manifestのmetadata4ファイルだけで、builder／原稿は変更していない。

### Repeated pre-deploy gate

全PASS：release manifest／各file hash／metadata／現在HEAD、expected9files、第一記事の現在productionとの一致、3self-canonical、valid3URL sitemapと旧記事日付維持、robots／noindex、内部file／fragment／CSS、localhost／placeholder0、9source URL、22claim joinと全行の順序付き本文照合、4表、除湿機affiliate0。

27℃／60%・周波数・モード実動作・IJC／Panasonicのunknown・排水／自動停止・購入保留も生成HTMLで保持。既存の独立claim監査範囲を変える本文編集なし。

`go test ./...`／`go vet ./...`／`git diff --check`はローカルPASS。検証済みartifactを再生成せずsiteへ配置し、stage後のrelease検査・fixtureもPASS。今回もGitHub側で同3検査・exact artifact verification・pinned Wrangler install・dry-runがすべてSUCCESS。

### New workflow result

[New Actions run 39603116921](https://github.com/AkiraShimizu0/rakuten-product-search/actions/runs/39603116921)をworkflow_dispatchで新規実行した。**旧run39484227719のrerunではない。**

2026-10-21 21:53:24 JST、production deploy stepで失敗。Cloudflare API pathは`/accounts/[masked]/workers/services/choicelen-canary`。

返された本文：

```text
No access to the specified service.
```

**数値のCloudflare error codeは今回のWranglerログに表示されていない（unknown）。Wrangler process exit codeは1。** 旧失敗時の10000／9109を今回のcodeとして転載しない。

今回は`Invalid access token`という表示ではないが、これだけでtokenが正しいaccount／serviceに対して認証・認可されたとは断定しない。対象serviceへのtoken権限、tokenのAccount Resourcesと既存`CLOUDFLARE_ACCOUNT_ID`の対応を確認する必要がある。新project作成、route変更、token作成／変更、secret値の読み出し・表示は行っていない。

ユーザー指示に従い、**同じtokenで追加deploy／反復retryなし**。新version ID、成功deployment timestamp、実deploy済みartifact hashはない。

### Production / GSC / rollback

失敗後のHTTPS観測：home200、1本目200、新記事404、共有CSS200、sitemap200、robots200。home／1本目／CSS／sitemap／robotsは今回の試行直前snapshotと完全一致（edge beaconのみ除去、挿入改行正規化）。1本目のbody／canonical／source／広告／CSSに回帰なし。本番homeは引き続き既存1記事、sitemapも旧2URLのまま。

home／1本目beacon各1。新URLの404にもbeacon1件あったが、**新記事の配信／Analytics確認としては数えない**。新記事の本番canonical／公開日／本文／source／広告0の配信確認は未公開のため未完了。

本番成立の条件を満たさないため、GSC sitemap再送信／URL Inspection／index requestは未実行。用意したprivate API utilityは実行していない。以前blockedだったbrowser操作も再試行していない。

productionは不変なのでrollback不要・未実施。既存評価・過去decision、PR #25、main、Radar、Analytics設定を変更しない。

### Remaining action / provenance

必要なのは、既存Worker `choicelen-canary`を保持したまま、**既存accountと新tokenの対象・Workerアクセス権の対応をユーザーが確認すること**。tokenはチャットへ貼らない。現状のtokenで公開を反復せず停止する。復旧後の再開では、その時の実時刻を再取得して新metadata／artifact／gateから進める。

今回追加のLLM／楽天／GSC API呼出0、追加LLM API費用$0。Cloudflareはこの新規workflowのserviceアクセス失敗1回、成功deploy0。Actions実課金はunknown、新しい有料planなし。

private記録は上記retry folderのmetadata、release、before／post-attempt snapshots、validation、attempt／deploy receipt、GSC未実行理由、final manifest。旧private folder／旧失敗記録は上書きしない。Draft PR #26を維持し、この最新失敗を追記した。
