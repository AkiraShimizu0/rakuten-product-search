# Dehumidifier Content Prototype Validation

## Objective

**Decision: GO — content prototype validation only; not STRONG GO or publication approval.**

「冬も部屋干しに使う除湿機を、室温・運転モード・定格条件・排水条件で比較する」記事を非公開Markdownで検証した。定格能力の順位ではなく、必要な使い方から購入前の確認・除外・保留へつなぐ比較は構成できた。独立した匿名要約評価ではoverall 5.00、最良競合4.00、差+1.00。ただしevidence quality 4.00は事前固定したStrong GO条件4.30を満たさない。

実施日2026-10-16 Asia/Tokyo。base `dehumidifier-market-validation` / `1d5f4e7c15535185edf0a71c045b1765cab3e0e1`、parent Draft PR [#19](https://github.com/AkiraShimizu0/rakuten-product-search/pull/19)。市場GO 29/35、Day4/Day5、D+7、互換フィルターHOLDの記録を変更しない。今回の結果はSEO順位・需要・収益性の予測ではない。

## Frozen Scope

モデル・19項目・尺度・判定閾値・競合3ページをprototype評価前にprivate `outputs/dehumidifier-content-validation/protocol.md`へ固定。seed `20261016`、rubric `day5-content-v1`、judge prompt v1。方式解説・おすすめ10選を中心にしない。実機試験、独自の冬能力推定、年間電気代計算はしない。

4機種は入れ替えず全て残した。モデルidentityはメーカー公式型番で確認し、HOLD中のFamily Resolver v2へ依存しない。

## Fixed Models

|モデル|主な原本|比較での役割|
|---|---|---|
|シャープ CV-T71|SM 説明書、SS 仕様|除湿と衣類消臭、ホース排水、解除可能な自動停止|
|Panasonic F-YEX90D|PM 説明書、PS 仕様、PC 比較表|音ひかえめでの除湿、連続排水非対応、自動停止と連続設定の区別|
|cado DH-C7200|CM 説明書、CS 製品情報|ナイト送風、ホース固有条件、全mode約12h停止|
|アイリスオーヤマ IJC-P70|IM 説明書、IS サポート|効率説明と運転限界の区別、24h無操作停止、unknownの保持|

## Buyer Problem

読者が判断できることを1文で定義する：**自分の冬の室温、夜間に必要な動作、水捨て・排水経路、運転時間をもとに、各型番の何を原本で確かめ、どの条件が未解決なら購入を保留するかが分かる。**

最も強い問題は「夜間に除湿したいのに、名称だけで送風モードを選ぶ」「連続排水できるから止まらないと考える」こと。方式・L/日の一般解説だけでは、この型番別の違いが見えにくい。50/60Hz、住宅構造、定格と実使用の混同も扱う。

## Evidence Matrix

private `evidence-matrix-v2.json`に4機種×19項目=76セルを保存。各セルはmodel/field/value/source IDs/page/status/解釈/比較方針を持つ。source manifestとclaim joinはGo検証PASS。

内訳：**verified 64、qualified 10、unknown 1、not applicable 1**。全セル記入は全仕様確定を意味しない。qualifiedにも未確認の部分項目がある。一次資料9件（4原本PDF＋5公式Web）、別に再利用した楽天API証拠1件。HTML比較表の再確認は元checked_atを保持し、`rechecked_at 2026-10-16T12:44:53+09:00`、HTTP200、hashを別記した。PDF4本はHTTP200で取得、SHA256を保存。既存原本の要確認部分だけ再利用・補完した。

以下は要約。数値の列は左50Hz／右60Hz。略記だけを施工・安全説明として使わない。

|項目|CV-T71|F-YEX90D|DH-C7200|IJC-P70|
|---|---|---|---|---|
|方式|コンプレッサー|エコ・ハイブリッド（コンプレッサー＋空冷除湿）|コンプレッサー|コンプレッサー|
|定格L/日|6.3／7.1、除湿強|7.8／8.5、除湿強|6.5／7.5、除湿強|5.9／7.3、定格mode名unknown|
|定格試験温度|27℃|27℃|27℃|27℃|
|定格試験湿度|60%|60%|60%|60%|
|周波数|50／60Hz別記|同左|同左|同左|
|適用畳数|木造8／9、鉄筋16／18|木造10／11、プレハブ15／16、鉄筋20／21|木造8／9、鉄筋16／19|木造7／9、プレハブ11／14、鉄筋15／18|
|衣類乾燥|自動・弱・強。消臭は別|標準・音ひかえめ・速乾、連続設定切替あり|衣類乾燥自動等。自動約5h後は送風|専用modeはunknown、衣類への風向調節案内|
|通常除湿|自動・弱・強|自動・弱・強|自動・弱・強|弱・強・おまかせ。湿度で送風切替|
|低温挙動|約18℃未満で霜取りの場合|低温低湿で量低下、約10℃以下等で風量増の場合|約15℃以下で霜取り、約5℃以下で圧縮機停止送風|低温低湿で量低下、霜取り中は除湿停止|
|ナイト／静音|専用ナイトunknown。除湿弱と消臭弱は別|衣類音ひかえめは除湿。保護制御あり|ナイトは送風だけ|専用ナイトunknown。弱は除湿|
|消費電力W|除湿強27℃60%で205／210|衣類標準・速乾175／185、音ひかえめ160／170、除湿弱155／165等|除湿強27℃60%170／187、衣類180／200、ナイト22／26|定格180／190。mode・温湿度別条件unknown|
|タンク|約2.5L、自動停止量|満水ライン約2.4L|約3.5L、公称|約2.5L、公称|
|満水停止|あり。衣類消臭は別扱い|あり。タイマーは進む|あり|あり、タンク不在でも停止|
|連続排水|対応|非対応（PC型番列で確認）|対応|対応|
|排水条件|内径15mm、外径21mmまで、下り勾配、タンク必要、凍結・水没回避等|非対応につきN/A|内径12mm、外径16mm、ホース口より低く、先端20cm以上下、タンク必要等|内径15mm、排水口より高くしない、水没不可、タンク必要等|
|運転温度|約4～38℃|約1～40℃、湿度約20%以上|約5～35℃。約5℃以下・35℃以上は圧縮機停止|全範囲unknown。15～30℃は効率設計の説明|
|自動停止／timer|約12h停止は解除可能、timer2/4/6h|衣類自動停止設定は最長約15h。連続設定上限unknown。timer2/4/8h|全mode約12h停止、timer1/4/8h|24h無操作停止、timer2/4/6h|
|手入れ|タンク週1、フィルター2週等|タンク月1、吸込フィルター2週等|内部乾燥週1推奨等|本体・吸気・タンク月1等|
|測定規格|面積JEMA、能力の規格番号unknown|面積JEMA、衣類乾燥試験JEMA-HD090:2017。能力規格番号unknown|能力27℃60%明記、規格名unknown|面積JEMA、能力mode・規格番号unknown|

### Source manifest

原本PDFのpageは紙面番号。cadoはPDFファイルページが紙面+1となる点を記録した。原本全文・抽出・ページ画像はignored privateに置く。

- SM: [CV-T71説明書](https://jp.sharp/restricted/support/manual/dehumid_con/cvt71_mn.pdf)、主に3、7～13、18p。
- SS: [CV-T71仕様](https://jp.sharp/joshitsu/products/cv-t71/spec/)。
- PM: [F-YEX90D説明書](https://panasonic.jp/content/dam/panasonic/jp/ja/pim-assets/support/manual/000/000/004/397/657/000000004397657/F-YEX90D_web.pdf)、主に10、19、23p。
- PS: [F-YEX90D仕様](https://panasonic.jp/joshitsu/products/F-YEX90D/spec.html)。
- PC: [Panasonic機種比較](https://panasonic.jp/joshitsu/comparison.html)。列順200D／120B／90Dと連続排水の●／●／－を照合。
- CM: [DH-C7200説明書](https://cdn.cado.com/manual/manual_safety_dh-c7200.pdf)、主に8～11、16～17、25、30～31p。
- CS: [ROOT 7200製品情報](https://cado.com/pages/root-7200)。
- IM: [IJC-P70説明書103701](https://www.irisohyama.co.jp/products/manual/pdf/103701.pdf)、主に2、5、8～10、15p。今回公式サポートから辿れた原本を採用。
- IS: [IJC-P70サポート](https://www.irisohyama.co.jp/products/support/4967576654197)。

## Condition Normalization

直接照合できるのは、型番ごとの公表数値・設定の有無・動作条件と、自分の設置条件である。L/日は27℃60%・地域の周波数・明示modeを保持した公称参照値として並べる。全規格・modeの一致や冬の実性能順位まで比較できるとはしない。

**NOT DIRECTLY COMPARABLE**：冬の実際の乾燥速度、異なるmodeのW値、異なる周波数・住宅区分の畳数、方式名からの冬能力、タンク公称量からの水捨て間隔、メーカー別の試験条件が異なる乾燥時間。騒音・年間電気代・健康効果の順位も作らない。独自計算は0件。

## Low-temperature Behavior

温度範囲内で運転できることと、定格能力を維持することを分けた。特にIJC-P70の15～30℃を使用禁止境界に読み替えない。F-YEX90Dは低温時に保護の風量増があり、音ひかえめの名称を全条件での静かさ保証にしない。方式名だけで冬向き・不向きを断定しない。

## Operating Modes

最も具体的な追加価値は、DH-C7200のナイトが送風のみなのに対し、F-YEX90Dの衣類音ひかえめは除湿する点。CV-T71の除湿弱と衣類消臭弱も別機能。IJC-P70の弱は除湿するが、専用ナイトの有無を補完しない。22/26Wの送風と160/170Wの除湿を省エネ順位にしない。

## Drainage Conditions

F-YEX90Dはホース必須という条件の候補から外せる。他3機種でもホース径・高さ・勾配を流用せず、タンク装着・凍結・水没・点検と停止を原本で追う。連続排水は無人永久運転の許可ではない。CV-T71の自動停止解除方法があることも、無人運転推奨へ変換しない。取説の全安全項目や設置図の代替にはしない。

## Rakuten Mapping

追加取得なし。PR #19のAPI観測 `2026-10-15T14:58:07～12Z`（23:58 JST）を再利用。商品title/caption・source_id・shop・API itemPrice/min/max/basefield・availability/postage flag・selected variant unknownをprivate JSONで保持。以下は価格順位でも現在の注文価格でもない。

|公式型番|楽天shop / source_id|API価格|照合上の制約|
|---|---|---:|---|
|CV-T71|XPRICE / a-price:10849817|23,800円|min23,800～max27,500。選択SKU価格unknown、送料別flag、availability1|
|F-YEX90D|天六なごみ / nagomi-pharmacy:10024835|80,080円|選択SKU未確定、送料別flag、availability1|
|DH-C7200|roomy / roomy:10015559|59,950円|選択色/SKU unknown、送料込flagの地域条件unknown、availability1|
|IJC-P70|ウエノ電器 / u-denki:10027672|19,800円|availability0。当該出品を購入可能候補としない|

中古表記がないことは実物新品の独立証明ではない。source_idとseller独自SKU／選択variantは別物。別型番・セット・可変SKUの最安欄を一律の新品価格として使わず、クーポン後・ポイント後価格を採用しない。通常商品URLの照合までで、affiliate発行なし。条件確認後の通常商品ページ導線は自然だが、購入候補が全4出品揃った状態ではない。

## Prototype Structure

非公開原稿：`outputs/dehumidifier-content-validation/prototype-v2.md`。v1と監査指摘を残し、採点前にv2へ修正した。採点後の書き換え・再採点はしない。

1. 室温・必要な夜間動作・水捨て条件から始める短い導入
2. 「L/日」は冬の部屋の水量ではない：定格条件表
3. 室温が下がったとき何が変わるか：低温条件
4. 夜のmode名だけで選ばない：送風と除湿の具体差
5. ホースで排水できても運転は別条件で止まる：排水・停止表
6. 自分の条件から購入前の確認欄を決める
7. 楽天出品へ進む前の型番・SKU・価格snapshot
8. 調べた範囲とunknown

### Decision Table

|自分の必須条件|確認する対象|満たすと分からなければ|
|---|---|---|
|冬の無暖房室|実際の室温と型番別停止・霜取り条件|方式から能力を推定せず候補保留|
|夜間にも除湿|modeが本当に除湿するか、保護制御・停止|DH-C7200のナイトをこの目的に使わない|
|ホース排水必須|対応と径、排水先までの高さ・経路|F-YEX90Dを外す。他機種も経路不明なら保留|
|長時間の運転|満水以外の自動停止・無操作停止|タンク容量だけで解決したとしない|
|狭い設置場所|本体寸法＋吸排気・手入れの原本図|このprototypeだけで置けるとしない|
|仕様数値の比較|周波数・mode・木造/鉄筋の列|異条件の順位を作らない|

## Claim Audit

掲載claim 22件：verified 13、qualified 9、unsupported 0。除外候補5件（冬最強／省エネ・静かさ順位／健康等保証／乾燥時間・費用保証／無人永久運転）はremoveで保持し不掲載。

|claim IDs|内容|根拠|
|---|---|---|
|DD01～03|定格数値と比較可能性の限界|SM/PM/CM/IMの定格表・条件|
|DD04～07|低温条件とIris unknown|SM3/7、PM19/21、CM25、IM2等|
|DD08～11|夜間・弱・消臭の実動作|CM8、PM10/23、SM8/9、IM8/9|
|DD12～13|W比較の限界とタンク表現|PM23、CM30/31、SM18、IM15|
|DD14～17|連続排水と機種別ホース条件|PC、SM13、CM16/17、IM10|
|DD18～19|自動停止の設定・例外・unknown|SM12/13、IM5/10、CM9/11、PM10/19|
|DD20|編集上の条件照合・保留手順|上記原本からの限定したinference|
|DD21|楽天価格・SKU・availabilityの限定|再利用楽天API、manufacturer資料扱いにしない|
|DD22|連続排水を無人放置許可にしない|SM/CM/IMの停止・排水・点検条件から限定したinference|

独立したsource-backedレビュー1では**major1、minor1**を発見。DD19の「衣類運転は最長約15h」は連続設定にも適用する過剰一般化だった。PM19の自動停止設定に限定し、PM10の連続設定上限はunknownへ修正。DD21のsource_typeをメーカー資料から楽天APIへ訂正。修正を独立再確認したレビュー2の残存は**critical0 / major0 / minor0**。初回指摘を消さずprivate保存。これは限定原本監査での未発見件数であり完全な安全認証ではない。表の全セルも監査対象とし、機械join検査で出典不明・重複・orphanなし。

## Competitor Comparison

固定した3ページは2026-10-16に現行内容を確認。過去市場調査のsnapshotを上書きしない。原文全文は転載せず、同じ6パート（対象、能力、冬/夜間、排水、判断支援、販売/限界）に要約した。

|ページ|強み|今回scopeで残る追加確認|
|---|---|---|
|[どれ買う？ 除湿機](https://dorekau.net/dehumidifiers/)|12機種、用途・予算・診断と比較で早く絞れる|型番別の低温・送風mode・排水停止条件の原本照合|
|[LiviqLens 能力・排水・湿度設定](https://liviqlens.com/dehumidifier-capacity-drain-humidistat/)|洗濯量と運転時間、タンク運搬と排水経路を先に整理|掲載商品欄の尺度が混在。各モデルのmode・停止条件と原本対応|
|[Panasonic 選び方](https://panasonic.jp/joshitsu/select.html)|6軸、試験注記、方式・衣類量・仕様への導線|自社中心。複数メーカーの夜間動作と排水可否を一度に追う比較|

最良競合はPanasonic公式ガイドだった。一般記事にも具体的な排水・条件判断があり、単に表を増やすことを差別化とはしない。今回の差は「機種固有の必要条件により候補を外す・保留する」接続にある。

## Independent Evaluation

fresh isolated judge 3名、履歴・市場GO・自他ラベル・サイト/メーカー/型番名・他judge回答を非提示。同一プロンプト、1回ずつ、4入力/4出力。匿名要約の文字数は自稿662、競合639/651/659で、全文量の差を抑えた。seed付きGo permutationでR01～04へ割当。全12行・6軸整数1～5・ID/notesの完全性を検査してからkey join。入力/prototype/protocol/証拠/audit/prompt hashを照合し、変更時と既存結果への上書きを拒否する。

|対象|判断有用性|証拠品質|比較深度|行動可能性|明瞭さ|overall|
|---|---:|---:|---:|---:|---:|---:|
|ChoiceLen prototype|5.00|4.00|4.67|5.00|4.00|5.00|
|Panasonic公式ガイド|4.00|4.00|3.33|4.00|4.00|4.00|
|LiviqLens|4.00|3.00|3.00|4.00|4.00|3.33|
|どれ買う？|3.33|3.00|3.00|4.00|4.00|3.00|

|judge|自稿overall|そのjudgeの最良競合|差|比較|
|---|---:|---:|---:|---|
|1|5|4|+1|win|
|2|5|4|+1|win|
|3|5|4|+1|win|

overallの勝ち/同点/負けは3/0/0。agreementは方向の一致のみ。小標本・同一providerであり、統計的有意性、独立provider間再現性、一般読者の実証とは言わない。overallは各judgeの独立評点で、5軸平均ではない。

**尺度上の限界**：要約は調査者が作成した。匿名化しても内容から自稿を推察し得る。元ページの全リンク先や全文の読みやすさ・図・導線は評価していない。証拠品質は要約で表現された品質で、judge自身の一次資料認証ではない。別途原本監査と分離した。いずれも条件表の密度・未確認規格/動作・実測なしを制約として指摘。得点はこのscopeでのstructured-summary比較に限る。

### LLM provenance / cost

- Provider: OpenAI/Codex、exact model IDは環境から取得できず`unknown`。judge prompt v1 / rubric day5-content-v1 / seed20261016。
- 独立judge実行3、各input4/output4、出力失敗0。claim reviewerは別の独立セッションで初回＋修正確認。
- 新規有料Jev/Claude等の外部API request **0**、その直接追加費用 **$0**。Codexのinput/output tokens、追加請求・按分費用は**unknown**であり、LLM全体の請求額を$0と断言しない。private metadataに記録。

## Risks

1. Irisの全運転温度・定格mode、Panasonic連続設定の上限、全機種の定格規格番号の一致が未確定。必要条件次第では保留になる。
2. 冬の実際の乾燥速度・静かさ・効率はunknown。公称仕様から性能優劣を出せない。
3. 楽天は過去snapshot、可変SKU・新旧/新品状態・地域送料・availabilityに制約。Iris例は購入可能でない。
4. 原本改訂・mode名・販売SKUの更新リスク。無人運転や設置安全の一般化は禁止。
5. 要約比較の上昇幅を、完成記事全文や読者実測の優位へ一般化しない。

## Decision

**GO**。buyer problem、4モデルのevidence matrix、条件照合と購入停止表は成立。seller-only証拠は価格補助に限定され、中心はメーカー原本。claim修正後critical/major0、自然な型番照合後の商品導線を設計できた。

既存Day5ゲートをそのまま適用：Strong GOはoverall≥4.3、最良競合差≥0.30、全3judge非劣位、evidence≥4.3、critical/major0。GOはoverall≥4.0、差≥0、2judge以上非劣位、critical/major0。今回はoverall・差が高くても**evidence4.00 < 4.30**のためStrong GOにしない。定格/unknownの制約と要約比較の限界を併記し、29/35の市場GOから自動昇格させない。

## Reproduction / Validation

以下はローカルprivate artifactを使用する。judgeはツールが生成せず、独立回答を保存する。既存出力がある場合は拒否し、元データを消してやり直さない。

```text
go run ./cmd/compare-content -mode prepare -dir outputs/dehumidifier-content-validation
# Independent judges write judge-1.json ... judge-3.json, with metadata.
go run ./cmd/compare-content -mode analyze -dir outputs/dehumidifier-content-validation
go run ./cmd/compare-content -mode audit -dir outputs/dehumidifier-content-validation
go test ./...
go vet ./...
git diff --check
```

新Go utilityはDay5のValidateRatings/Decisionを再利用し、既存の評価・Gate・reranker・Resolver・Radarを変更しない。fixtureでmissing judge/invalid score/duplicate・unknown ID/key join、deterministic aggregation/seed、frozen input改変、output overwrite拒否、orphan/unknown source/duplicate claim/欠損セル/unsupported使用を検査する。既存Day5判定境界テストも維持する。

2026-10-16の実行結果：`go test ./...` PASS、`go vet ./...` PASS、`git diff --check` PASS。baseに対する`site/`・`content/published/`・`wrangler.jsonc`・既存市場検証文書の差分は0。

非公開成果物：protocol、manifest v1/v2、原本PDFと抽出・対象ページ画像、matrix v1/v2、Rakuten mapping、claims v1/v2、prototype v1/v2、claim review1/2、匿名要約/key/hash、judge3名のraw回答とmetadata、score summary、integrity、実行provenance。raw商品データ・PDF・スクリーンショット・出力はGitに追加しない。Gitの履歴に今回の文書・utility/testsのcommitを保持し、private provenanceにもcommit SHAを追記する。

## Next Step

ここで停止。**STRONG GOではないためpublication candidate preparationを提案しない。** 将来このscopeを再検討するなら、未確認の原本条件を解決し、要約評価では未検証の本文全体の読みやすさを別途確認する。現在の評価を救うための再採点は行わない。

`site/`、`content/published/`、既存評価文書、Workers/Cloudflare設定は変更しない。公開・affiliate発行・sitemap/GSC送信・deployは0件。
