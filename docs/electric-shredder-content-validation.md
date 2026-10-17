# Electric Shredder Content Prototype Validation

## Objective

在宅でためた書類の処分について、最大枚数だけでは見えない購入条件を原本から比較できるかを検証する。非公開Markdownの検証であり、公開記事・使用試験・処理速度ランキングではない。実施日2026-10-17 JST。

基点は `electric-shredder-market-validation`、PR #21、commit `d57fdec2dcf25bf10b2cd27f2d0d08d3062a7d50`。既存のD+7 CONTINUE、互換フィルターHOLD、除湿機market/content GO、シュレッダーmarket GOは変更しない。

## Frozen Scope

「在宅のまとめ処理用シュレッダーを、手差し／オートフィード、定格枚数、運転・冷却・復帰、綴じ／折り／素材条件で比較する」。最大枚数を日常の処理枚数へ読み替えず、止まった後の動作と投入前の準備までを購入条件にする。

`outputs/electric-shredder-content-validation/protocol.md` を新規prototype・採点前に固定。36項目、4機種、競合3ページ、Day5尺度、判定条件、seed `20261017` を事前指定した。採点後の閾値変更・競合交換・原稿再採点はしない。

## Fixed Models

|メーカー|固定型番|原本の範囲|扱い|
|---|---|---|---|
|アイリスオーヤマ|P5GCX2|メーカー型番別仕様、公式通販|残す。X2固有取説は未確認。旧P5GCXの取説で補完しない|
|アスカ|SZK01|メーカー仕様、SZK01/SZKA01取説|残す。同仕様の明示はメーカー記載に限定|
|サンワサプライ|PSD-M200AT|メーカー仕様、現行リンク取説v1.1|残す。重量不一致は解消せず別記|
|フェローズ|16MC / FE-4704101|メーカー仕様、16MC取説|残す。16MC-Rとの同一性を仮定しない|

除外・差し替え0。資料不足を隠すための入れ替えは行わず、復帰や素材の未確認が購入条件に響く場合は読者の候補を保留する。

## Buyer Problem

最も強い問題は、**ためた紙の状態と、待てる冷却・停止後の操作が、購入候補の運用条件に合うか**である。auto200枚でも折り紙・綴じ紙・金具をそのまま任せられるとは限らず、長時間運転の表示だけでは停止後の復帰を判断できない。

## Evidence Matrix

4機種×36項目＝144セルをprivateの `evidence-matrix-v2.json` に保存した。verified 89、qualified 23、unknown 32。144セルが存在することと、全条件の根拠が揃ったことは別である。各セルに型番、値、出典ID、紙面番号／HTML型番欄、status、direct/inference、限定条件を付ける。PDF物理頁と紙面頁、2-up印刷の紙面番号を区別する。unknownは非対応・0の意味ではない。

対象項目は、型番、方式、切片、定格／最大、紙条件、投入口、手差し／auto、運転時間、冷却、保護停止、復帰、リセット、針／クリップ／カード／ディスク、専用口、折り紙／綴じ紙、逆転、詰まり対策、箱、満杯停止、安全スイッチ、子供の警告、騒音と条件、電力、外形、重量、手入れ、注油、禁止素材。未確認セルの数は後掲の再現性記録に残す。

主要資料（市場検証の保存資料を再利用し、今回の原本照合と区別）：

- [アイリスオーヤマの型番別製品仕様](https://www.irisohyama.co.jp/products/electrical-appliances/office-appliances/shredder/personal-shredder/shredder-p5) のP5GCX2欄。旧型欄を混ぜない。
- [アイリスプラザ公式通販](https://www.irisplaza.co.jp/index.php?KB=SHOSAI&SID=H562091)。今回追加確認した最大枚数・針の表示。
- [アスカ製品情報](https://www.asmix.co.jp/iteminfo/szk01/)、[取説](https://www.asmix.co.jp/wp/wp-content/uploads/SZK01_SZKA01_Manual_1911.pdf)。紙面4/5/7/8等を視認。
- [サンワ製品情報](https://www.sanwa.co.jp/product/syohin?code=PSD-M200AT)、[取説v1.1](https://cdn.sanwa.co.jp/support/setsumeisyo/pdf/P/PSD-M200AT_ver1-1.pdf)。旧版v1.0を現行として使わない。
- [フェローズ製品情報](https://www.fellowes.co.jp/product/shredder/deskside-shredder/FE-4704101.html)、[16MC取説](https://www.fellowes.co.jp/pdf/user_manual_16MC.pdf)。紙面3/10–17等。

## Rated vs Maximum

|型番|定格／最大の記載|紙条件・比較上の限定|
|---|---|---|
|P5GCX2|メーカー定格5枚／公式通販最大5枚|A4上質紙64g/m²。通販の最大という語の試験定義は未確認|
|SZK01|定格5/5枚（50/60Hz）、最大6枚|A4コピー紙。坪量は今回unknown|
|PSD-M200AT|手差し定格8／最大8、auto最大200枚セット|A4コピー紙64g/m²。autoセット容量は同時細断枚数ではない|
|16MC|定格5枚／最大7枚|A4コピー紙64g/m²。最大は10回連続、定格は定格時間内の細断率90%以上という原本の定義を維持|

同じ「定格」でも試験定義・紙の坪量・投入方法を保持する。枚数の単純順位、必ず詰まらないという保証は出さない。

## Manual vs Auto Feed

PSD-M200ATのautoは紙を順次送るセット数。手差し8枚と別で、両者の同時使用も不可。autoには折った紙・綴じた紙を入れず、ホッチキスやクリップを外す。手差しの折り紙条件をautoへ移さない。他3機種について確認できないauto欄を0や非対応の証明にしない。

## Run / Cooldown / Recovery

|型番|公表運転時間|冷却・停止後の条件|
|---|---|---|
|P5GCX2|定格2分|X2固有の休止時間と復帰操作はunknown。復帰が必須条件なら確認まで保留|
|SZK01|10分|警告時は原本の逆転・詰まり確認。警告が消えない場合は「切」で約60分冷却|
|PSD-M200AT|auto約60分／手差し約20分|保護停止ではOFF・抜栓して約40分冷却。詰まりは給紙方法別の原本手順へ|
|16MC|定格30分、状況で前後|過熱停止後は空冷。約15分後、温度が下がりランプ消灯で再使用可。途中の細断が再開する場合がある|

16MCの停止物・投入口に触れない／衣類を近づけないという警告を保持する。「全機種は冷やせば自動復帰」とまとめない。過熱と詰まりも別の状態として扱い、逆転反復を万能な除去法にしない。この表は修理・使用手順の代替ではない。

連続時間・冷却時間は公表条件の比較として使えるが、同一負荷の実処理能力ではない。回転数×枚数×時間の理論処理量、1時間あたり枚数、詰まりにくさ、実寿命は推計しない。

## Media Conditions

|型番|針・折り／綴じ紙|カード・ディスク|
|---|---|---|
|P5GCX2|公式通販は10/11号。束・位置などの詳細unknown|X2固有条件unknown|
|SZK01|10/11号を定格内。他の針・クリップは除去。針なし綴じは1枚ずつほぐし枚数確認。折り紙は枚数確認、厚いDM束不可|プラスチック・CD/DVDは原本で禁止|
|PSD-M200AT|金具除去、auto折り／綴じ紙不可。手差し二つ折りは1枚を2枚相当として定格内|指定プラスチックカード厚0.2–0.8mmを手差し1枚。12cm CD/DVDは専用口1枚|
|16MC|10号まで。折り紙は紙厚増で枚数減。クリップ条件unknown|指定カードを紙口中央へ1枚。薄型・ラミネート等は除外。CD/DVDは今回未確定|

WebアイコンのALT文字だけを対応根拠にしない。全ホッチキス、全カード、全メーカー共通の禁止素材表へ一般化しない。切片寸法や公式P-5表記から情報漏えい防止・総合安全順位を作らない。

## Noise Comparison Limits

Sanwaは空転約65dB／1枚約64.3dB（取説）、8枚約62.4dB（Web）を区別。16MCの公称60dBも負荷・距離・環境が一致していない。Askaの静音表現やIrisの未確認値を補わず、**NOT DIRECTLY COMPARABLE**とする。夜向け・静音No.1を出さない。

## Maintenance

16MCの外装清掃（OFF・抜栓、紙面15）と、専用オイルを紙に塗り細断する手入れ（紙面16）は分けた。他社へ油種・注油方法・頻度を転用しない。Askaの外側・投入口清掃、Sanwaの外側・箱清掃はそれぞれの原本へ。全機種の注油間隔は揃わずunknownを残す。可燃性スプレー禁止などの原本警告を保持する。

## Rakuten Mapping

2026-10-16の市場検証API snapshotを再利用。3機種は22:23 JST台、16MC補完は22:25 JST台の別観測で、今回の現価格ではない。追加Rakuten取得0。価格順位・最終支払価格・新品実物状態の保証はしない。

|公式型番|shop / source_id|API itemPrice|availability / postageFlag|限定|
|---|---|---:|---|---|
|P5GCX2|工具ワールドARIMAS / arimas:10910213|3,999円|1 / 0|選択variant unknown|
|SZK01|リコメン堂インテリア館 / rcmdin:11312923|15,642円|1 / 1|送料無料titleと送料別flag不一致|
|PSD-M200AT|ハイパーファクトリー / factory:10067065|52,182円|0 / 1|購入可能候補とはしない|
|16MC|くらしのeショップ / e-kurashi:10051804|29,999円|1 / 0|16MC-Rへ同一化しない|

各snapshotのmin/maxは同値でも選択SKU IDはunknown。中古表示が見当たらないことだけで実物の新品を保証しない。クーポン・ポイント・送料をitemPriceへ合算しない。型番joinはメーカーidentityで行い、HOLD中のFamily Resolver v2を採用しない。affiliateリンク発行0。

## Prototype Structure

private `prototype-v2.md` は短い導入→定格／最大→autoの準備→運転／冷却／復帰→金具・紙・媒体→購入前判断表→騒音・手入れの限界→楽天型番／条件→unknown。27掲載claimをEC01–EC27で追跡する。v1と初回監査を保存し、事実修正後にv2を採点入力として固定した。公開HTMLは作っていない。

## Decision Table

|読者の必要条件|購入前に残す条件|停止・保留条件|
|---|---|---|
|まとめ処理で待ちを減らしたい|紙量・頻度と給紙、定格、運転／冷却の組|復帰unknownが用途に響くならX2を保留。時間処理量は推計しない|
|autoに任せたい|紙を開き金具を外せる|折り／綴じたまま使う目的では候補を残さない|
|ホッチキスを残したい|号数、定格、束・位置の条件|全号数可としない。針なし綴じも別条件|
|カード／ディスクも処分したい|素材、1枚条件、投入口|対応unknownを可にしない|
|停止後に再開してほしくない|原因と停止／再開の原本動作|16MCの途中細断再開を無視しない|
|子供のいる場所に置く|原本の接近・使用警告と設置条件|安全スイッチを子供のそばでも安全の根拠にしない|
|狭い場所へ置く|外形、箱の取り出し、投入口、電源|本表だけで設置可としない。Sanwa重量は未解決|

## Claim Audit

初回は独立reviewerが27掲載claim・144セル・全比較表を保存原本と照合し、critical 0 / major 0 / minor 5。評価score・競合・過去GOを提示していない。初回指摘は削除せず保存した。

修正内容：X2禁止素材のstatusをunknownへ、Aska針なし綴じ条件追加、16MC外装清掃／注油工程分離、セル別出典位置とdirect/inference整理、Sanwa自動再開の編集判断をqualified・Aska専用口をunknownへ。補足でX2手差し容量の公式通販出典を追加し、詰まらない保証の推論を明示的にqualifiedとした。

高リスク6類型（詰まり・耐久、夜間騒音、安全／漏えい、業務用／処理枚数時、全機種自動復帰、全針／auto全素材）はremoveとして非掲載。unknownをseller値や方式名から埋めない。

修正後の独立再確認は **critical 0 / major 0 / minor 0**。初回5件と出典位置・版表示の追加修正履歴を `claim-review-2.json` に保存し、監査対象3ファイルのhashも記録した。保存原本と限定範囲の監査であり、メーカーへの照会や実機での再現試験まで完了した意味ではない。機械的な12出典ID／27掲載claim／144セルのjoin検証もPASS。

## Competitor Comparison

prototype前に固定した3ページを2026-10-17に確認した。特定の回復条件が見当たらないことを全文・サイト全体の欠落と断定しない。

|比較対象|実際の強み|今回scopeとの境界|
|---|---|---|
|[360LiFE実測ガイド](https://360life.shinyusha.co.jp/articles/-/37812)|7機種の実物、細断性能・音・扱いやすさ・機能、負荷を伴う騒音測定|autoの紙条件と原因別復帰の横断照合は確認した表の中心でない|
|[PROBEFOLIO条件比較](https://probefolio.com/articles/shredder-crosscut-microcut-difference/)|原本付き2機種、紙分類、定格／最大と休止、購入停止条件、買わない選択肢|少量手差し中心。auto／原因別復帰の4機種比較ではない|
|[ころもちnote定格・休止ガイド](https://note.com/koromochi_tips/n/ne1fc92d72943)|3機種の原本定義、Web／取説差、冷却・操作、媒体と不向きな用途|autoの束条件・途中再開状態は確認範囲の横断軸でない|

競合にも定格・休止・紙条件の価値が既にある。単に原本を読んだことや注意書きの量を差別化としない。

## Independent Evaluation

3名のfresh isolated judgeには同一prompt `judge-prompt-v1 / day5-content-v1` と4候補の匿名要約だけを提示。自稿ラベル、型番・ブランド名、過去判定、他judgeの回答、役割keyは見せていない。各要約は同じ6区分で649～658文字。原稿／根拠／修正後audit／要約／promptをhash固定後、各judge1回、4候補×6軸＝24整数、計72採点を得た。scoreを見た修正・再採点0。全12行の欠損・重複・範囲・notesを検証してから役割keyを結合した。

尺度は既存Day5の1～5整数。overallは軸の算術平均ではなく独立軸。各数値は3judge平均。

|候補|decision usefulness|evidence quality|comparison depth|actionability|clarity|overall|
|---|---:|---:|---:|---:|---:|---:|
|ChoiceLen prototype|5.00|4.00|5.00|5.00|4.00|5.00|
|360LiFE|4.00|5.00|4.00|4.00|4.00|4.00|
|ころもちnote|4.00|4.00|4.00|4.00|4.00|4.00|
|PROBEFOLIO|3.67|4.00|3.00|4.00|4.67|4.00|

**overallの最強競合は3ページ同率4.00**。実測根拠では360LiFEが5.00でprototypeの4.00を上回る。休止・定格というscopeで最も重なる競合はころもちnoteで、既存の原本比較を過小評価しない。

各judgeのprototype overallは5/5/5、各judge内のbest competitorは4/4/4。差は+1/+1/+1、clear win 3 / tie 0 / loss 0、非劣位3/3。全体平均差+1.00。全軸完全一致ではなくPROBEFOLIOのusefulness・clarityには評価差がある。今回の狭い課題への要約適合という結果であり、同一provider・小標本による一致を独立した人間読者3名の検証や統計的有意差と呼ばない。

事前固定した既存gate：STRONG GOにはoverall≥4.3、競合差≥0.30、非劣位3/3、**evidence≥4.3**、critical/major 0が必要。GOにはoverall≥4、競合差≥0、非劣位2/3以上、critical/major 0が必要。HOLD/NO-GOもprotocolの既存ルールを維持。今回はevidence 4.00がSTRONG GO条件未達で、Go集計のscore gateはGO。独立auditの重大指摘0と残存unknownを別途確認して最終判定へ反映した。

## Risks

- X2固有取説が未確認で、完全な4機種の復帰・素材比較は未成立。
- Sanwa重量はWeb約17.2kg／取説v1.1約15kgの不一致。購入可能在庫も今回のsnapshotでは0。
- Aska紙の坪量、騒音条件、SKU・最終送料などはunknown。
- 実機の速度、詰まり、寿命、安全性、騒音を検証していない。
- 匿名要約はrootが作成。同形式・近い長さでも編集者バイアスがあり、同一providerのjudge間相関もある。原文全文・UI・読者行動・SEO順位・CVRの比較ではない。

## Decision

**GO。STRONG GOではない。公開承認でもない。**

条件の意味が異なる数値をランキングせず、紙の下準備・給紙・定格・待ち時間・復帰状態から候補を保留できる記事価値は成立した。採点は最強競合に非劣位以上、修正後critical/major 0で既存GO条件を満たす。market GO28/35からの自動昇格はしていない。

ただしevidence qualityは4.00でSTRONG GO条件を満たさず、X2固有原本不足で4機種の復帰横断比較は完全でない。実測の強い競合に根拠品質で勝ったとは言わない。安全なunknown／購入保留を維持した限定scopeのGOであり、X2の冷却・素材を断定する原稿へ変える場合は新たな原本確認が必要。現時点の非公開validationを止めるblockerはないが、完全な4機種比較やpublicationにはこれら制約が残る。

## Next Step

この工程で停止する。除湿機content GOと本結果を同一尺度・同じ資料制約で比較し、どちらを2本目のpublication candidateにするかは別工程で判断する。記事公開、affiliate発行、sitemap／GSC／Cloudflare変更には進まない。

## Reproducibility

private ignored folder `outputs/electric-shredder-content-validation/` にprotocol、model-freeze-note、source manifest／PDF／視認画像、144セル、楽天mapping、v1/v2 prototype／claim、初回／修正後review、同形式匿名要約、prompt、hash付きblind manifest、judge JSON／metadata、集計／integrity／validation manifestを保存する。公開GitへPDF、raw API、価格履歴、スクリーンショット、秘密値は追加しない。

再利用した `cmd/compare-content` は除湿機検証commit `dbf95a3` の2ファイルを選択的に取り込み、今回の明示seed指定を追加した。branch全体のmergeや既存判定変更はない。既存 `internal/contentvalidation.Decision`／rubricを変更せず、入力欠損・重複・範囲外・key join・hash変更・既存出力上書きを拒否する。ソースjoin検証と原本の意味の監査は別である。

```text
go run ./cmd/compare-content -mode audit -dir outputs/electric-shredder-content-validation
go run ./cmd/compare-content -mode prepare -seed 20261017 -dir outputs/electric-shredder-content-validation
# 独立judge各1回の出力を得た後
go run ./cmd/compare-content -mode analyze -dir outputs/electric-shredder-content-validation
```

生成済み出力へ再実行すると上書き拒否となる。独立judgeの回答を生成・再現するコマンドではなく、保存入力・回答を検証／集計するもの。score gateの重大指摘0という前提は別途の修正後auditで確認する。

費用：今回の新規Jev／Claude API呼び出し0、追加外部有料LLM API費用$0。独立reviewerとjudgeはOpenAI/Codex。exact model ID、入力／出力tokens、追加Codex課金額は取得不能につきunknown。全LLM費用を$0とは呼ばない。

LLM run数はclaim reviewer初回＋修正再確認の2、blind judge 3、合計5。reviewer入力は各回27掲載claim／144セルと原稿・出典、出力は各回1監査JSON。judge入力は各回4要約、出力は各回4採点行＋metadata。初回レビューと修正再確認は再採点ではない。provider／prompt version／countをprivate metadataとvalidation manifestへ記録する。

検証結果：`go test ./...` PASS、`go vet ./...` PASS、`git diff --check` PASS。追加utilityは明示seed、決定的集計、judge不備、範囲外値、hash変更、出力上書き、出典joinのfixtureを含む。`site/`、`content/published/`、親market validation文書への差分0。既存Jev／Claude／Gate／resolver／Radar設定への差分0。新規commit対象は本reportと `cmd/compare-content` の2 Goファイルのみ。production操作0。
