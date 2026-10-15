# Compatible Filter Validation

## Objective
互換フィルターをChoiceLenの2本目候補として進める価値を追加検証する。**今回の判定はHOLD**。狭い適合確認ガイドの可能性は残るが、汎用的な純正/互換比較は既存競合との差が小さく、互換品の独立した製品identityと新構成の競争力が未確定。Day4 GO、Day5 GO、D+7 CONTINUEは書き換えない。公開・本番変更はしない。

調査日: 2026-10-15 Asia/Tokyo。Web資料は日単位のchecked_at、APIは2026-10-15T14:44:14–14:44:16Z（23:44 JST）のsnapshot。価格・販売状態・検索結果はこの時点の情報。検索volume、順位獲得確率、CTR/CVR、収益予測はunknown。

## Prior Evidence
- Day4: filter-compatibility 28/35、GO。直接候補は純正加湿単体と互換セット。ELPA集じん用は別機能の隣接候補だった。
- Day5: prototype overall **4.000**、best competitor **4.000**、差0、judge別差[0,0,0]、GO。decision usefulness5/evidence4/comparison3/actionability5/clarity4、重大claim0。
- この評価は匿名化した構造化要約比較。記事全文が現在のSERPより優れる証明ではない。
- 読み直した入力: 既存Day4 query/source/product/gap、Day5 prototype/claims/products/results/protocol/audit。旧取得結果は上書きせず再利用。Jev・Claude・diversificationの再評価/設定変更なし。
- 今回のGit baseはintegration-v0.1.0（0dd05e4）。D+7は別branch/PR #17の確定履歴。本branchへ積み込んでいない。

## Search Intent
広い語から型番に絞る10 queryを確認した。取得はWeb検索ツールと選択ページ読解であり、地域を固定したGoogle上位10件の順位監査ではない。順位と地域はunknown。SERP存在を検索需要の量と同一視しない。

|Query群|Intent|確認した判断|
|---|---|---|
|空気清浄機 互換フィルター 純正 違い FZ Y80MF 注意 対応機種|commercial investigation|純正/互換の選択|
|FZ-Y80MF 互換 フィルター デメリット / 純正 互換 違い 枠|commercial investigation|枠、性能根拠、購入条件|
|FZ-Y80MF 対応機種 KC-G50 / 寸法 190 互換 楽天|compatibility check|本体・部品・サイズの照合。190は探索queryの一部で、確認済み寸法ではない|
|FZ-Y80MF フィルター 交換時期 臭い / 互換 入らない|troubleshooting|交換と手入れ、取付けの失敗回避|
|FZ-Y80MF 純正 購入 楽天|transactional|純正交換部品の購入|
|site.jp.sharp KC-G50 加湿フィルター 手入れ FZ-Y80MF|informational|対象機種の公式手入れ|
|FZ-Y80MF 専用 枠 純正 比較|compatibility check|互換枠と純正枠の違い|

検索結果には公式部品表、EC、affiliate/media、個人ブログ、レビュー/UGCが混在。Q&A/forumの独立した十分な読解sampleは得ておらず、問題発生率の根拠にはしない。複数の購入判断intentは確認できるが、月間検索数はunknown。

## SERP / Competition
2026-10-15確認。以下は見つかった関連ページであり固定Google順位ではない。外部本文全文を保存せず要約を記録。

|Page|既存の強み|今回読めた範囲での不足/限界|
|---|---|---|
|[くらしチョイス](https://kurashi-choice.net/articles/sharp-air-purifier-filter-genuine-compatible.html)|本体/部品/枠、公式出典、数量・送料・交換条件を説明|商品ごとの専用枠制約を判断表へ落とす余地。型番や枠の一般説明はすでにある|
|[買い替えlabo](https://kaikaelabo.com/reviews/sharp-air-purifier-filter-compare)|部品種別・単価・適合、未検証性能を区別|FZ-Y80MFの枠構成と選択SKUに限定した停止条件までは明確でない|
|[カネゴラボ](https://kanegolabo.com/gokan-filter)|KC-G50の具体的な交換体験とセット価格|実体験はChoiceLenが代替できない。第三者の体験を性能証明として使わない|
|[シャープ公式部品](https://cocorostore.jp.sharp/fz-y80mf.html) / [サポート](https://jp.sharp/support/air_purifier/doc/filter_top.html)|適合一覧、純正の仕様・条件・保証案内|他社の販売SKU・枠制約を横断して選ぶ表ではない|

全ページに記述がないと断言せず、取得・読解範囲の観察とする。競合も公式資料を参照しており、「一次資料を読んだ」だけの差別化は弱い。

## Market Gap
有力な狭い仮説は、**同じFZ-Y80MF表示でも、手元の枠へ付くろ材とは限らないことを販売SKU単位で示す**こと。
- ドーナムの枠なし出品は、専用枠のみ・純正枠では不可と販売者が明記。
- 枠付き出品には複数の枠種類が記載される。買う前の分岐が必要。
- 17出品はAPIのmin1/max1が異なる。表示価格と選択したセット内容を対応させないと比較を誤る。
- 既存競合が扱う「本体型番→枠→入数」の一般論を繰り返すだけなら追加価値は小さい。

価格だけのおすすめ一覧や性能ランキングにはしない。枠対応を独自に保証できず、現時点はgap仮説の確認まで。

## Product Dataset
既存Rakuten Go clientをそのまま使用。事前にprivate acquisition-plan.txtを固定し、keyword=FZ-Y80MF、standard pages1–2、-reviewCount page1、各30件、availability=all、minimum interval1.1s、timeout20s、最大retry2。取得は3 request成功、失敗0。全体704ヒットというAPI値はlisting数であり、family数/需要/全件検証数ではない。

|項目|結果|
|---|---:|
|取得行 / source+source_id重複除去後|90 / **65出品**|
|重複行 / 店舗数|25 / **39店舗**|
|API itemPriceの最小 / 中央 / 最大|780 / 2,480 / 7,480円|
|レビュー件数 min / median / max|0 / 43 / 758|
|titleに互換/非純正 / 純正 / 分類不能|49 / 6 / 10|
|API availability=1 / 0|64 / 1|
|postageFlag output=0(送料込) / 1(別)|47 / 18|
|API affiliateRate snapshot|65件とも2%|
|min1/max1が異なる出品|17|
|既存baseline family / largest / HHI|9 / 52 of 65 / 0.648284|

title分類は決定的な文字一致のみ。純正表示6件を正規品認証したわけではなく、分類不能10件を互換/純正へ推測配分しない。API availabilityは公式fieldのsnapshotで、購入確約ではない。送料の地域例外・注文条件は別途確認が必要。最安780円が説明文のセット一式の価格とは確認できないため、単体比較へ採用しない。

**family数の重要な制約:** baselineは純正と互換を同じfamilyへmergeしており、9を実メーカー/独立製品family数として使わない。一次資料で基準にできる純正FZ-Y80MFは1モデルfamily。互換のOEM/製造元identityを確定できず、実family総数は**unknown**。売り手39≠メーカー39。v2はHOLDのまま、resolverを変更・採用していない。

商品観測と集計はGoで実施し、計算式/条件をprivate sample.go、analyze.goに保存。実験データ・SQLite・secretはpublic Gitへ追加しない。

### Representative offers（推薦順位ではない）
|出品|API価格|扱い|
|---|---:|---|
|[THINK RICH STORE 純正表示](https://item.rakuten.co.jp/thinkrich/e00022/)|2,675円|純正の基準比較候補。sellerの古い交換時間表記より公式の現行条件を優先|
|[Attrape 枠付き1個表示](https://item.rakuten.co.jp/attrape/fz-y80mf-waku/)|2,000円|互換。枠付きの表示、実適合/同等性能は未検証|
|[ドーナム 枠付き](https://item.rakuten.co.jp/donumshop/fz-y80mfa/)|2,270円|互換。枠種類と選択SKUを確認する候補|
|[ドーナム 枠なし専用ろ材](https://item.rakuten.co.jp/donumshop/fz-y80mfgs1n/)|1,350円|純正枠不可というseller制限を示す境界例。純正枠ユーザーへの推薦対象外|

商品名・couponを確定価格へ転用せずAPI itemPriceのみ。比較対象の数量・選択SKU・送料が揃わないため、割引率・年額費用・最安順位は算出しない。複数の販売構成はあるが、4 independent familiesの意味ではない。継続販売はDay4/5からの複数時点の一部観測に限られ、将来供給保証ではない。

## Scope
推奨scope候補: **「FZ-Y80MFの純正・互換フィルターを買う前に、型番と枠を確かめる」**。
- シャープの当該加湿フィルターに限定し、本体例は公式一覧で確認できるKC-G50-W。
- 本体型番と部品型番、純正枠/専用互換枠、入数・選択SKUを照合する。
- HEPA/集じん用、脱臭用、Ag+カートリッジ、他メーカー、海外互換表の横断比較は対象外。
- 別機能セットは同価格比較から外し、混同の境界例に留める。
- 1本目は空気清浄機本体の配置、今回は既存所有機の交換部品購入。メーカー資料を読む方法は共通だが、読者判断・主queryは異なり、強いテーマ重複はない。

## Compatibility Risks
1. 本体型番と部品型番の取り違え。公式一覧は純正部品のものとして扱い、互換品へ転用しない。
2. 純正枠と販売者専用枠の取り違え。「FZ-Y80MF対応」だけでろ材単体が入ると判断しない。
3. 枠品番・世代・マグネット等の構成差。seller記載を公式検証済みの対応表へ昇格しない。
4. 同名・国内海外型番・色記号・販売店モデル差。曖昧な一致は購入を止めて確認する。
5. 寸法・材質・取付方向。互換寸法はseller-only、純正との公差比較/取付方向の原本は今回不十分。図解施工手順は作らない。
6. bundle・枚数・可変SKU。最低表示価格とセット説明を結び付けない。
7. 返品不可/保証主体とメーカー保証の混同。店舗保証はメーカーの性能保証と別。

## Pure vs Compatible
|比較軸|純正の基準情報|互換側で必要な確認|
|---|---|---|
|fit certainty|公式の本体対応一覧|販売者・選択枠・本体の対応。実装確認は別|
|published specs|公式品番・枠なし・条件付き交換目安|選択SKU、枠、入数、seller仕様|
|manufacturer support|シャープの純正品質確認と非純正起因不具合の条件|販売者の返品/保証とメーカー条件を分ける|
|material disclosure|今回の公式商品頁から材質同一性は不明|材質の記載があっても同等性能を証明しない|
|replacement interval|所定条件で約10年、手入れ必要|純正周期を適用しない、長期検証はunknown|
|seller reliability|純正表示だけで真贋確認とはしない|事業者・返品条件・問合せ窓口。レビューを品質証明にしない|

「互換は劣る」「純正なら必ず優れる」「すべて保証無効」の一般化は禁止。メーカーが説明する条件と、未確認の性能を分ける。

## Evidence Audit
checked_at=2026-10-15、timezone=Asia/Tokyo、Web取得は日精度。原本/公式で直接確認できた主要sourceはS01–S04の4件。S05のKC-G50 PDFは検索resultのみ/本文fetch timeoutのためclaim根拠へ使わない。旧Blueair調査・ELPA集じん資料をシャープ加湿互換の性能/保証へ流用しない。

|ID|Source|根拠の範囲|
|---|---|---|
|S01|[シャープ FZ-Y80MF](https://cocorostore.jp.sharp/fz-y80mf.html)|純正適合・枠・交換条件|
|S02|[シャープ別売品サポート](https://jp.sharp/support/air_purifier/doc/filter_top.html)|純正の性能/品質確認、非純正起因不具合の有償条件|
|S03|[形名確認](https://jp.sharp/support/model_name/air_purifier.html)|本体形名確認案内|
|S04|[KC-G50/H50手入れ](https://jp.sharp/support/air_purifier/mt_doc/filter_humi_care_kcg50.html)|対象機種の保守案内|
|S07–S10|上記代表出品とAPI caption|seller記載であることの根拠。性能/fitの独立証明ではない|
|S06|[楽天API公式仕様](https://webservice.rakuten.co.jp/index.php/documentation/ichiba-item-search)|価格/availability/postage/variant field定義|

### Claim inventory
詳細のclaim/source/type/confidence/risk/statusはprivate claim-audit.jsonに保存。

|Claim ID|内容|Source / status / confidence / risk|
|---|---|---|
|CF01|純正のKC-G50-W適合|S01 / verified / high / compatibility|
|CF02|純正には枠なし|S01 / verified / high / purchase error|
|CF03|純正交換目安の条件|S01 / verified / high / lifetime|
|CF04|シャープの非純正起因不具合条件|S02 / verified / high / warranty|
|CF05|本体形名確認|S03 / verified / high / model|
|CF06|KC-G50/H50保守|S04 / verified / high / maintenance|
|CF07|専用枠/純正枠不可というseller表示|S07 / qualified / seller statementのみhigh / compatibility|
|CF08|複数枠種類のseller表示|S08 / qualified / seller statementのみhigh / variant|
|CF09|互換寸法|S09 / qualified / seller-only / dimensions。本文比較値に使わない|
|CF10–11|枠確認の判断手順|S01/S07/S08 / qualified / editorial inference / compatibility|
|CF12|純正同等HEPA等の性能|根拠なし / remove / unsupported / performance|
|CF13|原因を問わず全保証無効|S02の限定条件に反する / remove / unsupported / warranty|
|CF14|互換にも純正10年|根拠なし / remove / unsupported / lifetime|
|CF15|全互換が純正枠適合|S07に反する / remove / unsupported / compatibility|
|CF16|family一致なら同一製造元|根拠なし / remove / unsupported / identity|
|CF17|780円で説明文の全セット|選択SKU未確認 / remove / unsupported / price|
|CF18|取付方向/全枠世代の関係|S05原本未取得 / unsupported / unknown / safety。本文不使用|

source照合の自己audit: 使用claim10件、critical0 / major0 / minor0。未検証8件は本文から除外（qualified寸法1、remove6、unsupported1）。seller文言はその主体を明記。これは独立reviewではなく、公開許可/製品適合証明ではない。

## Prototype Structure
private prototype-structure.mdは構成検証用Markdownで、production publish path外。
1. 本体型番と手元の枠を控える（短い導入）
2. 「枠なし」を買う前に確認すること（専用枠の実在する制限表示）
3. 本体→純正部品→枠→選択SKUの判断表（不明なら購入保留）
4. 純正/互換で根拠のある欄だけ比べる（上記4出品、順位なし）
5. 交換時期と保守は対象機種の説明へ戻る
6. 購入前メモ、調べ方/不明点

図版・hero・FAQ量産・本番HTMLは作らない。商品導線は適合確認後のみ。実機体験を装う文章はない。

### Prototype evaluation
既存day5-content-v1の1–5軸を使った**編集者自己評価**。独立blind評価として扱わず、旧3 judgeの結果を更新しない。overallはDay5同様独立採点軸で、5軸の算術平均とはしない。

|Candidate|decision usefulness|evidence quality|comparison depth|actionability|clarity|overall|
|---|---:|---:|---:|---:|---:|---:|
|今回の狭い構成案|5|4|3|5|4|4|
|現在のC01|4|4|4|4|4|4|
|現在のC02|4|3|4|4|4|4|
|現在のC03|4|2|3|4|4|3|

自己評価差0は独立judge差0ではない。新しいindependent overall / judge別差はunknown。既存Day5の4.000対4.000は維持するが、新構成への再利用採点ではない。今回Strong GO/GOを新規blind criteriaで通過したと主張しない。比較深度と性能evidenceが残るためHOLD。

## Monetization Feasibility
楽天には複数店舗・販売構成があり、型番/枠の照合後に商品候補へ進む導線は自然。ただし現時点では性能同等を訴求せず、純正枠に使えない出品を安いから推す導線にしない。

API affiliateRate=2%は今回65出品のsnapshot。注文・報酬の保証、将来料率、収益予測ではない。[楽天公式リンク案内](https://affiliate.rakuten.co.jp/guides/link/)と[規約入口](https://affiliate.rakuten.co.jp/guideline/)を確認。既存のChoiceLen利用サイト登録/1本目リンクは変更しない。今回このテーマの正式個別affiliate URL発行は未実施で、生成可否のend-to-end検証はunknown。将来公開前に選択SKU/料率/広告表示/rel=sponsoredを確認する。新規リンクは本番へ追加していない。

## Risks
- 意味ある適合ミス回避は示せるが、枠ごとの実適合はseller依存。原本/実装を確認できない場合は購入停止条件にする。
- 現在競合が型番・枠・単価・公式根拠を説明しており、一般論のgapは小さい。
- family9は純正/互換を混ぜる。listing breadthを独立family breadthへ読み替えられない。
- 可変SKU17、純正表示品の一部availability0、送料差/返品条件で比較更新が必要。
- sellerの性能/抗菌/寿命表現、旧ショップの1日12時間表現を現行純正条件へ混ぜない。
- exact search demand/SEO可能性/継続供給/収益はunknown。

## Decision
**HOLD — 追加検証する価値のある狭いscopeはあるが、2本目の公開候補確定には不足。**

|観点|判断|
|---|---|
|intent / buyer problem|買い間違い回避・純正互換選択は明確|
|market / competition|既存情報が強く、一般比較との差が小さい|
|product data|65 listings/39 shops。ただし真のfamily breadth不明、同内容の価格比較は限定的|
|evidence quality|純正4一次資料は強い。互換の性能・fit・寿命はseller-only|
|compatibility risk|専用枠の具体例で説明可能。正しい全対応表を自作するには不足|
|content differentiation|専用枠/SKU別の停止条件に絞る仮説は有用。優位性は未実証|
|monetization feasibility|自然な導線は設計可能。個別リンク発行/収益性は未検証|

旧GOを否定・改定する判定ではない。新たな候補構成と現在の競合に対する追加validationのHOLD。STRONG GOを出すだけのfamily identity、適合根拠、独立競合比較はない。

## Next Step
HOLD解除を検討する場合だけ、対象を絞り、(1)選択SKUと枠品番/内容物/返品条件の証拠を固定、(2)純正/専用互換枠の購入停止条件を保った全文案、(3)現在競合と同形式の新しい独立blind比較・別claim auditを実施する。体験/同等性能を補わないと差別化できない場合はNO-GOも許容。今回これ以上の工程、記事公開、他テーマ市場検証は開始しない。

## Reproducibility / Git / Validation
private artifacts: outputs/compatible-filter-validation/ にacquisition plan、Go sampling/aggregation、65件sample/CSV、product-stats、source manifest、SERP notes、claim audit、prototype構成/scoreを保存。API sample SHA256=a4911b29a77f875a7b86b21ded2725e7dcba772df183df829c1076b1fd662e3d。旧Day4/5入力は元のprivate outputsで保持し、このpublic docへ全文を転載しない。
Jev/Claude新規request=0、input/output tokens=0、追加LLM費用=$0。楽天API3成功/0失敗、有料サービス契約なし。
branch=compatible-filter-validation、base=integration-v0.1.0。public変更は本docのみ。site/、content/published/、D+7 protocol/results/log、wrangler/workflow/schedulerの差分0を確認する。
必須検証: go test ./...、go vet ./...、git diff --check。実行結果はPRへ記録。既存PRのmerge/close、main merge、deployは行わない。


実行結果: go test ./... PASS、go vet ./... PASS、git diff --check PASS。private Go validatorは18 claim/使用10 claim/15 sourceのjoinを確認し、orphan0/unknown source0/除外claim参照0。最初はprivate調査スクリプトの複数mainが全体test対象となったため、Go build ignoreを追加して分離し、全checksを再実行して成功。validation-manifest-final.jsonに最終artifact hashを保存。本番・D+7・workflowの差分は0。
