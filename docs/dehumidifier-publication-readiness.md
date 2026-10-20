# Dehumidifier Publication Readiness / Preparation

## 1. Objective

**CONDITIONAL — 完成稿・独立監査・非公開表示は完了。本番で2記事を共存させる生成・参照の確認を公開前条件として残す。公開はしていない。**

2026-10-20 JST、除湿機prototypeを完成Markdownとローカル非公開HTMLに仕上げ、全文監査と表示・技術確認を実施した。これはSTRONG GOへの昇格でも公開承認でもない。本番公開には別工程・別ユーザー承認が必要。

branch `dehumidifier-publication-readiness`、base `dehumidifier-content-validation` / `dbf95a3eeae0dfc0e810d52906fa8d4306bc36c1`。PR #19/#20を直接継承し、PR #23の選定記録はcommit `e14de43` からread-onlyで参照した。選定branchをmergeしていない。D+7 CONTINUE、market29/35 GO、content GO、SELECT_DEHUMIDIFIER、シュレッダーreserve、互換フィルターHOLDは変更しない。

編集・追加確認前にprivate `protocol.md` v1を固定した。scope、中心claim、unknown処理、独立全文監査、商品取得、390/1440px表示、READY／CONDITIONAL／HOLD条件を明記。READYは重大指摘0、公開上のminor0、unsupported0、中心根拠と追跡性、blocking0、商品導線検証又は安全な除外、正常な非公開previewを要求する。中心根拠不足が安全に限定できなければHOLD。結果による基準変更はない。

## 2. Fixed Scope

「冬の部屋干し用除湿機を、室温・運転モード・定格条件・排水条件で比較する」。CV-T71、F-YEX90D、DH-C7200、IJC-P70の4機種を維持し、新しいモデルは加えていない。全機種を冬用途に推奨する稿ではなく、室温・夜間の実動作・ホース・停止条件から候補を外す／保留する記事。

原稿タイトル：**冬の部屋干し用除湿機を、室温・モード・排水条件で比べる**。性能順位、乾燥時間、最も省エネ／静か、タンク量から満水時刻、年間電気代、健康・カビ・除菌保証は追加していない。実機試験は行っていない。

## 3. Reused Evidence

`outputs/dehumidifier-content-validation/` のprototype-v2全文、76セルmatrix-v2、22claim、原本4PDFと抽出・視認画像、source manifest、独立監査1/2、楽天mapping、judge3名の結果・匿名要約・競合比較が現在環境に存在することを確認して再利用した。存在を推測して再現扱いにした資料はない。過去評価の再採点・上書きはしていない。

今回のprivate folderは `outputs/dehumidifier-publication-readiness/`。公式PDF4本と既存の公式Web5件、追加のPanasonic便利機能／Iris製品情報を取得・保存。全11件HTTP200。PDF4本は10/16の監査済み版とbyte同一。HTML取得と内容照合を区別し、引用する箇所だけを確認した。Panasonic比較表のF-YEX90D列と連続排水「－」の対応は独立reviewerも保存HTMLから再確認した。

追加の公式FAQ11834は別に保存したが、対象例・方式の違いがあるため当該型番の上限根拠には採用していない。秒単位の取得日時・hash・URL・取得結果は `source-refresh.json`、`source-manifest-v3.json` に保持。PDF紙面と物理頁の区別は原本で確認した。メーカー問い合わせは送っていない。

## 4. Unknown Resolution

|ID|今回の追加確認|最終処理|
|---|---|---|
|D-U1 IJC-P70の全運転温度|公式サポート、そこからの取説103701、公式製品情報、対象を絞った公式検索。15～30℃以外の全範囲は確定できず|unknown維持。15～30℃は効率設計の説明で、使用可能範囲にも15℃未満禁止にも転用しない。必要な冬の室温への適合が不明なら購入保留|
|D-U2 F-YEX90D連続設定の上限|同一取説10／19頁を再視認。公式便利機能ページは自動停止の最長約15hを確認。一般FAQの24hは別方式の文脈|unknown維持。自動停止と連続設定を分離。約15h／24hを連続設定へ流用せず、無期限運転を保証しない。必要な連続条件が未確認なら購入保留|
|関連 IJC-P70定格測定モード|既存原本・公式仕様の確認範囲|未確認のまま。弱・強・おまかせという実動作と、定格値を測ったモード名を混同しない|

参考：[Iris公式サポート](https://www.irisohyama.co.jp/products/support/4967576654197)、[Panasonic便利機能](https://panasonic.jp/joshitsu/feature/useful.html)、[対象例・方式の異なる一般FAQ](https://jpn.faq.panasonic.com/app/answers/detail/a_id/11834/)。一般FAQからF-YEX90Dのホース対応や24h停止を導いていない。取得不能を0や非対応へ変える処理もない。

## 5. Updated Evidence Matrix

既存76セルをmatrix-v3、監査修正後matrix-v4として別保存。最終内訳はverified64／qualified10／unknown1／not applicable1で、旧版と同じ。qualifiedの中にも未確定の部分条件があり、status件数は全仕様の確定率ではない。

変更は3セル。`evidence-deltas.json`／`evidence-deltas-v2.json` にold/new value、source、checked_at、reason、confidence、statusを保存した。

|model / field|旧値の要旨|新しい版での変更|根拠／status|
|---|---|---|---|
|IJC-P70 / operating_temperature|全範囲unknown、15～30℃を転用しない|公式追加確認でも確定できなかったことと保留境界を明記|IM2、公式support/product確認／unknown|
|F-YEX90D / auto_stop_timer|自動最長約15h、連続上限unknown|一般FAQの他方式24hも転用しないと明記|PM10/19、追加公式確認の範囲／qualified|
|CV-T71 / drainage_conditions|径・下り勾配・タンク・凍結／水没・点検|原本13頁ホース図の10cm以上の高低差も照合する条件を追加|SM13図を視認／verified|

残り73セルの意味を変更していない。他モデル／seller値で補完していない。10cmの条件はSharpだけの図条件で、cadoの20cmやIrisの排水条件へ転用していない。

## 6. Final Article Structure

private `article-final-draft.md` は **Markdown全4,271文字**。別指標として、Markdownリンク記法・claim ID・空白を除く文字数3,282（リンク文字も除く機械的定義）。H1×1、H2×7、表4個、メーカー資料リンク9件。内部DD01～DD22を維持した。

構成：短い導入と読順→定格条件→低温の動作→夜間モード→ホース／停止→6条件の購入判断表→型番・選択内容→調査範囲／原本。モード名・型番・周波数・27℃60%を保持し、測定規格まで完全一致とは断定しない。定格値を冬の実性能へ換算していない。

判断表は「低い室温」「夜間除湿」「ホース必須」「長時間」「狭い設置場所」「水捨て軽減」から、原本の照合対象と不明時の保留へ接続する。F-YEX90Dをホース必須から外し、DH-C7200ナイトを除湿目的で選ばない。CV-T71の除湿／衣類消臭、約12h停止・解除の原本参照、Irisの弱・強・おまかせ／24h無操作停止も残した。

prototype-v2→draft-v1→v2（編集9点）→v3（Sharp図条件）→v4（表の列案内）→final同一byteという履歴を保持。差分・理由はeditorial-changes-v2、audit-fix-history、v3-v4記録、prototype-to-final.diffに保存。最終SHA256：`b15b4b3f0648a965d63452a32a7a2087057174b5e65b7acfa36fe445a2533e47`。

## 7. Full-text Review

独立編集reviewerがdraft-v1全文を読み、固定競合3ページ（[Panasonic](https://panasonic.jp/joshitsu/select.html)、[LiviqLens](https://liviqlens.com/dehumidifier-capacity-drain-humidistat/)、[どれ買う？](https://dorekau.net/dehumidifiers/)）の取得した本文全体と比較。画像内文字・動的UI・リンク先全文は未確認と記録した。検索要約だけの比較でも、視覚的な全ページの検証でもない。

導入に読む場所の案内、L/日の単位、周波数順、排水表の読み方、SKUの日常語化など9点を反映。独立review2はv3全文と修正対応を再確認。残る「左列」の曖昧さはv4で列名へ修正し、事実監査もv4を対象にした。長い排水セルは表の横スクロールで対応し、安全上の必要な条件は削っていない。一般方式論への脱線はほぼなく、実際に使う条件から候補を保留する役割は明瞭との所見。

原稿の長所は型番別の例外を追えること。実測の乾燥速度・騒音・寿命を選ぶ読者の最終決定までは支えない。過去の匿名要約scoreと今回の全文レビューは分離し、scoreや既存Content GOへ統合していない。

## 8. Independent Claim Audit

|段階|対象|critical|major|公開上のminor|unsupported残存|
|---|---|---:|---:|---:|---:|
|初回|draft-v1全文、全表／判断表、22claim、76セル|0|0|1|0|
|修正後|v3→v4／final全文、claims-v2、matrix-v4、初回指摘差分|0|0|0|0|

初回CA-001はCV-T71の排水要約に原本13頁図の10cm条件が抜けていた点。図を再視認して補足し、別の修正確認として解消を記録した。初回履歴は保存。各claimの該当本文、source/type、紙面・位置、direct/inference、confidence、risk、statusは `claim-audit-v1.json`／`claim-audit-v2.json` にある。

監査はauthorと別のセッションで、過去GO／scoreを与えず実施。原本4PDF、現行HTMLと楽天保存値の照合、対象38ページの追加render／視認を行った。再監査は最終原稿のhashを記録し、finalとの差異0を検証した。未確認3項目はunknownのままで、根拠なしの断定を弱い語尾に変えただけの解消ではない。監査対象外の実機安全認証を意味しない。

source/claim join、22claimの原稿残存とpreview内trace、76セル重複／欠損／source ID、unsupported使用、全中心原本の参照を検査しPASS。読者画面のDD表記は隠し、`data-claim-ids` とprivate `claim-map.json` に内部対応を保持。末尾の原本リンクと紙面案内からメーカー資料へ戻れる。

## 9. Rakuten Product Refresh

2026-10-20 JST、既存4出品をitemCodeで各1論理request、4成功・取得失敗0。既存Go client、serial interval1.1s、timeout20s、最大追加retry2を使用。raw応答とAPI取得時刻はprivate `rakuten-refresh.json` に保存。実際の内部retry回数は計測していない。大量検索・他機種追加・resolver v2採用はしていない。

|モデル / shop / source_id|API価格／min–max|availability／送料flag|公開への扱い|
|---|---|---|---|
|CV-T71 / XPRICE / a-price:10849817|23,800円／23,800–27,500|1／別|可変価格・選択SKU未確定|
|F-YEX90D / 天六なごみ / nagomi-pharmacy:10024835|80,080円／同値|1／別|選択SKU・新品実物・セット・最終送料未確認|
|DH-C7200 / roomy / roomy:10015559|59,950円／同値|1／込|特典付きtitle、選択色・構成・地域送料未確認|
|IJC-P70 / ウエノ電器 / u-denki:10027672|19,800円／同値|0／込|この出品を購入可能と表示しない|

API itemCodeと4公式型番のtitle joinはPASS。source_idは選択SKUではなく、新品表示不在／中古語不在だけで実物新品とはしない。API価格、クーポン／ポイント後、送料、可変SKU価格を混ぜない。上表も取得時点の記録であり、今後の現価・在庫保証ではない。

## 10. Affiliate / Disclosure Validation

**今回の完成稿は商用導線を安全に除外した版。** 選択SKU・構成・送料の確認が完了したと仮定せず、4機種すべてについて価格・楽天購入リンク・affiliateリンクを本文／previewから外した。正式URL発行0、推測によるURL組立0、広告クリックテスト0。メーカー資料リンクだけを残す。

「商品リンクについて：この準備稿には販売価格・購入リンク・広告リンクを掲載していません」とpreview／footerで説明。広告があるという誤った開示は出さない。広告リンク0件なので `rel=sponsored` の実リンク検証はN/Aであり、検証済みとは記録しない。将来広告を加える場合は、公式発行・型番／選択内容・表示・rel属性を別途確認し、その版を再監査する。このREADYを広告導線readyへ読み替えない。

## 11. Private Preview

`outputs/dehumidifier-publication-readiness/preview/index.html`。既存 `internal/canary.Render` をprivate Go wrapperで再利用し、生成先だけをprivateへ。旧記事専用のdescription／広告開示はpreview内で除去・差し替え、現在の `site/assets/site.css` をそのままコピーした。wrapper追加の小CSSはpreview限定の表幅・focus・tap target・header整列のみ。framework、外部font、画像、JavaScriptは追加していない。

serverは `127.0.0.1:8791` のloopbackのみ。公開URL・Cloudflare deploymentはない。HTML robotsと全HTTPレスポンスのX-Robots-Tagはnoindex/nofollow/noarchive。canonical・公開日・Article structured data・Analytics beaconは出力しない。robotsは補助で、非公開性の主たる根拠は外部公開しないloopback／ローカルファイルである。原本／raw応答はserver root外。

1本目のローカル表示比較用コピーはscript／canonicalを除去したprivate copy。本番記事は変更していない。未公開本文とQA画像はGitに追加していない。

## 12. Desktop / Mobile QA

in-app browserでDesktop1440×1000、iPhone相当390×844を確認。実端末での操作検証ではない。冒頭、4表、排水表の横スクロール、出典、footer、1本目との組版の統一をスクリーンショットへ保存した。

|確認|結果|
|---|---|
|本文|白、system sans、1カラム。本文16px、Desktop H1 32px、Mobile H1 22px。700pxの本文幅|
|Desktop|document幅=client1425px、本文700px、4表はいずれも700px内で横スクロール不要|
|Mobile|document幅=client375px（viewport390pxのscrollbar分を除く）。本文の横はみ出しなし。表container335px／table640px|
|表の操作|各表のrole/label、caption、thead、列th13／行th18、tabindex0、focus outlineを確認。ArrowRightによる横スクロールを実動確認|
|リンク|メーカー資料9件のtap高さ44px。緑の下線、focus、skip linkあり。長い型番は折り返す|
|広告／装飾|購入CTAなし。図は今回の条件照合に必須でなく追加0。hero／新しいデザインなし|

初回QAでheader間隔とDesktop表の不要な横スクロールを発見し、private CSS内で修正した。初期画像も保存し、最終画像と区別する。画像保存名のfinalは採用稿と最終CSSの確認用。screenshots／DOM測定値はprivateに保存。

## 13. SEO / Technical QA

title／H1は§2、description候補は「冬の部屋干し用除湿機4機種を、室温・モード・定格条件・排水と停止条件で照合する購入前の確認記事。」。H1×1→H2×7の構造で飛躍なし。予定canonical候補は `https://choicelen.page/articles/winter-laundry-dehumidifier-conditions/` だが、未公開・未確定の記録に留めHTMLでは宣言していない。

内部参照は本文skip、広告説明anchor、local baseline、CSS、claim mapを照合。外部出典9件はmanifestの公式URLと一致し、対応保存取得はHTTP200。静的HTML内の参照整合、localhost HTTP200、noindex header、script／Analytics0、購入先0、canonical0を検査した。未公開のためdatePublished・Article schemaは省略。公開工程でURL・日付・公開用robots／canonicalを正式設定する必要があるが、今回は設定を本番へ送っていない。

## 14. Remaining Issues

|ID|分類／対象|処理・根拠・検証状態|
|---|---|---|
|R-U1|CONDITIONAL：DD07/IJC全温度|unknownと必要時購入保留を本文／判断表で維持。IM2、独立監査2で限定方法を確認済み。追加根拠で解消済みではない|
|R-U2|CONDITIONAL：DD19/Panasonic連続上限|自動停止15hと別。必要条件未確認なら保留。PM10/19と独立監査2で境界確認済み|
|R-U3|CONDITIONAL：DD02/IJC定格mode|未確認を表に明示、完全同条件の性能順位なし。IM15・監査2確認済み|
|R-C1|CONDITIONAL：全機種の販売SKU／状態／送料／正式広告|今回の公開候補版では価格・購入／広告リンクをすべて除外。source/preview検査確認済み。広告を戻す版には未解除の必須確認|
|R-M1|NON-BLOCKING：資料・販売情報の将来更新|資料確認日を明示。原本改訂・型番・後日の商品状態は公開／更新時に再確認。現在状態の保証はしない|
|R-E1|NON-BLOCKING：実性能・最適機種|実機未試験と限界を明示。速度・音・費用順位を足さない。解消は本scopeに不要|
|R-T1|公開工程の必須条件：承認／正式URL／公開日／本番設定|この判定で実行してよい操作ではない。別工程でユーザー承認を得て、公開用技術要素・広告有無を確認する|
|R-T2|CONDITIONAL：本番の2記事共存・生成と参照|既存build-choiceLenは1本目のfrozen hash／固定pathを要求する専用builder。今回のprivate wrapperは本番用の複数記事builderではない。公開工程で1本目を保持した2記事の出力、内部リンク／canonical／sitemap／robotsの整合を検証する必要がある。今回は未実施|

**現限定稿の未解消claim BLOCKING0。** R-U1～3／R-C1は「unknownを解消した」ではなく、本文範囲・購入保留・商用導線除外という対処方法を検証済み。R-T2の本番用生成・参照確認は未完了の公開前条件で、公開はそれが完了するまで不可。限定を外した肯定推奨や広告付き版は、この判定をそのまま使えない。今回の原本監査初回minor CA-001は解消済みとして別履歴に保持した。

## 15. Readiness Decision

**CONDITIONAL**。完成稿、22中心claimの追跡、修正後critical/major/公開minor/unsupported0、安全な商品・広告導線除外、非公開previewの表示・技術確認は完了。ただし、ローカルpreviewの正常表示を本番用builderの2記事共存確認と同一視しない。R-T2を未確認の公開技術条件として残すため、全面的なREADYは出さない。

原本条件と購入保留を示す、商用リンクのない稿としての品質は成立した。冬の適合・乾燥・省エネ・静音の性能保証、4機種すべての購入推奨、affiliate収益化readyではない。旧Content GOは変えず、公開技術条件の再確認と公開承認を別工程へ残す。新しい評点・閾値を後付けしたのではなく、固定した「公開用技術要件に問題なし」をまだ確認しきれていないための区別である。

## 16. Publication Requirements

別工程でのみ、ユーザーの公開承認、最終URL／公開日・資料確認時点、scope・unknown境界の維持、公開用canonical／robots／内部導線を確認する。広告を入れる場合はその前に正式URL・選択SKU・販売条件・広告開示／sponsoredを検証し、完成稿とpreviewの変更部分を再監査する。広告なしを選ぶことも可能。

公開後のHTTP／HTTPS／canonical／sitemap／Analytics等の確認も公開工程の責務。今回、site／content/published、sitemap、GSC、Cloudflare、Analytics、既存affiliate、Radar schedulerを変更しない。READYでもdeploy／merge／tagは行わない。

## 17. Reproducibility

private protocol、source-refresh／manifest-v3、原本4PDF・抽出・対象page画像、matrix-v3/v4／delta、draft-v1～v4／final、編集review1/2、claim-audit1/2、楽天refresh、preview、QA PNG／DOM測定、checks／final manifestをignored folderへ保存。最終原稿のhashは監査記録と一致する。機密値・raw account情報・tokenはreport／Gitへ出していない。API credentialは既存ignored envを読み込み、値は表示していない。

private Go toolsはprepare、refresh、edit、revise、preview、check。preview生成は既存canary rendererを再利用、checkはsource/claim join、unsupported、76セル、22claim、変更履歴、監査hash、9公式参照、4商品型番joinとloopback参照を検査する。private入力がないclone環境で再現済みとは扱わず、このfolderの復元が必要。公開Gitは本reportだけで、未公開本文・raw・原本は含めない。

追加Jev／Claude有料API calls0、直接追加LLM API費用$0。OpenAI/Codexの独立claim reviewer初回＋再確認、全文編集review初回＋再確認を分離して保存。初回claim reviewerは利用上限で一時中断し、ユーザーの再開後に完了した。中断前に確認済みの取得・描画を繰り返して評価を作り直していない。exact model／tokens／Codex追加課金額はunknown。新しいjudge採点や過去GOへのscore統合はない。

検証結果：`go test ./...` PASS、`go vet ./...` PASS、`git diff --check` PASS。source/claim join、unsupported0、missing evidence0、76セル／22claim、最終監査hash一致、prototype変更履歴、private preview参照、4商品型番joinもPASS。親branch比でsite／content/published／既存internal・cmd・README・Wrangler・workflow・既存評価文書の差分0。今回のtracked差分は本reportのみ。
