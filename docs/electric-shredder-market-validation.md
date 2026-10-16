# Electric Shredder External Market Validation

## Objective

**Decision: GO — 28/35。STRONG GOではない。** 電動シュレッダーには、メーカー原本で説明できる購入ミス、複数メーカーの商品、楽天の商品導線がある。一方、定格/最大・連続/休止・静音条件は既存記事にも具体的な説明があり、これらを追加するだけでは差別化が弱い。次の検証候補は**在宅でまとめ処理する際の手差し／自動給紙と、紙・金具の条件、運転・冷却・復帰条件の照合**に絞る。

調査日2026-10-16 JST。公開記事・content prototype・HTML・affiliate URLは作成しない。本番・既存評価・scheduler・API評価設定を変更しない。

## Prior Evidence

Category Discoveryのreview-order Claude meanは62.75、price-stratifiedは60.20（差−2.55）。既存Sampling Bias protocolではROBUST、優先順位2位。ただし条件付きsampleの内部機会評価であり、市場需要や収益性ではない。Family baselineはシュレッダーgold subsetでprecision .2432 / recall .5294 / F1 .3333という制約がある。v2はHOLDのまま採用しない。

既存判定：D+7 CONTINUE、互換フィルター追加検証HOLD、除湿機market GO、除湿機content GOを維持。今回branchは`integration-v0.1.0`の`0dd05e4345f6e06620fae0bfc8fb06a4cc4bc78a`から直接作成した独立branchで、除湿機PRの上には積まない。

## Fixed Method

既存Day4 market rubric・thresholdをそのまま使用し、private protocolへ記録した。初期の方向確認検索はprotocolファイル作成より前であり、「すべての検索前に新規事前登録した」とは扱わない。閾値は既存固定値で、今回の結果による変更は0件。

- A～G各0～5、合計35。Strong GO≥27かつA/B/C≥4・F≥3。GO≥23かつA/C/F≥3。HOLD19～22又は重要情報不足/borderline。NO-GO≤18又はhard fail。Goの既存`market.Assess`を直接再利用。
- 3リスク各0～5、高いほどリスク。主scoreから減点しない。
- 12 queryを記録して検索し、公式型番・原本取得用queryを追加。検索ツールの結果と関連ページを読む方式。地域/個人化を固定したGoogle上位10位の観測ではなく、順位・検索volumeはunknown。
- genre303156を既存Discoveryから再利用。標準順2page＋レビュー件数順1page、30件/page、availability=all。追加は初期sampleにない16MCの型番照合1requestのみ。母集団へ混ぜない。

## Search Intent

|Query|分類|確かめたいこと|
|---|---|---|
|シュレッダー 家庭用 選び方|commercial investigation|量・場所・対応する紙|
|電動シュレッダー おすすめ|commercial investigation / transactional|候補比較と購入先|
|クロスカット マイクロカット 違い|informational / commercial investigation|名称と実寸法|
|連続使用時間 休止時間|specification check / commercial investigation|一度の作業と冷却待ち|
|静音 dB 測定条件|specification check / commercial investigation|空転/負荷、測定条件|
|A4 最大 定格 枚数|specification check|紙質と数値の定義|
|ホッチキス 対応 針|specification check / transactional|号数・紙の束の条件|
|カード CD 対応 投入口|specification check / transactional|素材・別口・1枚ずつ|
|オートフィード 紙 条件|commercial investigation / specification check|セット枚数と手差し、折り/綴じ紙|
|ゴミ箱 容量|commercial investigation|ごみ捨て・処理量の見方|
|詰まる 止まる|troubleshooting|過負荷・過熱・再開条件|
|家庭 会社 違い|informational / commercial investigation|ラベルではなく運転量・条件|

結果には選択ガイド、仕様、購入先、トラブル対応が現れた。検索結果の存在、レビュー数、API reported totalから需要量を推定しない。

## Buyer Problems

強い問題は「大量にまとめて処分できると思ったが、定格枚数・運転時間・冷却待ち・処理できる紙が目的と合わない」こと。

|購入ミス|一次資料で説明できた点|
|---|---|
|最大枚数を毎回の推奨枚数と読む|16MCは最大7/定格5枚、SZK01は最大6/定格5枚を分離|
|長い連続時間だけで選ぶ|SZK01の警告時約60分冷却、16MCの空冷・約15分後再使用と再開の注意|
|auto200枚を同時細断と読む|PSD-M200ATは自動給紙セット200と手差8を分離。運転時間も別|
|対応金具/紙を一般化する|SZK01は10/11号、16MCは10号、PSD-M200ATは金具除去。auto折り紙も不可|
|静音表示なら夜間可と判断|空転/細断枚数・距離等が揃わずNOT DIRECTLY COMPARABLE|
|紙とCDを同じ入口へ入れる|PSD-M200ATは紙・カードとディスクの投入口が別|
|安全装置があるから子供も安全と判断|安全スイッチと使用禁止・接近注意は別。安全保証に変換しない|
|どの機種にも同じ油を使う|16MCの専用オイル案内を他機種へ流用しない。可燃性スプレー不可の原本注意を保持|

定格時間×枚数や速度から、実処理量・何分で片付くかは算出していない。保証された日常推奨枚数や詰まりにくさを独自に付けない。

## SERP / Competition

以下は関連結果から選んで読んだページ。順位付きの市場シェア測定ではない。未読・partialを「書かれていない」の証拠にしない。

|Type / Page|確認できた価値|gap解釈|
|---|---|---|
|メーカー直販：[サンワ 選び方](https://direct.sanwa.co.jp/contents/sp/howtouse/how-shredder.html)|検索excerptは定格・連続時間・箱・mediaを扱う|直接取得はメンテナンスへ遷移。partialとして保持|
|大手EC：[ビックカメラ](https://www.biccamera.com/bc/i/topics/osusume_shredder/index.jsp)|方式・連続時間・静音・対応物と多数モデル|一般選び方の追加価値は小さい|
|比較media：[マイベスト](https://my-best.com/35492)|比較・検証・購入候補整理|原本整理だけで実機テストを上回るとしない|
|実測media：[360LiFE](https://360life.shinyusha.co.jp/articles/-/37812)|最大/定格を分離、定格枚数細断中の騒音計測、使い勝手|測定の強みがある。読んだ表では冷却/復帰/auto紙条件の横断整理は限定的|
|office専門店：[OA機器お助けマガジン](https://oa.ohw.jp/shredder/shredder_guide/home-shredder-matome.html)|設置・習慣・用途別候補・media・手入れ|条件の一般論だけでは重なる|
|ブログ/guide：[PROBEFOLIO](https://probefolio.com/articles/shredder-crosscut-microcut-difference/)|2機種の連続/休止・対応紙・設置・購入停止条件を原本付きで扱う|強い直接競合。休止時間を足すだけではgapにならない|
|UGC記事：[ころもち note](https://note.com/koromochi_tips/n/ne1fc92d72943)|紙条件、最大/定格、3機種の冷却、製品ページと説明書差、対応media|原本を読むこと自体も既存。単純な空白市場ではない|
|比較guide：[Room & Tech](https://room-and-tech.com/electric-shredder-crosscut-compare/)|公式/seller根拠を分離、型番差、定格/最大、空転音|休止時間は未確認、autoは対象外。今回scopeとの差はここに限定|
|価格比較：[価格.com](https://kakaku.com/kaden/shredder/)|モデル・仕様・価格、レビューへの入口|掲載数/人気を売上や需要量としない|
|guide：[シゴト環境ラボ](https://work.kurashitomanabi.com/home-shredder-quiet-choice/)|検索excerptに機種別の連続/休止表|直接open timeout。全文不在の主張は禁止|

Q&A/UGCは上記noteや価格比較のレビュー結果も確認したが、使用感・故障率の一般化や商品仕様の証拠には使わない。技術claimは次節のメーカー原本へ戻す。

## Market Gap

**B=3/5**。一般の細断方式・枚数・静音・容量は充実し、定格と休止まで扱う小規模ページも存在する。読んだ範囲で残る仮説は、**メーカー横断で手差し/autoの紙条件、金具、定格枚数、運転/休止、停止後の復帰を一緒に照合すること**。自己復帰する機種とOFF・抜栓して冷却する機種を同じ「休めば使える」にしない。

この統合情報が既存ページより有用かは未検証。「他にない」「SEOで勝てる」とは言わず、原本比較のprototype検証に進める仮説に留める。

## Product Dataset

既存Go Rakuten client、API `20260701`。公式[API仕様](https://webservice.rakuten.co.jp/documentation/ichiba-item-search)のhits/page/sort/availabilityを確認した。反復アクセス制限・429に従いserial interval1.1秒、timeout20秒、最大追加retry2。公式ページから保証された固定quotaを読み取れたとは扱わない。affiliateIDを渡さず正式affiliate URLを生成しない。

主sampleは**2026-10-16 22:23:06～22:23:10 JST**。logical3request成功、最終取得失敗0。transport attempt/retry回数は既存clientから独立集計できずunknown。sampleはpublic Git/既存DBへ追加しない。

|指標|主sample（標準60＋レビュー30）|
|---|---:|
|返却行 / unique source+source_id / 重複|90 / **90** / 0|
|shop数|**55**|
|API price min / median / max|450 / **13,300** / 999,999円|
|review min / p25 / median / p75 / p90 / max|0 / 0 / 0 / 233 / 394 / 4,984|
|review0|60出品|
|API availability1 / 0|65 / 25|
|送料flag0 / 1|44 / 46|
|min1≠max1の価格幅出品|11（同値でもvariant不存在とは言えない）|
|affiliateRate snapshot|90件とも2%|
|baseline family proxy / 最大family / HHI|71 / 7 of90 / 0.02148148|

価格範囲は部品・別カテゴリ・中古・大型を含む。最高999,999円は中古表記かつavailability0の出品で、相場の代表値に使わない。priceはitemPriceであり、商品名のクーポン価格/ポイント後価格ではない。送料flag0=込、1=別として保存するが、地域・選択SKU・注文条件は別確認。

### Cleanup and sampling limits

Go title-screen最終集計：本体候補78、専用オイル1、ごみ袋/MSパック2、別カテゴリ9。替刃0、手動0、展示品0、明示セット0は**今回title規則で該当なし**という意味で、カテゴリ全体で存在しないことではない。重複するflagは中古5、大型office proxy8、訳あり1、メンテナンスシート特典1。中古・大型等のflagは主分類と加算しない。

初回分類で、レビュー特典のメンテナンスシートを本体から誤除外した1件を発見。body/title条件で修正しfixture追加。初回結果を残し、`*-final.json`へ保存した。これは意味の誤分類修正であり、残す商品を価格・scoreに合わせた閾値調整ではない。

中古/展示/セット/大型title proxyを除く本体候補は66出品・42shop、API価格3,880～148,600円、中央値13,935.5円、availability1が50件。ただし大型・用途・新品状態の全件原本確認はしておらず、**66の家庭用適格モデルや新品価格帯とは言わない**。中古表記なしは新品実物の保証でない。選択variant IDはunknown。

標準60は本体候補48、オイル1、袋2、他カテゴリ9。レビュー30は本体候補30。標準sampleの混入とレビュー蓄積方向のbiasを記録し、需要・genre全体の質へ一般化しない。price-stratifiedの再取得や既存Jev/Claude再評価は不要と判断し実施0件。

### Family count interpretation

71は既存resolverの**proxy**であり実モデル数ではない。別カテゴリ、singleton、型番分割/併合、複数型番SKUを含む。全sampleの真のfamily数はunknown。公式identityの実例は次節の**4メーカー4基本model**。SZK01/SZKA01はメーカーが同一仕様と明記するが、P5GCX/P5GCX2、16MC/16MC-Rを推測で同一化しない。v2を採用・変更しない。

## Representative Models

メーカー原本を取得できることを優先し、料率・レビュー点数で選んでいない。ランキングではない。4機種を23項目で確認したprivate `evidence-grid.json`にunknownを残す。

|メーカー / 型番|公式identity / 定格・最大|連続・冷却|楽天出品例 / snapshot|
|---|---|---|---|
|アイリスオーヤマ P5GCX2|[公式型番別仕様](https://www.irisohyama.co.jp/products/electrical-appliances/office-appliances/shredder/personal-shredder/shredder-p5%C3%AF%C2%BC%E2%80%B0) 定格5枚（64g/m²）、最大unknown|定格2分。X2固有の休止/復帰条件unknown|[工具ワールドARIMAS](https://item.rakuten.co.jp/arimas/562091/) 3,999円、availability1、送料flag0|
|アスカ SZK01|[公式製品](https://www.asmix.co.jp/iteminfo/szk01/) 定格5/最大6枚|10分、警告が消えない場合は切で約60分冷却|[リコメン堂](https://item.rakuten.co.jp/rcmdin/4t-1955041/) 15,642円、availability1。title送料無料だがAPI送料flag1で不一致|
|サンワサプライ PSD-M200AT|[公式製品](https://www.sanwa.co.jp/product/syohin?code=PSD-M200AT) auto200/手差8。取説は定格・最大を各明記|auto約60分/手差約20分、約40分休止|[ハイパーファクトリー](https://item.rakuten.co.jp/factory/sanwa_psdm200at/) 52,182円、availability0、送料flag1。購入可能としない|
|フェローズ 16MC / FE-4704101|[公式製品](https://www.fellowes.co.jp/product/shredder/deskside-shredder/FE-4704101.html) 最大7/定格5枚（64g/m²）|30分、過熱時空冷で約15分後に再使用可。停止途中なら再開の場合|[くらしのeショップ](https://item.rakuten.co.jp/e-kurashi/t1o48/) 29,999円、availability1、送料flag0|

16MCは主sampleにないため、22:25:15 JSTにgenre303156＋keyword16MCのstandard1pageを追加した。30出品を得たが、16MC-R・他SKUを含むため90件の統計へ混ぜず、上記FE-4704101の明示例だけ照合した。追加30を30独立familyとしない。4 logical request合計の最終失敗0。

全出品の新品実物・選択variant・checkout価格・地域送料はunknown。SKU同居/セット・別型番を同一新品モデル価格として混ぜない。出品例から注文・在庫保証を導かない。

## Primary Evidence

中核は**4製品ページ＋3モデル取説=7 source**。加えてIris一般FAQ2件（機種固有値の補完には使わない）。現行Sanwaからリンクされた取説はv1.1。検索で先に見つかったv1.0も履歴として残し、根拠をv1.1へ合わせた。current primary9件、superseded別1件。PDFは本文抽出と該当ページ画像を併用し、スキャンのAskaは目視で読んだ。

|ID / Source|確認範囲・取得状態|
|---|---|
|IP P5GCX2公式欄|HTTP200。ページ見出しはP5GCXだが、X2のJAN/仕様欄と旧型欄を分離|
|AP SZK01公式|HTTP200。定格/最大・仕様・SZKA01同仕様の明示|
|AM [SZK01/SZKA01取説](https://www.asmix.co.jp/wp/wp-content/uploads/SZK01_SZKA01_Manual_1911.pdf)|HTTP200、12page、文字抽出不可。紙面4/5/7/8等を画像確認|
|SP PSD-M200AT公式|HTTP200。仕様、現行manual linkを確認|
|SM [PSD-M200AT v1.1取説](https://cdn.sanwa.co.jp/support/setsumeisyo/pdf/P/PSD-M200AT_ver1-1.pdf)|HTTP200、PDF8pageの見開き。紙面2/8～10/14等、表・注記を視認|
|FP 16MC公式|HTTP200。FE-4704101、最大/定格、P-5明示、動作音|
|FM [16MC取説](https://www.fellowes.co.jp/pdf/user_manual_16MC.pdf)|HTTP200、18page。紙面3/10/11/13～17等、特に保護復帰を視認|
|IQ1 [ホッチキスFAQ](https://www.irisohyama.co.jp/support/faq/detail.php?ID=5984)|型番別の10号/11号/不可を列挙。P5GCX2は対象列挙外、転用しない|
|IQ2 [細断できない物FAQ](https://www.irisohyama.co.jp/support/faq/detail.php?ID=237)|一般的な対応物の入口。個別原本より優先しない|

P5GCX2の完全な取説は今回未確定。旧P5GCXや別P5HCSの取説を使わない。ナカバヤシの候補探索もしたが今回原本を確定できず代表4機種へ加算していない。取得不能/partialから仕様の不存在を推論しない。

### Important spec differences

- 細断サイズ：P5GCX2 4×40、SZK01 4×30、PSD-M200AT紙3×9、16MC 2×12mm。P-5は16MCの公式明示だけを保持。方式名・寸法から他機種のDIN levelや完全な安全性を導かない。
- media：SZK01は10/11号針を定格枚数内、カード/CD/DVD等不可。16MCは10号針、カードは指定素材を紙投入口中央へ1枚。PSD-M200ATは金具を外し、カードは手差し1枚、ディスクは専用口1枚。Fellowes Webの画像ALTにCD/clipという語があっても、対応の印と決め付けず未確定を保持。
- auto：PSD-M200ATはA4コピー紙64g/m²、折り紙・綴じ紙不可。auto200はセット数であり200枚同時に刃へ入る意味ではない。手差しと同時使用しない。auto-startとauto-feedを同一視しない。
- noise：Sanwaは空転約65dB/1枚約64.3dB、productには8枚約62.4dBもある。16MCは60dBだが条件未確定、Askaは静音表現でdB未確認。距離・負荷が揃わず**NOT DIRECTLY COMPARABLE**。夜間使用可の保証や静音順位を作らない。
- maintenance：16MCの純正専用オイル案内と可燃性スプレー禁止を保存。Aska/Sanwaの外側・投入口等の清掃へ一律の注油法を流用しない。箱・sensor・安全スイッチも機種の原本へ戻る。
- 寸法/重量/電力/容量：IP/AP/FPは公式表から取得。SPはW375×D279×H558mm/240W/約28Lだが、重量はHTML約17.2kgに対しリンク取説v1.1約15kgで**不一致**。確定重量を決めず、将来の設置比較では要解決。capacityだけから実処理量や水準を算出しない。

## Claim Risks

市場検証メモ12 claimをsourceへjoinし、7 direct verified / 5 qualified editorial rule（うち資料間不一致1）として保存した。公開原稿の全文claim auditではない。unknown-source/orphan/duplicateはGo検証で0。

除外7候補：最大枚数なら詰まらない、60dB以下なら夜間可、細かければ完全安全、auto200同時細断、全号数/位置の針対応、全機種の自動復帰、全機種に同じ油。実機性能、故障率、実推奨枚数、セキュリティ保証を作らない。裁断物・復帰条件・手入れの誤記は機械損傷やけがにつながるためclaim riskを高めに扱う。

## Commercial Viability

55店舗の主sampleと複数価格帯を確認できた。公式identity4例のうちAPI販売可例は3、Sanwa例は不可。型番・必要条件を絞った後に楽天の通常商品ページへ接続する設計は自然だが、今は導線仮説の確認まで。affiliateRateはsnapshot2%で、メーカー/モデル選定を料率に寄せていない。

将来の正式料率・承認・在庫・継続販売・成約・報酬はunknown。検索volume、ranking probability、CTR、CVR、monthly revenue、expected commission、profitabilityを推定しない。追加の有料LLM API request0、tokens0、直接追加API費用$0。既存Jev/Claude評価を再実行していない。

## Market Score

scoreは市場資料からの編集判断であり、検索量や将来収益の実測値ではない。

|Axis|0～5|理由|
|---|---:|---|
|A Search Intent Strength|4|選択・仕様照合・購入・troubleshootingの実在。量はunknown|
|B SERP Content Gap|3|定格/休止/音/紙条件も既存にある。auto/手差し・原本復帰条件の統合に限定した余地|
|C Primary Evidence Depth|4|4メーカー公式＋3取説。X2取説不足、Sanwa資料差等で完全比較には未到達|
|D Buyer Decision Complexity|5|紙質・最大/定格・給紙・時間/休止・media/投入口・復帰・手入れが交差|
|E Product Breadth|4|90出品、公式4基本model。実family全数はunknown、型番同居等の制約|
|F Affiliate Fit|4|複数shop/価格帯、販売可3代表例。SKU/送料/不可例の注意が必要|
|G Durability|4|紙/運転/対応物の判断は持続。型番・版・在庫・音条件は更新が必要|
|Total|**28/35**|**GO**。Strong条件B≥4を満たさない|

## Risks

別表示：**Claim Risk4 / Maintenance Burden4 / SERP Dominance Risk4 =12/15**。主scoreから差し引かない。公式、EC、実測媒体、専門店、原本付き小規模記事が競合する。これは定性的riskで、上位獲得確率やSERP占有率の推定ではない。

更新対象は型番・派生model、manual revision、最大/定格の定義、media条件、価格・送料・SKU・availability。方式・運転/休止・紙条件という問いは持続し得るが、商品表を放置できる意味ではない。

## Decision / Next Step

**GO、公開許可ではない。** 原本から説明できるbuyer problemと商品幅があり、seller-onlyを中心にしなくても比較の素材は取れる。一般的な「家庭用おすすめ」や「休止時間も見る」では弱い。

次工程候補scopeは1つ：**在宅のまとめ処理用シュレッダーを、手差し／オートフィード、定格枚数、運転・冷却・復帰、綴じ/折り/素材の条件で比較する。** 静音の順位・セキュリティ保証を主軸にしない。

現時点のblockerは将来prototypeでの差別化未実証、X2固有原本、Sanwa重量不一致、販売SKU・送料の不確定。市場feasibilityのhard failではないが、将来記事の必要claimになれば解決/除外が必要。本工程ではcontent prototype・記事・HTMLを作らず停止する。除湿機・互換フィルター・D+7の既存判定も維持する。

## Reproducibility

branch `electric-shredder-market-validation`、base `integration-v0.1.0`。private ignored `outputs/electric-shredder-market-validation/`にprotocol、queries、SERP notes、sampleと原本/版、source manifest・hash、23項目grid、Rakuten mapping、claim map、score入力とGo検証manifestを保存。原本全文・raw商品データ・local snapshots・secretsはGitへcommitしない。

主sample SHA256 `1f9c5c7c4454ab63c417f5cbff7cd0a3e31ee4123c8860705dac09b5f291e232`。統計は`product-stats-final.json`と`screened-stats-final.json`が最終。初回の分類誤り・旧manualを上書きせず保存し、補完sampleのwindow/分母を分離した。sourceにはURL/checked_at/timezone/取得方法/取得精度/hash（保存可能分）を保持。Git commit SHAはprivate run-provenanceへ保存する。

```text
go run ./cmd/summarize-market-offers -input <private-sample.json> -output <new-private-stats.json>
go run outputs/electric-shredder-market-validation/validate.go
go test ./...
go vet ./...
git diff --check
```

集計utilityはGo、API取得/LLM評価/公開は行わない。中央値は偶数件なら中央2値平均、percentileはnearest-rank。fixtureはdedupe（別shopは別出品）、unknown flag、価格欠損、title分類、特典の誤分類、deterministic aggregationを検査。既存market.Assessの閾値・missing/hard-failテストも維持。新規統計出力は既存ファイル上書きを拒否する。

2026-10-16検証結果：`go test ./...` PASS、`go vet ./...` PASS、`git diff --check` PASS。12市場claimの出典join PASS、orphan/unknown source/duplicate 0。base比の本番`site/`、`content/published/`、Wrangler、workflow、既存resolver/Jev/Gate/rerankerの差分0。schedulerは操作していない。結果とcommit SHAをprivate run-provenanceにも保存する。
