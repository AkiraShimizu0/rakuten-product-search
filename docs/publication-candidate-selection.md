# ChoiceLen — Second Publication Candidate Selection

## 1. Objective

**SELECT_DEHUMIDIFIER**。次のPublication Readiness / Preparationの優先候補を除湿機とし、シュレッダーは **VALIDATED_RESERVE_CANDIDATE** とする。両候補のContent判定はGOのまま。STRONG GOへの昇格、公開許可、SEO・収益性の予測ではない。

実施日2026-10-20 JST。既存結果を比較する選定工程だけを実施し、市場調査、記事編集、原本の新規取得、商品取得、独立judge再実験、公開準備は行っていない。

## 2. Input Evidence / Provenance

baseは `integration-v0.1.0` / `0dd05e4345f6e06620fae0bfc8fb06a4cc4bc78a`。以下の固定commitの文書を `git show` で参照し、branchのmerge/cherry-pickは0。privateのv2 matrix全76／144セル、prototype全文、claim再監査、score summary、楽天mapping、source manifestも読み直した。

|ID|固定資料|commit / 観測時点|
|---|---|---|
|D19|[PR #19 市場検証](https://github.com/AkiraShimizu0/rakuten-product-search/blob/1d5f4e7c15535185edf0a71c045b1765cab3e0e1/docs/dehumidifier-market-validation.md)|`1d5f4e7c15535185edf0a71c045b1765cab3e0e1`、市場10/15夜～16未明、楽天10/15 23:58 JST|
|D20|[PR #20 content検証](https://github.com/AkiraShimizu0/rakuten-product-search/blob/dbf95a3eeae0dfc0e810d52906fa8d4306bc36c1/docs/dehumidifier-content-validation.md)|`dbf95a3eeae0dfc0e810d52906fa8d4306bc36c1`、原本・prototype確認10/16|
|S21|[PR #21 市場検証](https://github.com/AkiraShimizu0/rakuten-product-search/blob/d57fdec2dcf25bf10b2cd27f2d0d08d3062a7d50/docs/electric-shredder-market-validation.md)|`d57fdec2dcf25bf10b2cd27f2d0d08d3062a7d50`、市場・楽天10/16|
|S22|[PR #22 content検証](https://github.com/AkiraShimizu0/rakuten-product-search/blob/abb4b2ad8cee68efea8899c24c5c268110f8fbaf/docs/electric-shredder-content-validation.md)|`abb4b2ad8cee68efea8899c24c5c268110f8fbaf`、原本・prototype確認10/17|

出典略号は候補内で独立（例：D.SM=Sharp、S.SM=Sanwa）。D20/S22のsource IDs・紙面位置・claim IDsへ追跡する。今回の再読は10/20の現行Web・在庫の再確認ではない。両content実験のblind-manifestに記録された全入力hashを照合し、不一致0。今回の参照コピー／privateファイルhashは `outputs/publication-candidate-selection/input-provenance.json` に保存した。原本確認範囲の限界も引き継ぐ。

|既存評価（変更なし）|除湿機|シュレッダー|
|---|---:|---:|
|Market / SERP gap|29/35 GO / 3/5|28/35 GO / 3/5|
|Content overall / evidence|5.00 / 4.00|5.00 / 4.00|
|Best competitor / 差|4.00 / +1.00|4.00 / +1.00|
|Judge非劣位 / 修正後critical・major・minor|3/3 / 0・0・0|3/3 / 0・0・0|
|Content decision|GO|GO|

別実験の同点は同等の読者価値の証明ではない。セル数・unknown率を品質順位にしない。D+7 CONTINUE、互換フィルターHOLDも維持し、互換フィルターを比較候補へ加えていない。

## 3. Frozen Selection Protocol

`outputs/publication-candidate-selection/protocol.md` を候補資料の今回の再読・採点前に固定した。SHA256 `4daf3d344704ef6995ff873d28ed88fe9432558a569165834947ea1fe3e13878`。既存結果は会話から既知であり、結果を知らないblind preregistrationと称さない。今回の独立したselection基準v1で、既存protocolの改変ではない。

8軸の重みはA有用性15%、B差別化20%、C一次根拠20%、D低リスク15%、E低保守負荷10%、F広告導線10%、G時期適合5%、H追加作業の少なさ5%。各0～5整数：0不成立／重大欠陥、1大きな不足、2不足が目立つ、3制約付きで実用、4十分な根拠、5非常に強く重要制約が少ない。D/E/Hは負担・リスクが少ないほど高得点。

Hard Gateを先に評価。PASS／CONDITIONAL／FAILとし、CONDITIONALは既存のunknown表示・購入保留・比較範囲限定で誤推薦を防げる準備候補に限る。中心判断に必須の根拠がなく安全な限定scopeも成立しなければFAIL、点数で救済しない。両者FAIL又はタイブレークでも決まらなければHOLD_BOTH。

合計100点＝Σ（重み%×得点÷5）。差が **5点以内（5を含む）** は接戦。①中心判断の根拠、②具体的差別化、③公開前blockerの少なさ、④低保守負荷の順に判断する。採点後の重み・閾値変更0。未知の更新頻度・工数・検索量は補完しない。

## 4. Eligibility / Hard Gates

|Gate|除湿機|シュレッダー|
|---|---|---|
|1 中心判断は一次資料で支持|PASS：D20 DD08–19。夜間動作、排水可否・停止をメーカー横断で照合|CONDITIONAL：S22 EC07–19。3原本の給紙・復帰は支持。X2の肯定的な復帰／素材推薦は不可|
|2 修正後critical/major=0|PASS：review2で0/0。最長15hの誤一般化は修正済み|PASS：review2で0/0。清掃／注油や針なし綴じの修正を保持|
|3 誤使用・安全リスクを管理|CONDITIONAL：ホース設置・点検・停止を型番別に保持。無人永久運転を許可しない|CONDITIONAL：過熱と詰まりを分離。16MC途中再開・接近警告、Sanwa OFF／抜栓を保持。除去手順の代用にしない|
|4 必須identity・mode・適合に重大な未解決がない|CONDITIONAL：4型番identityと中心mode／排水は確認済み。Iris低温適合・Panasonic連続上限の肯定判断は保留|CONDITIONAL：4型番identityは確認済み。X2の復帰・素材適合は保留。16MC-Rを混ぜない|
|5 unknownを購入保留へ落とせる|PASS：D20判断表。室温・排水経路・長時間条件が不明なら保留|PASS：S22判断表。紙の準備・冷却・素材条件が不明なら保留|
|6 未検証の保証を必要としない|PASS：冬の乾燥速度・省エネ・騒音順位なし|PASS：処理枚数/時・詰まり・耐久・完全漏えい防止保証なし|

両者とも **限定境界付きで準備候補として適格**。FAILはないが、publication READYとは判定しない。Gate3/4は完成稿で保留境界が失われれば未達になる。肯定的に「使える」と言うための根拠と、根拠不足ゆえ「保留」と言う記事価値を区別した。

## 5. Side-by-Side Comparison

|判断対象|除湿機の既存原本で言えること|シュレッダーの既存原本で言えること|
|---|---|---|
|数値の条件|4機種の27℃60%・50/60Hzを保持。IJC-P70の定格modeはunknown。冬の実性能へ変換しない|最大／定格、A4紙64g/m²等を保持。SZK01坪量はunknown。autoセット数を同時細断枚数にしない|
|用途と実動作|DH-C7200ナイトは送風のみ。F-YEX90D音ひかえめは除湿、低温保護で風量増の場合|PSD-M200ATはauto200セットと手差し定格8を分離。同時使用不可。auto折り／綴じ紙は不可|
|停止・復帰|CV-T71約12h停止は解除可、cado全mode約12h、Iris24h無操作、Panasonic自動停止設定と連続設定を分離|SZK01警告が消えなければ切で約60分、Sanwa OFF・抜栓約40分、16MCは空冷約15分・ランプと途中再開。X2はunknown|
|運用条件で候補を外す|ホース必須ならF-YEX90Dを外す。ホース径・勾配・高低差・タンク条件を3社別に照合|金具を外せないauto用途ならSanwaを残さない。Aska10/11号・針なし綴じ、16MC10号、カード／CD入口を分離|
|公開上の制約|温度範囲が定格維持を意味しない。設置・点検・停止原本の代替ではない|時間から実処理量を算出しない。安全スイッチを子供の接近安全へ変換しない|

D20 matrixの主要根拠：D.SM3/7–13/18、D.PM10/19–23、D.PC当該型番の連続排水欄、D.CM8–11/16–17/25/30–31、D.IM2/5/8–10/15。S22：S.IPのX2欄、S.AM4/5/7、S.SM2/8–14、S.FM3/10–17。低温・運転・冷却の略記だけで使用手順を提示しない。

## 6. Evidence Readiness

分類単位は「その条件で候補を残せるか」という質問群でありセルではない。同じ欠損を複数項目で重複加算しない。

|ID / 未確認事項|種類と現scopeでの扱い|取得実現性／失敗時の限定scope|
|---|---|---|
|D-U1 IJC-P70の全運転温度|Non-blocking。寒い部屋への適合を肯定するならBlocking|保存IM2は効率15～30℃の説明に留まる。追加根拠が取れる保証なし。失敗時は低温適合候補から保留し、mode／排水条件の確認例だけ残す|
|D-U2 F-YEX90D連続設定の停止上限|Non-blocking。長時間連続の適合を肯定するならBlocking|保存PM10/19で別設定は確認、上限未確定。約15hを転用しない。失敗時は連続時間保証を外し、自動停止設定とホース非対応の判断に限定|
|D-U3 IJC定格mode・専用衣類／ナイト、全規格一致|Non-blocking。未確定のmodeを存在／不存在としない|保存原本の既知の弱・強・おまかせ動作を使う。機能名・能力の完全同条件ランキングを行わない|
|D-U4 冬の実乾燥速度、最静音、最省エネ、年間費用|Optional／scope外|仕様だけで解消できない。今回価値に不要で、性能順位・費用保証を追加しない|
|S-U1 P5GCX2固有の冷却・復帰・原因別操作|Non-blocking。まとめ処理適合を肯定するならBlocking|固有取説自体が未発見で取得不確実性が大きい。旧型を流用しない。失敗時はX2を積極比較から保留し、原本のある3機種中心へ限定するかを準備段階で決める|
|S-U2 X2の紙・素材・束条件、安全な投入条件の詳細|Non-blocking。素材適合の肯定推薦にはBlocking|公式通販10/11号だけでは束等の条件不足。失敗時は詳細unknown／購入保留、非確認素材の許可を出さない|
|S-U3 Aska坪量・針位置・冷却後自動再開、16MCクリップ／CD、他機種auto確定|Non-blocking。断定を避け機種固有の確認済み操作だけ使う|保存原本では限界が残る。自動再開する／しないや媒体対応を推測せず、その条件が必須なら保留|
|S-U4 Sanwa重量17.2kg対15kg|Non-blocking。重量条件で設置可能と肯定するならBlocking|Web／取説差が保存済み。軽い値を選ばず重量比較を外す。解消成功は保証しない|
|S-U5 1時間処理量・詰まり・寿命・統一騒音・完全な漏えい防止|Optional／scope外|新しい実測なしでは解消不可。判断表の成立に不要|
|共通C-U1 選択SKU／新品・セット状態・現行販売・送料・正式リンク|Non-blocking（今回の選定）。具体的商品導線を公開する前の必須確認|過去APIから現況は確定不能。確認できない出品は購入可能として掲載せず、affiliateなし／確認できる型番のみに限定。リンクは今回発行しない|

**既存の保留境界を守る記事価値に、無条件の公開停止を要する未解決の中心欠損群は両者0。** ただし「全固定4機種を積極的に選べる」とする強いscopeには、除湿機D-U1/U2の2群、シュレッダーS-U1/U2の2群が不足する。これは同等の深刻さという意味ではない。除湿機は確認済み4原本の中の限定条件、X2は中心の復帰と素材を読む固有原本全体が不足している。S-U4は重量を主軸にしないため必須群に加算しない。商用確認C-U1は別の共通公開前義務で、中心根拠不足と混算しない。

## 7. Competitive Differentiation

**除湿機：比較価値あり。ただし独占的gapではない。** D19/D20では方式・費用・タンク・排水の説明が既にあり、Panasonic公式は試験注記と選択軸に強い。独自の接続は「夜間に除湿したい→送風ナイトを外す」「ホース必須→当該Panasonicを外す」「排水しても機種固有の停止がある」をメーカー横断で一つの購入条件表にすること。L/日や方式の説明を増やすことではない。

**シュレッダー：実測とは別の価値あり。** S21/S22の360LiFEには実物・騒音・使い勝手の優位があり、PROBEFOLIO／ころもちnoteにも定格、休止、原本差、紙分類がある。独自の接続は「auto前の金具・綴じの準備」「手差しとの別条件」「過熱・詰まり・途中再開の違い」から候補を保留すること。実測媒体より性能の根拠が強いとはしない。

両者ともB=4。具体的な判断ミスを減らす接続は根拠付きで成立するが、SERP gap3/5や要約比較の限界を超えて「他にない」「読者に勝つ」とは判定しない。表の数、文字量、一次資料の本数は加点根拠にしない。

## 8. Commercial Readiness

|過去snapshotの照合|除湿機（10/15）|シュレッダー（10/16）|
|---|---|---|
|公式identity|CV-T71、F-YEX90D、DH-C7200、IJC-P70|P5GCX2、SZK01、PSD-M200AT、16MC|
|availability1の代表出品|3/4：XPRICE、天六なごみ、roomy|3/4：ARIMAS、リコメン堂、くらしのeショップ|
|availability0|IJC-P70 / u-denki:10027672|PSD-M200AT / factory:10067065。唯一この検証でauto条件を確認したモデルの販売例が不可|
|SKU・送料|CV-T71は23,800～27,500円の価格幅、選択SKU unknown。Panasonic送料別、cado地域条件unknown|選択SKUは全例unknown。Aska送料無料titleと送料別flag不一致|
|型番・状態|suffix／色、セット、実物新品を再確認。販売表示だけで新品保証しない|旧P5GCX／X2、16MC／16MC-Rを混ぜない。セット・実物新品を再確認|
|正式affiliate|型番・販売条件確認後の発行は設計可能。未発行|同左。autoの現行販売導線は未確認|

Fは両者3。3出品という過去例から現在の在庫・商用優位を決めない。シュレッダーのauto販売例不足は実際のPreparation負荷・制約として残すが、今も売れないと推測しない。両者とも最終価格・送料・継続販売・正式料率・CTR・CVR・月間検索数・収益はunknown。snapshot価格を更新せず記事へ載せる前提にもせず、価格を省く代替も可能。今回affiliate発行0。

## 9. Maintenance / Seasonality

除湿機はD19でMaintenance Burden4。温度・湿度・周波数・mode・自動停止・ホース条件と改訂型番、価格／在庫を維持する必要がある。冬scopeは10/20時点の今後の季節と整合する編集上の理由があるが、公開日未確定で、検索季節性の大きさ・冬の需要増は未測定。

シュレッダーもS21でMaintenance Burden4。派生型番、取説v1.1などの改訂、最大／定格の定義、紙・金具・復帰条件と価格／在庫が対象。選定scope自体に特定季節の必須条件はないが、需要が通年一定とは実証していない。

**低運用コストの勝者は既存データでは決めない（E=3／3）。** 除湿機の季節限定scopeは観測対象を絞れる一方で温度・modeの確認が要る。シュレッダーは季節に依存しない構成にできても、素材・再開・版管理を省けない。更新回数、モデル寿命、保守時間はunknown。Gも3／3とし、カレンダー上の適合だけで季節ボーナスを与えない。

## 10. Preparation Workload

両候補とも、原本の対象箇所の再確認、完成稿編集、unknownと購入保留の一貫性確認、全文の独立claim audit、競合全文との読みやすさ確認、商品SKU／販売条件／正式広告URL確認、モバイル・アクセシビリティ・source mappingのQAが未実施。匿名要約評価を完成稿レビューで代替しない。

除湿機は4モデルの原本が揃い、D-U1/U2のunknownを解消できなければ保留を維持する方法が既に原稿にある。H=3（通常のPreparationと限定条件の確定が必要）。シュレッダーはこれに加え、中心のX2復帰・紙条件を固有原本で解決するか積極比較から限定するかというscope判断と、autoの販売例の再確認が残る。H=2（追加の不足が目立つ）。セル数や原稿長で工数を算出しておらず、時間・人日は見積もらない。ここでは必要作業の比較だけで、どちらの準備も実行していない。

## 11. Weighted Scores

得点は今回rootの編集判断。新たな独立judge評価ではない。点差は除湿機−シュレッダー、重み付き差は100点換算。

|軸|重み|除湿機|シュレッダー|得点差／加重差|既存根拠と不確実性|
|---|---:|---:|---:|---|---|
|A 購入判断の有用性|15%|4|4|0 / 0|D20判断表・S22判断表。双方、用途に合わない候補を外せる。実読者効果は未測定なので5にしない|
|B 具体的差別化|20%|4|4|0 / 0|D20 mode／排水／停止、S22 auto紙条件／原因別復帰。既存競合も強く、独占gapではない|
|C 一次根拠の充足|20%|4|3|+1 / +4|D20は4原本で中心mode・排水・停止を支持。S22はX2固有原本がなく中心の復帰・素材を肯定できない。unknown率ではなく判断の欠損深度|
|D 未解決リスクの少なさ|15%|3|3|0 / 0|両review2重大0、保留境界で管理可能。ただし排水・停止、細断・再開とも誤使用リスクあり。安全性の順位にしない|
|E 更新・保守負荷の低さ|10%|3|3|0 / 0|D19/S21ともmaintenance4、原本・版・mode／素材・SKU更新が要る。頻度・費用の優劣は不明|
|F Affiliate導線|10%|3|3|0 / 0|D20/S22 mappingで販売可過去例3、SKU／送料は未確定。Sanwa auto例不可、Iris除湿機例不可を保持。収益性未測定|
|G 時期・季節性|5%|3|3|0 / 0|D19季節性の強さunknown、S21まとめ処理scopeに季節指定なし。10/20と冬scopeは整合するが公開日・需要は未確定|
|H 追加作業の少なさ|5%|3|2|+1 / +1|D20の4原本の限定条件確定に対し、S22はX2中核原本／scopeとauto販売例を含む追加判断。全文QAは共通で未完|
|合計|100%|**71/100**|**66/100**|**+5**|編集上の順番決定の補助。市場点の1点差を換算していない|

private `score-input.json` を計算前に保存。Goで整数の分子355／330を計算し、最後に5で除した。0点・全5点、範囲外、重み合計、5点境界のfixtureもPASS。算術は決定的だが入力評点が客観的な収益期待値になるわけではない。

## 12. Sensitivity / Tie-break

差5点は事前ルールで **実質接戦**。単純な点数勝ちでは選ばない。第1タイブレーク「中心判断を支える証拠」で除湿機を優先する。4機種の送風／除湿、ホース可否、自動停止という主要な除外判断を型番別原本で追える。一方、シュレッダーはX2の固有原本不足が購入判断の中心である復帰と素材に同時に及ぶ。3機種に限定すれば価値は残るが、そのscope判断を先に要する。

この差はセルの件数でも、修正履歴にmajorが一度あったかでもない。D20のmajorは適切に修正済みで罰点にせず、現在の原本でどこまで言えるかで比較した。第2軸の差別化は両者にある。第3軸の無条件中心blocker群数は0／0で決着根拠にせず、第4軸の保守負荷も勝者不明。第1で根拠付きの優先順が付いたため、強引に後続軸の差を作らない。

1評点の100点換算影響はA～H順に3／4／4／3／2／2／1／1点。単一軸の片方を1点変更しても数値上の先行は逆転しないが、差は最大4点縮む。CとHの評価を併せて下げれば数値差は消える。新しい重みの最適化や採点やり直しは行わず、**根拠の限定に依存する接戦の編集判断**と解釈する。X2原本取得等の新証拠で将来の優先順は変わり得るが、今回の選定記録は書き換えない。

## 13. Final Selection

**SELECT_DEHUMIDIFIER**。優先scopeは「冬の部屋干し用除湿機を、室温・運転モード・定格条件・排水条件で比較する」。中心の型番別除外判断を支える原本と、保留にすべき限定条件が既に揃っており、最初の準備工程に進める根拠が相対的に強い。

**公開承認ではない。** 両Content GOを維持。SEO成功、利益、乾燥・騒音・安全性の優位は確認していない。D20の当時の「ここで停止／STRONG GOでないため準備を提案しない」は履歴として維持し、今回の明示依頼による新たな候補選定と区別する。過去GOを昇格させて先へ進んだことにはしない。

## 14. Reserve Candidate

シュレッダーは **VALIDATED_RESERVE_CANDIDATE**。market GO28/35、content GO、独立比較と監査結果はそのまま有効な履歴。非選定をHOLD／NO-GOへ変更しない。紙の準備と原因別復帰は独立した編集価値があり、X2原本不足を補えたと仮定して優先順だけ上げることもしない。今回はreserveのPreparationを提案・実行しない。

## 15. Remaining Blockers

選定を止めるFAILはないが、選定候補を今すぐ公開できる状態ではない。除湿機の公開前には次が必須。

1. D-U1/U2を「追加根拠で解消」又は「unknown／購入保留・肯定比較からの限定」で処理する方針を完成稿で確定。解消できたと仮定しない。
2. 27℃60%・周波数・modeと、ナイト送風、音ひかえめの保護制御、満水とauto-off／連続設定の違いを完成稿でも保持する。
3. 現行資料の原本位置・型番・改訂、商品SKU／新品・セット／販売・送料条件と、使用する場合の正式affiliate先を別工程で確認する。
4. 完成稿の独立claim auditでcritical/major 0、source mapping、読みやすさ・mobile・広告表示を検証し、別の最終公開可否判断を得る。

上記は今回の選定作業では未実施。全文・表示・広告導線の確認を済ませたという意味のREADYは付けない。reserve側の未解決は§6/8に記録したまま保持する。

## 16. Next Stage

次の独立工程として **除湿機 Publication Readiness / Preparationのみ**を提案する。実行順：既存scopeと採用モデル／肯定比較の境界を固定→必須原本箇所と未確認条件の処理を確認→完成稿を編集→競合全文との可読性確認→全文の独立claim audit→必要な商品情報・広告リンク照合→非公開previewでmobile／SEO条件／source traceabilityを検証→公開可否を明示的に判断する。

根拠が取れない場合はD-U1/U2の保留を維持した限定scopeへ戻す。中心判断が弱くなるならPreparationをHOLDできる。SELECTを根拠に公開、affiliate発行、sitemap／GSC／deployへ自動接続しない。今回はこの次工程を一切実行せず停止する。

## 17. Limitations / Reproducibility

この選定はrootの一回の資料比較で、新規独立LLM審査ではない。既存judgeは異なる競合・要約・日付の別実験であり横断的な優越を実証しない。76対144セル、unknown率、調査量で品質比較をしていない。元の原本監査も対象ページ・保存資料の範囲に限定され、今回全原本の安全認証を再実施したわけではない。

private ignored `outputs/publication-candidate-selection/` にprotocol、固定commit文書コピー／blob IDs、参照入力hash、score input、Go計算とfixtures、score output、実行manifestを保存。算式と全評点は本書§11にあり再計算可能。計算例：除湿機=(15×4+20×4+20×4+15×3+10×3+10×3+5×3+5×3)/5=71。シュレッダーはC=3、H=2として66。private計算は既存出力の上書きを拒否する。

新規Rakuten／Jev／Claude API calls0、新規独立judge0、追加外部LLM API費用$0。通常の本Codex作業のtoken数・追加課金額はunknown。市場・本番・Analytics・Price Radar schedulerへの操作0。公開Gitへ追加するのは本書のみで、他branchの文書／code／原本／raw dataは取り込まない。

検証結果（2026-10-20 JST）：`go test ./...` PASS、`go vet ./...` PASS、`git diff --check` PASS。private Go計算fixturesもPASS。base比の`site/`、`content/published/`、既存`internal/`／`cmd/`、Wrangler／workflow／D+7 protocolの差分0。新規tracked差分は本書のみ。参照4branchのHEAD・既存artifact hashの不変を確認し、PR #19～#22の判定・採点・protocolを変更していない。
