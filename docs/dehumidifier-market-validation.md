# Dehumidifier External Market Validation

## Objective / Decision
**GO — 29/35。STRONG GOではない。** 除湿機には一次資料で扱える購入判断と楽天での複数モデルがある。方式名だけの一般比較は競合が充実しており、次に進めるなら**部屋干しに使う運転モード・室温・定格条件・排水条件**を機種単位で揃える検証へ絞る。今回の記事作成・公開はなし。

調査期間: **2026-10-15夜〜2026-10-16未明 JST**。API snapshotは10/15 23:58:07–23:58:12 JST。Webはこの調査期間の読解で、個別fetchの秒単位記録はなく日範囲精度。検索/価格/販売状態を恒久的事実としない。

## Prior Evidence / Fixed Protocol
Category Discoveryは除湿機Claude mean64.35で1位・EXPLORE。既存price-stratified検証でも65.00で1位、ROBUST。Jev Gateは100%→97.78%。ただしreview/price条件付きsampleの結果で、検索需要や収益性を意味しない。詳細は既存sampling-bias-validation.mdを再利用し、再評価しない。

外部結果を見る前にprivate protocol.txtへ既存Day4 rubric/thresholdを保存した。A–G各0–5、35点。STRONG GO>=27かつA/B/C>=4・F>=3。GO>=23かつA/C/F>=3。HOLD19–22又は重要情報不足。NO-GO<=18又はintent/一次根拠/楽天商品/追加価値のhard fail。risk3項目は別表示。結果後に閾値を変更しない。

今回base=integration-v0.1.0、互換フィルターPR #18はHOLDのまま追加調査せず、D+7 PR #17を変更しない。

## Search Intent
初期8queryを先に固定し、gap確認1queryを追加。Web検索ツールの結果と関連ページを読む方式。地域/個人化を固定したGoogle上位10位snapshotではなく、順位と月間volumeはunknown。

|Query|Intent / 判断問題|
|---|---|
|除湿機 コンプレッサー デシカント 違い|commercial investigation / 季節と方式|
|除湿機 冬 乾かない|troubleshooting / 室温と除湿量|
|衣類乾燥除湿機 選び方 部屋干し|commercial investigation / 衣類量・用途|
|除湿機 電気代 比較|comparison / モードと消費電力|
|除湿機 タンク 連続排水|operating conditions / 水捨てと設置条件|
|除湿機 適用畳数 木造 鉄筋|informational / 建物・周波数別能力|
|除湿機 小型 ペルチェ 違い|commercial investigation / 小空間と部屋干しの目的差|
|除湿機 おすすめ 比較|commercial investigation / transactional|
|除湿機 比較 温度 満水 連続排水|gap確認 / 既存ページの深さ|

比較・購入判断に結び付く結果は確認できた。検索候補や結果件数から需要量を推定しない。季節性の強さ/正確な検索volume/SEO順位獲得可能性はunknown。

## SERP / Competition
現在の関連結果にはメーカー公式ガイド、EC/価格比較、affiliate/media、ブログ、動画/UGCがある。以下は実際に確認した主要ページ。順位を付けた市場占有率の測定ではない。

|Page|強い点|今回の追加価値仮説との関係|
|---|---|---|
|[Panasonic 選び方](https://panasonic.jp/joshitsu/select.html)|方式、面積、サイズ、費用、手入れを説明|一般的な方式説明はすでに充実。他社横断のモード条件表とは別|
|[三菱電機 比較](https://www.mitsubishielectric.co.jp/home/jyoshitsuki/product/comparison/)|メーカー内の機種別仕様・費用|正確な比較の強い基準。ChoiceLenはメーカー横断で条件を揃える必要|
|[どれ買う？](https://dorekau.net/dehumidifiers/)|用途/予算診断、モデル比較、価格と評価|単純な用途別おすすめを追加するだけでは弱い|
|[Liviq Lens](https://liviqlens.com/dehumidifier-capacity-drain-humidistat/)|能力・温湿度・タンク・排水経路も説明|「タンクと排水も確認」は既存価値。機種別の原本とモード制約の照合が必要|
|[電気代比較ページ](https://al-collabo.com/dehumidifier-cospa/)|1L当たり費用の比較を提示|初回取得後の再取得はpartial。全条件を読めたとは扱わず、性能/効率の事実根拠に使わない|
|[価格.com](https://kakaku.com/kaden/dehumidifier/ranking_2165/)|多くのモデルと価格の入口|検索snapshot確認。人気順位を販売数や市場需要とみなさない|

SERP gapは**3/5**。一般選び方、電気代、排水の説明が存在するため4を付けない。関連ページに一切ないと断言せず、読めた範囲の不足と追加案を分ける。

## Buyer Problems / Potential Gap
追加調査に値するのは「最大値が高いか」より、**実際に使う条件で何が動くか**。
- カタログL/日は定格条件の値。冬の低温室内で同じ値を期待する買い間違い。
- 50Hz/60Hz、木造/鉄筋、衣類乾燥/除湿の数字を混ぜる比較。
- 「ナイト」なら静かに除湿できると思う誤解。DH-C7200の原本ではナイトモードは送風で除湿しない。
- 方式の名前を一括りにする誤り。エコ・ハイブリッドと従来ハイブリッドは構成が同一ではない。
- 連続排水・満水停止・自動停止と、ホース/排水先の条件を確認せず購入すること。

作れる独自価値候補: **機種×使用モード×試験温湿度×周波数×排水/停止条件**の比較資料。未確認欄を残し、除湿しないモードを小さいW値だけで省エネ除湿として順位付けしない。これは仕様/説明書比較の仮説で、実測性能優位は未検証。

## Product Dataset
既存Rakuten Go client、genre=204546（既存Genre確認済み）、standard pages1–2と-reviewCount page1、30件/page、availability=all、interval1.1s、timeout20s、最大retry2。3 request成功/失敗0、新規LLM request0。旧sample/DBに保存せずprivate出力へ分離。

|Metric|Snapshot|
|---|---:|
|取得行 / source+source_id unique|90 / **90**|
|重複出品ID行 / 店舗|0 / **69**|
|API価格 min / median / max|1,999 / 16,800 / 131,450円|
|レビュー件数 min / median / max|0 / 0 / 4,078|
|review_count=0|58出品|
|API availability=1 / 0|68 / 22|
|可変価格min1/max1のある出品|15|
|API affiliateRate|90件とも2%（snapshot）|
|baseline family proxy / 最大 / HHI|83 / 4 of90 / 0.013580|

**83 familyを実family数としない。** 型番不明のlisting singleton、色・年式、セット、部品、中古を含み、baseline分割誤りがあり得る。Family Resolver v2はHOLDのまま採用/変更しない。全sampleの実family数はunknown。別途、公式model identityで4メーカー4familyを確認した（下表）。

titleを使った事前方針どおりの粗分類:
- 中古/未開封/展示品:21。
- 部品:1、他製品（衣類乾燥/竹炭/排気ファン等）:3。
- 無電源/再生乾燥剤等:3。
- 残りの部屋用本体候補:62（方式/仕様未検証を含み「62適格機種」ではない）。
- ペルチェtitle:11（上記群と重複し得る）。

standard60件には中古等21件が含まれ、review30件は本体候補28・吸湿剤等2。標準順とレビュー順の構成差は大きい。レビュー閾値で弱いカテゴリと判定せず、既存Sampling Bias検証の結論も変更しない。本調査は市場全体の無作為sampleではない。価格範囲には部品/中古/小型/業務用が混ざり、家庭用衣類乾燥の相場へそのまま転用しない。

### Representative models / family identity
選定は公式資料へjoinできるモデルを優先。ランキング/売れ筋ではない。価格はAPI itemPrice、coupon/points後の実質価格ではない。

|メーカー・model|公式family根拠|楽天出品例 / API価格 / 状態|
|---|---|---|
|シャープ CV-T71|[公式spec](https://jp.sharp/joshitsu/products/cv-t71/spec/)|[XPRICE](https://item.rakuten.co.jp/a-price/2980000208638/) 23,800円、review732、API availability1。可変SKUあり、選択後価格の再確認が必要|
|Panasonic F-YEX90D|[公式spec](https://panasonic.jp/joshitsu/products/F-YEX90D/spec.html)|[天六なごみ](https://item.rakuten.co.jp/nagomi-pharmacy/nag0gsfv4w94/) 80,080円、review0、API availability1|
|cado DH-C7200|[公式product](https://cado.com/pages/root-7200)|[roomy](https://item.rakuten.co.jp/roomy/ecl16jun22e01/) 59,950円、review487、API availability1|
|アイリス IJC-P70|[公式support](https://www.irisohyama.co.jp/products/support/4967576654197)|[ウエノ電器](https://item.rakuten.co.jp/u-denki/569285/) 19,800円、review654、API availability0。当該出品は購入可能とみなさない|

この4familyの価格snapshotは19,800–80,080円だが、販売可フラグの3出品では23,800–80,080円。最安や最終checkout価格を保証しない。IJD-I50等は複数model/物干しセットの同居listingがあり、今回4familyへ無理に加算しない。CD-P63A3等もsuffixと公式modelの同等性を推測しない。

## Primary Evidence
|ID|Source / 取得状態|使用できる根拠|
|---|---|---|
|P01|Panasonic 選び方、HTML読解|方式差と住宅/洗濯用途の選択観点|
|P02|CV-T71公式spec、HTML読解|周波数別能力と定格条件|
|P03|F-YEX90D公式spec、HTML読解|当該modelの方式・仕様|
|P04|IJC-P70公式support、HTML読解|model identity|
|P05|[IJC-P70取扱説明書](https://www.irisohyama.co.jp/products/manual/pdf/103079.pdf)、本文＋該当page視認|低温/低湿度で除湿量が減る。冷房機器ではない|
|P06|DH-C7200公式product、HTML読解|当該modelの方式/仕様/保守|
|P07|[DH-C7200取扱説明書](https://cdn.cado.com/manual/manual_safety_dh-c7200.pdf)、本文＋印刷p9視認|ナイトの機能、温度条件・停止制約|
|P08|三菱電機公式比較、HTML読解|メーカー内の仕様差、比較軸の確認|
|P09|[アイリス排水FAQ](https://www.irisohyama.co.jp/support/faq/detail.php?ID=39)、HTML読解|記載対象機種の条件。IJC-P70へ転用しない|
|P10|[シャープ排水FAQ](https://cs.sharp.co.jp/faq/qa?qid=142268)、HTML読解|対象機種別に点検/排水条件を読む入口。全モデル共通値にしない|

主要一次source10件。IJD-I50の追加検索resultは別途partialとして保存、確定family/claim根拠に加算しない。PDFのページは視認したが、全sourceの仕様を完全抽出した比較表はまだ作っていない。仕様と取扱条件を取り出せることの検証段階。

## Claim / Compatibility Risks
検証メモ6 claimは出典へjoin、うち4 direct/2 editorial rule。記事本文は作成していないため全文claim audit完了とは呼ばない。
- DM01: CV-T71の周波数/定格条件（P02）。
- DM02: IJC-P70の低温/低湿度（P05）。
- DM03: DH-C7200のナイトは除湿しない（P07）。
- DM04: エコ・ハイブリッドを従来方式と同一視しない（P01/P03）。
- DM05: 排水は型番ごとの条件照合（P09/P10からの編集判断）。
- DM06: 定格L/日を実乾燥/満水時刻と同一視しない（P02/P05からの編集判断）。

除外4 claim: 全機種ナイトは送風のみ、方式だけで冬の乾燥時間/最安電気代保証、健康/カビの確実な解決、連続排水なら無人永久運転。
健康・除菌・カビ防止の効果を販売titleから採用しない。タンク容量÷定格除湿能力から実際の満水時刻は出さない。ホース径/勾配/安全距離を他社・旧型へ転用しない。除湿可能畳数と空気清浄/イオン適用畳数を区別する。

## Market Score / Risks
既存Day4 rubricをそのまま使用。scoreは編集判断であり、需要・順位・売上の観測値ではない。合計/判定はGoで計算する。

|Axis|0–5|理由|
|---|---:|---|
|A Search Intent Strength|4|冬の不調・方式・部屋干し・電気代・排水の購入判断がある。volumeはunknown|
|B SERP Content Gap|3|一般的な方式・タンク・費用説明は既存に多い。機種別モード条件の照合に限定した余地|
|C Primary Evidence Depth|5|複数メーカー公式specと本文確認できるmanualがあり、機能制限を直接抽出できた|
|D Buyer Decision Complexity|5|方式、温湿度、周波数、モード、住宅、排水、停止機能が組み合わさる|
|E Product Breadth|4|少なくとも公式identity4family、現在API販売可例3family。全cohortの真のfamily数は不明|
|F Affiliate Fit|4|楽天複数店舗/価格帯に接続可能。高料率優先や正式リンク発行は未実施|
|G Durability|4|季節を跨ぐ判断は持続。年式・方式・価格・仕様の更新が必要|
|Total|**29/35**|GO。Strong条件のB>=4を満たさない|

riskは別表示: **Claim Risk3 / Maintenance Burden4 / SERP Dominance Risk4（合計11/15）**。公式/価格比較/大手EC/媒体が強く、記事更新も必要。主scoreから引いていない。数値の大きさでSEO難易度を定量推定したわけではない。

## Commercial Viability
楽天の商品URL・API availability・price/feeを確認でき、比較条件を読んだ後の商品候補導線は設計可能。API affiliateRate2%は今回のsnapshotに限る。商品別の正式リンク発行/料率承認/成約/報酬は未確認。ChoiceLenの既存登録・affiliate・本番設定は触れない。正式発行時は既存楽天公式手順で対象SKU/広告表示/rel=sponsoredを確認する。新規有料契約、収益予測は行わない。

消耗品repeatを主な根拠とせず、本体購入判断を価値にする。中古・匿名generic・複数SKU商品は価格表を膨らませるために使わない。継続販売/将来在庫の保証はunknown。

## Comparison to Internal Evaluation
内部Claude review sample64.35/price sample65.00は「コンテンツ機会」の評価。今回market29/35は外部資料と市場の編集判断。尺度が異なり、相関/スコア差を計算しない。内部有望候補が一次資料/商品幅につながった1カテゴリの事例であり、他カテゴリ一般化やengine精度を証明しない。

## Next Step / Stop
次に進める価値はある。候補scopeは**冬も部屋干しに使う除湿機を、室温・運転モード・排水条件で比較する**こと。メーカー名/方式だけのおすすめ一覧は避ける。

次の独立工程でのみ、代表SKUとmodelを固定し、機種別原本から条件の整合を確認したevidence matrixと購入判断表を作り、content prototype・競合比較・claim auditへ進める。**今回は記事を1行も公開用に書かず、構成prototypeやHTMLも作らない。** GOは公開許可/既存SERPより優れる実証ではない。互換フィルターは保留、他カテゴリ調査はしない。

## Reproducibility / Git
private ignored outputs/dehumidifier-market-validation/ にprotocol、Go acquisition/aggregation、sample/CSV、final stats、source manifest、SERP notes、evidence map、score計算を保存。Web全文・credential・DBはpublic Gitへ入れない。
API source SHA256=b6b1fc787f6feb15ba0bff3157c3d968d1d0ab52cfcb0d4493f13e91e27438cd。product-stats-final.jsonが最終統計。初回の再利用統計にはfilter用labelが残ったため、familyの解釈と今回の分類を明示したfinalを別保存し、古い出力は上書きしない。
Jev/Claude requests0、input/output tokens0、追加LLM費用$0。既存API利用以外の契約なし。
branch=dehumidifier-market-validation、base=integration-v0.1.0。本番site/content、D+7、PR #18、protocol、Workers/workflow、scheduler、resolver/評価設定は変更しない。必須checksはgo test ./...、go vet ./...、git diff --check。


検証結果: Goによるscore29/risk11/GO判定と6件の根拠join PASS（orphan0、unknown source0）。go test ./... PASS、go vet ./... PASS、git diff --check PASS。本番/D+7/評価・resolver設定/workflowの差分0。原本を越える性能/健康/安全claimは除外し、記事化は未実施。
