# Dehumidifier Publication Integration / Final Release Check

## Objective

**Decision: READY_TO_PUBLISH（2026-10-20 JST）。本番公開は未実施。**

PR #24で残ったR-T2「1本目を保持した2記事の本番生成と参照整合」を、隔離したローカルRCで確認した。これは次工程で明示承認を得て公開できる技術状態の判定であり、公開許可やコンテンツ評点の更新ではない。

## Input State

- 親：`dehumidifier-publication-readiness` / PR #24 / `f8452833402a9a05490a51e2c103dda92e1dc9f0`。別系統のbranchはmergeしていない。
- 既存完成稿：`outputs/dehumidifier-publication-readiness/article-final-draft.md`。実在を確認して再利用。SHA256 `b15b4b3f0648a965d63452a32a7a2087057174b5e65b7acfa36fe445a2533e47`。
- 監査済み22claim：`claims-v2.json`、SHA256 `f958bcb5d543169e9aff7923173be9cb512566f85f0ecbb03d0b3107190799ce`。原稿・監査入力のhash不一致や欠損時は停止する。
- D+7 CONTINUE、互換filter HOLD、除湿機market/content GO、SELECT_DEHUMIDIFIER、readiness CONDITIONAL、シュレッダーreserveは変更していない。本文・市場・judge・claimの再評価も行っていない。
- private protocolを実装前に固定。本工程の成果物・スクリーンショット・完成稿はignored `outputs/dehumidifier-publication-integration/` 以下へ保存し、Gitへ含めない。

## Existing Builder

従来の`build-choiceLen`は公開済みMarkdownの固定hash、固定URL、公開日、1記事用のhomeとsitemapに依存し、6ファイルを生成していた。固定hashは公開済み本文の回帰防止であり、削除・緩和していない。`build`と既存fixtureは維持する。

Workers Static Assetsの`site/`、trailing slash、404、domain設定はそのまま。Cloudflare Web Analyticsは配信時の自動挿入で、static source内に実行beaconはない。既存のJSON claim mapはtracking scriptではない。

## Minimal Integration Design

既存CLIへ`legacy`（既定）、`rc`、`release`の明示stageを追加した。2記事buildは従来の6ファイルを生成してから、2本目HTML・専用の小さなCSS・manifestを追加する。homeは記事リンク1件だけを追加、sitemapは3URLへ拡張する。CMS、DB、framework、font、画像、実行JavaScriptは追加していない。

`canary.RenderReference`は広告なしの資料記事を描画する入口。既存`Render`の広告開示要件と従来の生成結果は保持。2記事版は新規のignored `outputs/` 配下にしか出力できず、`site/`への出力・既存directory上書きを拒否する。完成稿はGitへ埋め込まず、明示指定されたprivate入力を要求する。

## Article Registry

小さな2件のregistryとmanifestでtitle、description、slug、canonical、source hash、資料確認日、公開／更新日、indexability、Analytics方式、表示順を記録する。実際のHTML整合は別途検査する。

|項目|1本目（維持）|除湿機（候補）|
|---|---|---|
|URL|`/articles/compact-air-purifier-placement/`|`/articles/winter-laundry-dehumidifier-conditions/`|
|title / H1|小型空気清浄機はどこに置く？机・床・棚を選ぶための条件表|冬の部屋干し用除湿機を、室温・モード・排水条件で比べる|
|description|小型空気清浄機の机・床・棚への配置を、公式説明書の寸法・吸排気・手入れ条件で確認する購入判断ガイド。|冬の部屋干し用除湿機4機種を、室温・モード・定格条件・排水と停止条件で照合する購入前の確認記事。|
|資料確認|2026-10-05|2026-10-20（既存完成稿の値）|
|公開日|2026-10-06|RCでは未設定。release時に明示入力|
|更新日|既存表示を維持|未設定。必要時だけ明示入力|
|一覧順|2|1|

一覧は除湿機→既存記事の明示順。未公開日の捏造、placeholder、`time.Now`による生成はない。新しいArticle structured dataは追加しない。

## First Article Preservation

1本目HTML、共通`assets/site.css`、404、robotsは、従来生成およびtracked `site/`（CRLFのみ正規化）と一致した。HTML全体一致のためtitle／description／H1／body／日時／claim map／リンク／広告開示／affiliate blockのsemantic変更もない。本文やサイトデザインを古い版へ戻していない。

1本目HTML SHA256：`93087787351a4736782fe8d3cefd2ee2b1b19bcc5cf48e99cd300ef3558d3767`。

共有CSS SHA256：`effeef4cb704cd13f490b3335df536d3e4a6a48d2ddc0ebfb26de4e93e30c3fe`。

## Dehumidifier Integration

監査済みMarkdownの全行について、本文・表cell・数字・型番・unknown・出典ラベルを順序付きでHTMLの可視テキストと照合した。22個のDD参照は非表示の`data-claim-ids`とJSON claim-source mapに保持し、9件の公式URLも各1件・完全一致。価格・購入URLは加えていない。

構造変更だけ：共通shell、予定canonical、description、資料確認metadata、4表のcaption／thead／列・行th／横scroll領域を付けた。DD参照の視覚上の非表示以外に本文を編集していない。DH-C7200ナイトの送風、F-YEX90D音ひかえめの除湿、IJC-P70の全運転温度未確認、Panasonic連続上限未確認、排水条件、27℃60%の前提は欠落していない。未確認事項を解決したとは扱わない。

専用CSSは2本目だけが読み込み、共通CSSや1本目に作用しない。図版・新claim・新しい監査スコアは追加しない。

## Homepage Integration

homeのtitleは`ChoiceLen — 商品選びの判断材料`、H1は`記事一覧`、descriptionは`メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料をまとめるサイト。`のまま。新記事一覧要素を除くと旧homeと完全一致する。カード・badge・CTA・copy変更・tracking変更なし。両記事からhomeへ戻れ、意味のない記事間リンクは加えていない。

## Canonical

home、1本目、2本目の3ページでH1各1・canonical各1。各canonicalはそれぞれのHTTPS公開予定URLに自己一致し、相互継承・localhost・preview URL混入0。2本目のURL候補は上表のtrailing slash付きpathに確定し、directory衝突なし。

RCのcanonicalは次回releaseの予定値であり、公開済みの主張ではない。RCはloopbackにのみ存在する。

## Robots

production候補HTMLのnoindex混入0。robotsは既存のAllow `/`と正式sitemap URLを維持。対してlocal serverは`127.0.0.1`のみで待ち受け、全responseに`X-Robots-Tag: noindex, nofollow, noarchive`、`Cache-Control: no-store`を付ける。preview用noindexは候補ファイルへ書き込まない。

## Sitemap

valid XMLでhome＋1本目＋除湿機の3URL。重複・旧記事消失・localhost0、canonical集合と一致。1本目のentryと`lastmod=2026-10-06`は維持した。

homeと2本目のRC lastmodは未設定。公開日時が未確定なため、homeの古いlastmodを新記事追加日として流用しない。releaseでは明示したpublished_at、更新がある場合はupdated_atをhome／新記事のlastmodへ反映する。

## Analytics

2026-10-20 20:54 JSTのread-only取得では、現行productionのhome／1本目は各HTTP200、Cloudflare beacon各1。既存static sourceに実行beacon0なので配信時自動挿入の構成として扱い、Cloudflare設定・token・Wranglerは変更していない。2本目も同じStatic Assets配信経路を使う設計で、手動beaconを重ねていない。

local previewは実行scriptをresponseから除去し、JSON claim mapは保持。さらにCSP `script-src 'none'; connect-src 'none'; object-src 'none'`を付与する。3ページのDOMで外部script0、候補ファイルは無変更。自動挿入されたbeacon付きfixtureも送信可能scriptが残らないことをテストした。

**2本目の実配信時beacon・計測受信は未公開なので未観測。** 次回公開後に1回だけ挿入されることを確認する。現在の設定や計測値は変更・再評価していない。

## Affiliate Behavior

1本目の正式affiliate URL／広告開示／`rel="sponsored"`1件はHTML一致により保持。除湿機は価格・購入・affiliateリンク0、sponsored0。メーカー出典9件を購入リンクへ変換していない。広告なしの2本目へ広告利用を示す偽の本文開示は追加しない。共通footerの広告についてはhomeの既存anchorへ戻る。

## Release Candidate Build

正式に検査したprivate出力は`outputs/dehumidifier-publication-integration/release-candidate-final/`。9ファイル：home、2記事、共通／専用CSS、404、robots、sitemap、build-manifest。builder commit：`a4cf59f03ee7d969e5b740c75b67e25c2e147027`、version：`two-article-v1`。

manifestにsource／claim hash、registry、builder version／commit、stage、8出力hashを保存し、manifest自身は自己参照しない。manifestは最後に書く完了marker。別directoryへ同一入力で再生成し、manifestを含む9ファイルのhashがすべて一致した。先行の動作確認用directoryは正式RCに使わない。

## Automated Validation

- 従来6ファイルfixture、2記事9ファイルfixture、生成再現性、1本目／home差分範囲。
- 3canonical、3URL sitemap XML／lastmod、内部link／fragment／CSS存在、HTTPS、localhost0、noindex0。
- 22claim ID／source join、全source prose順序照合、9source URL一致、2本目affiliate0。
- 不正canonical、偽日付、重複・欠損sitemap、失われた数字・出典、未定義fragment／CSS、追加の広告・購入リンク、script混入、claim欠落／unsupported／重複を拒否するmutation tests。
- RCのrelease metadata拒否。release必須値欠落、JST以外、不適切な過去日、更新日逆転、資料日／commit不一致を拒否。release日付の動作確認はsynthetic fixtureのみで、実稿のrelease buildは実行していない。
- local候補の8公開対象fileをHTTP取得し全200。候補全HTMLの内部参照はfile／fragmentまで確認。外部公式リンクは保存済み監査とURL一致を確認し、追加crawl／affiliate clickなし。

## Desktop QA

in-app browser、1440×1000、loopback HTTP。home、1本目、除湿機を実表示しスクリーンショット保存。白背景・system sans・1カラム、タイトル1件、記事一覧2件、footer、出典、表を確認。除湿機body幅700px、H1 32px。表4件は700px領域内に収まり、横方向のpage overflow0。1本目に表示変更なし。

## Mobile QA

390×844のiPhone相当viewport。3ページを実表示。pageのscrollWidthとclientWidthは各一致し、水平page overflow0。縦scrollbarを除く記事clientWidth375px、本文335px、H1 22px。4表は335pxの領域内で640pxの表だけを横scrollでき、captionで操作を示す。

表2をキーボードでfocusしてArrowRightを送り、scrollLeft22pxを確認。新記事の出典9リンクは各44px高、homeの両記事リンクは50px高。長い型番、H1の自然な改行、sources／footerまで確認した。Desktopのリンク遷移とMobileのEnterによるhome↔記事遷移も確認。Mobileでclick操作の結果が変わらない場面は直接URL／キーボードで表示を検証し、未確認のタッチ実機テストを行ったとは扱わない。実機iOS／VoiceOver試験は未実施。

## Regression Results

- 1本目HTML／共通CSS／robots／404：変更0（改行正規化後）。
- home：記事1要素の追加以外0。1本目sitemap entry：変更0。
- 原稿とclaims：監査済みhash一致。本文の追加・削除・数値改変0。
- tracked `site/`、`content/published/`、Wrangler、workflow、既存評価文書：親branchとの差分0。
- Gate／reranker／Family Resolver／Radar／scheduler変更0。

## Remaining Issues

blocking technical issue **0**。次は公開の実行条件であり、今回のtechnical checkを未完了扱いにする問題ではない。

|区分|残条件|処理|
|---|---|---|
|公開実行必須|ユーザーの明示公開承認|今回取得・実行しない|
|公開実行必須|実公開timestamp・release commit|release buildへ明示入力し、未設定ならfail|
|公開後確認|2本目HTTP／HTTPS／beacon1／計測・参照|deploy後に新ページとして確認。現時点で観測済みとしない|
|維持条件|IJC温度／定格mode、Panasonic連続上限unknown|既存完成稿の限定・購入保留を維持。技術工程で解決したとしない|
|将来の別工程|商品価格／広告導線を加える場合|商品条件・正式URL・claim差分を再確認。今回許可外|

## Final Decision

**READY_TO_PUBLISH**。固定protocolに従い、2記事build、1本目回帰、2本目HTML、canonical／sitemap／内部link／robots／noindex、Desktop／Mobile表示、再現性、testsを満たした。過去readiness CONDITIONALの履歴を上書きせず、R-T2を今回技術検証したという追記である。

公開承認、SEO成功、収益性、STRONG GOへの昇格を意味しない。Cloudflare deploy、GSC、main merge、tag、sitemap本番更新は行っていない。

## Release Requirements

次工程のみ：実公開承認を受け、実際にdeployするbuilder commitと実公開timestampを確定し、private JSONに`published_at`（RFC3339、`+09:00`）、`source_checked`（現稿2026-10-20）、`deployment_commit`（builder commitと一致）、必要時のみ`updated_at`を渡す。古い資料のままの長期延期なら先に資料時点を再確認する。このコードは未来の資料確認を自動的に行わない。

`-stage release -release-metadata <private-json>`と明示入力一式で新しいignored directoryへ再生成・検証する。RCを公開日未設定のままuploadしない。Wranglerのasset directoryへ持ち出す操作とdeployは、次工程の管理対象。URL／canonicalは上記確定値、lastmodは明示日時、1本目公開日は維持する。公開後にHTTP／canonical／sitemap／robots／source／Analytics1beaconを確認する。

## Reproducibility

必要なprivate入力2件は新規cloneに含まれない。指定hashの完成稿・監査ファイルを復元できない場合は停止し、本文の推測再作成やfixtureの代用は禁止。synthetic fixtureは本文を含まないため、Gitだけで一般の回帰testsは実行できる。

RC例（repo rootから、明示commitは上記builder commit）：

```text
go run ./cmd/build-choiceLen -stage rc -second-source outputs/dehumidifier-publication-readiness/article-final-draft.md -second-claims outputs/dehumidifier-publication-readiness/claims-v2.json -builder-commit <builder-commit> -out outputs/dehumidifier-publication-integration/<new-directory>
go run ./cmd/preview-choiceLen -dir outputs/dehumidifier-publication-integration/<new-directory> -port 8792
```

既存のlegacy生成方法は維持。private記録：protocol、production-observation、RC／repeat、validation-report（hash／SEO／HTTP／regression）、QA PNG／DOM測定、final manifest。Desktop／Mobile合計9枚のPNGを保存。viewportは確認後にreset。

`go test ./...` **PASS**、`go vet ./...` **PASS**、`git diff --check` **PASS**。Jev／Claude／楽天API新規呼出0、追加LLM API費用$0。独立judgeの追加実行なし。Draft PRは本工程のcode／tests／reportだけを親branchに対して提示し、生成物・未公開本文・screenshots・credentialは含めない。
