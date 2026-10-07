# D+7 evaluation protocol — frozen before outcome inspection
Version: day7-v1。固定日: 2026-10-08 Asia/Tokyo。観測日: 2026-10-13。
対象Canary1記事。実データを見てルールを変更しません。

## Window / provenance
公開基準日2026-10-06 JST。原則公開時刻以降〜2026-10-12 23:59 JSTの利用可能なデータを10月13日に取得。公開前は除外。
サービスのtimezone/集計遅延が異なる場合は実際のwindow、timezone、latest available dateを保存。窓の違う値を混ぜず、遅延日は0で埋めない。
取得ごとchecked_at/source/export/window/filter/aggregation/取得可否を保存。秘密値・個人識別情報はpublic repoへ保存しない。トップと記事、site totalは分離。

## Data
- Search Console: indexed state、indexability、crawl/canonical、sitemap status、impressions/clicks/CTR/average position、queries/pages。非表示queryがあるためquery合計≠page合計を許容。
- Cloudflare Web Analytics: page views、サービス上のvisits/unique相当の名称と定義、top pages、利用可能なreferrer。取得不能はunknown。設定検証アクセスを可能な範囲で注記、自然流入扱いせず分離不能はunclassified。
- Rakuten Affiliate: clicks/orders/commission JPY、period/反映遅延/帰属範囲。記事単位分解不能ならaccount aggregateと明記。正規reportが0を示す場合だけ0、未取得はunknown。
- Radar: scheduled/attempted/successful/failed/partial runs、dedup後snapshots/events/drop candidates、last successful run、Task Scheduler status。run manifestsから再現。expected daily slotsと欠測を区別。

## Pre-registered rules
複数問題は併記、CONTINUEで隠さない。
1. critical production issueは最優先の技術調査。収益判断とは分離。
2. INVESTIGATE_INDEXING: request後もnot indexed、sitemap/canonical/robots不具合、crawl error。記事追加を先にしない。
3. DATA_PIPELINE_ISSUE: 設定不備でSC/Analytics/Affiliate等が測定不能。report delay/低volumeと区別しunknownの理由と復旧策を記録。
4. CONTINUE: page indexable、measurement operational、critical issueなし。未indexなら2も併記。impressions/clicks/orders=0だけでNO-GOにしない。
5. NO SEO PROFITABILITY DECISION: D+7だけでSEO成功/失敗、profitability有無を確定しない。

## Query interpretation
article intent一致、placement intent、product-specific/informationalを分類。少数/非開示queryから一般化しない。queryを見てprotocolを書き換えない。

## Next branch options
A Indexed + impressions >0: compatible-filter article validation / dehumidifier external market validationを次候補に検討。
B Indexed + impressions=0または少量: wait/continue observation、intent review。市場検証は可能だが量産しない。「ほぼなし」は説明的観察で新しい数値GO閾値にしない。
C Not indexed: technical/indexing investigation最優先。追加記事は後。
測定不能ならpipeline recovery優先。いずれも自動公開を認可しない。

## Caveats
D+7前のUI/editorial wording変更はmeasurement-logに記録済み。design間の因果比較はしない。D+14/28/56は既存templateで継続可能。今回D+7実績取得/判定なし。
