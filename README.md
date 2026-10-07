# jev-money-engine

## Foundation validation

固定3カテゴリの検証と非公開価格履歴の運用方法:

- [Family rulebook](docs/family-rulebook.md) / [Family検証結果](docs/family-resolver-validation.md)
- [Sampling protocol](docs/sampling-validation-protocol.md) / [Sampling bias検証結果](docs/sampling-bias-validation.md)
- [Price Radar履歴・認証・schedulerの制約](docs/price-radar-history.md)

候補Family Resolver v2は実験専用で、本番へ置換していません。raw sample、gold label、API response、SQLite、秘密情報はGitへ保存しません。CLI flagsは `go run ./cmd/validate-foundation -h` と `go run ./cmd/radar-history -h` で確認できます。実験成果をChoiceLenへ自動公開しません。

個人向けの商品候補収集・選別システム。Day 1は楽天の商品取得・SQLite保存・一次フィルタ、Day 2は既存候補のJev評価、Day 3は固定Jev GateとLLM reranker・blind AI reviewの比較実験です。Web UIや記事・SNS生成はありません。

## 必要な環境とセットアップ

- Go 1.26以上、ネットワーク接続
- 楽天Web ServiceのApplication IDとAccess Key
- SQLiteはpure Goドライバ `modernc.org/sqlite` を使用。Python、Cコンパイラ、SQLiteの別途インストールは不要

```sh
git clone --branch day1-collector https://github.com/AkiraShimizu0/rakuten-product-search.git jev-money-engine
cd jev-money-engine
go mod download
```

Day 1のコードは現在 `day1-collector` ブランチのdraft PRにあります。mainへマージ後は `--branch day1-collector` を省略できます。

Windows用 `collect.exe` / `inspect.exe` を受け取った場合は、同じオプションで直接実行できます。実行時のGoインストールは不要です。

```powershell
.\collect.exe -keyword "空気清浄機" -pages 34 -sample 20
.\inspect.exe -sample 20
.\inspect.exe -state "shop:123456"
```

ソースからWindows用バイナリを作る場合:

```powershell
go build -o collect.exe ./cmd/collect
go build -o inspect.exe ./cmd/inspect
```

`.env.example` を `.env` にコピーし、ローカルで設定してください。

```dotenv
RAKUTEN_APP_ID=your_application_id
RAKUTEN_ACCESS_KEY=your_access_key
RAKUTEN_AFFILIATE_ID=your_optional_affiliate_id
RAKUTEN_ORIGIN=your_optional_registered_origin
```

Affiliate IDとOriginは任意。Originは楽天アプリの登録設定で要求される場合のみ指定します。OSの環境変数が設定されていれば `.env` より優先されます。値はログに出力しません。`.env` は単純な `KEY=value`、引用符付き値、コメント行に対応し、シェル展開は行いません。別ファイルは `-env-file /path/to/rakuten.env` で指定できます。

認証情報と取得データはGit管理対象外です。認証情報をソース、コマンド引数、コミットへ書かないでください。

`HTTP 403 CLIENT_IP_NOT_ALLOWED` の場合は、楽天Web Serviceの該当アプリ設定の許可IPに、実行PCの接続元グローバルIPを追加してください。IPが変わった場合も更新が必要です。再試行ではこの設定エラーは解決しません。

## 商品取得

```sh
go run ./cmd/collect -genre 123456 -pages 20
go run ./cmd/collect -keyword "空気清浄機" -pages 34 -sample 20
go run ./cmd/collect -genre 123456 -keyword "掃除機" -pages 34 -sample 20
```

`123456` は例なので実在するジャンルIDへ置き換えてください。`-genre` または `-keyword` が必須です。ジャンルIDは [楽天ジャンル検索API](https://webservice.rakuten.co.jp/documentation/ichiba-genre-search) で調べられます。

現行の [楽天市場商品検索API 2026-07-01](https://webservice.rakuten.co.jp/documentation/ichiba-item-search) を使用。Access KeyはHTTPヘッダに送信します。1ページ最大30件、ページ番号1〜100。34ページで最大1,020件、100ページで最大3,000件です。検索結果が少ない場合は最終ページで終了します。最大件数を超える探索は別キーワード・ジャンルで実行し、DBで重複排除してください。

```sh
go run ./cmd/collect -keyword "空気清浄機" -start-page 35 -pages 10
go run ./cmd/collect -keyword "空気清浄機" -pages 34 -sort=-reviewCount -interval 1500ms -timeout 30s -retries 4
```

既定は `sort=standard`、試行ごとのHTTP timeout 20秒、リクエスト開始間隔1.1秒（CLIでは1秒未満を拒否）。429、一時的な5xx、ネットワーク障害は最大4回追加試行（合計5回）、2/4/8/16秒のbackoff。`Retry-After` の秒数・HTTP日時を尊重し、2分を超える指定は待たずエラーで終了します。4xxの認証・入力エラーは再試行しません。途中失敗・Ctrl+C時はコミット済みページを残し、部分集計を表示して非ゼロ終了します。最終コミット済みページを確認して `-start-page` で再開できます。同一URLの短時間の繰り返しは楽天側で制限される場合があります。

## DB・モデルと一次フィルタ

DBはプロジェクトルートからの相対パス `data/money.db`。`-db` で変更可能。Day 2では `PRAGMA user_version=2` に自動移行し、既存productsを保持して評価テーブルを追加します。`data/` 全体はGitから除外されます。

`products` テーブルの `(source, source_id)` が主キー。再取得は更新となり、`first_seen_at` は保持、`last_seen_at` は更新します。日時はUTC。同一実行内の重複は最初の出現のみ処理。保存はページごとのトランザクションです。除外商品も保存するので再取得なしでフィルタを変更できます。元の個別商品レスポンスを `raw_json` に保持し、未知フィールドも残します。

一次フィルタの既定条件:

- 価格3,000〜50,000円（両端を含む）
- レビュー5件以上
- 説明80文字以上（UTF-8バイト数ではなくUnicode文字数）
- 名前・ID・URL、および必要項目の欠損・不正値がない

```sh
go run ./cmd/collect -keyword "空気清浄機" -pages 34 -min-price 2000 -max-price 60000 -min-reviews 3 -min-caption 60 -sample 20
go run ./cmd/inspect -min-price 2000 -max-price 60000 -min-reviews 3 -min-caption 60 -sample 20
```

フィルタは保存済みの固定eligibleフラグに依存せず、その実行で指定された条件で評価します。欠損した数値は0と `missing_fields` で記録し、欠損と実際の0を区別。ID欠損・壊れた商品JSONは当該商品だけスキップして件数を表示。説明はHTMLタグ・実体参照・余分な空白を正規化し、原文は `raw_json` に残します。

## レポートとサンプル

collectorは以下を表示します: Collected（APIから受け取った件数）、Unique（その実行で保存した重複なし件数）、Inserted、Updated（既存レコード再取得）、Eligible、Filtered out、平均価格、中央値、上位ジャンルID、欠損件数、除外理由。除外理由は複数該当するため合計が除外件数と一致しない場合があります。価格統計・上位ジャンルは当該実行のUnique全体が母集団（欠損価格の0も含む）。

```sh
go run ./cmd/inspect -sample 20
```

`inspect` はAPI認証なしでDB全体を再集計。サンプルはeligible商品のみから均等に重複なしで抽出します。20件未満なら全件。名前、ID、価格、レビュー数・平均、説明冒頭160文字、URLを表示します。

商品IDはショップ掲載単位。同じ製品が異なるショップに存在しても別商品です。ジャンルは名称ではなくIDです。価格は送料・ポイント・SKUバリエーションを考慮した実質価格ではありません。Affiliate ID指定時にはAPIの `itemUrl` もアフィリエイトURLになる場合があるため、URLを手動で書き換えずレスポンスを保存します。

## Jev用state

サンプルに表示されたIDを指定します。

```sh
go run ./cmd/inspect -state "shop:123456"
go run ./cmd/inspect -export > data/eligible-states.jsonl
```

`-state` は任意の商品（除外商品も含む）のJSONを表示。`-source` の既定は `rakuten`。`-export` はDB全体から現在のフィルタを通過した商品のstateをJSON Linesで出力し、stdoutにはJSON以外を混ぜません。PowerShell 7では次のようにUTF-8で保存できます。

```powershell
go run ./cmd/inspect -export | Set-Content -Encoding utf8 data/eligible-states.jsonl
```

stateはversion、source/source_id、商品名、価格、カテゴリID、説明、レビュー数・平均、ショップ名、URLで構成。数値比較・説明長・重複処理はGo側で完了します。

Day 2では `internal/jev/state.go` を入力境界として評価クライアントと別の評価結果保存処理を追加してください。商品説明は外部由来のデータとして扱い、Jevの指示とは分離します。API応答やDBモデルをJevへ直接渡す必要はありません。

## テスト

```sh
go test ./...
go vet ./...
go build ./cmd/collect ./cmd/inspect
```

unit testは正規化、欠損値、閾値境界、日本語説明長、SQLite upsert・ロールバック・再オープン、state、HTTP認証・429/5xx retry・上限・timeout・キャンセル、集計・重複なしサンプルを検証します。通常テストは実APIを呼ばず、秘密情報不要です。

実API検証は34ページの収集、同じコマンドの再実行、`inspect -sample 20`、`inspect -state` を実行してください。データは時間とともに変動するので再実行のUniqueが完全に同一になるとは限りません。DB主キーにより同一IDは増えません。

### 実データでの検証結果（2026-10-04）

「空気清浄機」、standard順、34ページを2回取得しました。

| 検証 | Collected | Unique（実行内） | Inserted | Updated | Eligible（実行内） |
|---|---:|---:|---:|---:|---:|
| 初回 | 1,020 | 1,020 | 1,020 | 0 | 395 |
| 2回目 | 1,020 | 1,019 | 0 | 1,019 | 394 |

DB全体は1,020件、eligibleは395件、重複主キーは0件。2回目にはページ間重複が1件あり、実行内Uniqueが1件少なくなりました。全1,020件のfirst_seen_atを保持し、再取得できた1,019件のlast_seen_atが進んだことをDBで確認しました。実商品のランダム20件表示、個別state表示、395件のJSONL出力も確認済みです。

説明欠損14件。キーワード検索には本体・交換フィルター・周辺商品が混在し、同一機種の別ショップ掲載や宣伝文の繰り返しもあります。商品名のクーポン価格がAPI価格と異なる例があり、Day 2の評価時には掲載文の数値を確定価格と混同しないようにしてください。Affiliate ID未設定での検証のためaffiliate_urlは全件空です。データと秘密情報はGitに含めていません。

## 構成

```text
cmd/collect/       楽天収集CLI
cmd/inspect/       DB集計・サンプル・state出力CLI
internal/product/ データソース非依存Product
internal/rakuten/ HTTP clientと楽天専用正規化
internal/store/   SQLite保存
internal/filter/  決定的な一次フィルタ
internal/jev/     Jev state・質問・HTTP・応答検証・スコア
internal/evaluate/ 評価実行・ランキング・人間レビューCSV
cmd/evaluate/    Day 2評価CLI
internal/report/  集計とサンプリング
internal/config/  ローカル環境設定
internal/cli/     共通フラグ
data/             実データ・DB（Git対象外）
```

将来のcollectorは専用のレスポンス型と正規化を追加し、共通Productに変換してstore/filter/reportを再利用できます。Day 1では他ECサイトや汎用collector frameworkを追加しません。

## Day 2: 固定候補のJev評価

Day 1の「空気清浄機」eligible 395件を使います。evaluateは楽天APIを呼びません。Day 2ブランチを使う場合は `git switch day2-jev-evaluation`。このdraft PRのbaseは `day1-collector` です。

### 公式APIと評価方法

2026-10-04に確認した[公式API](https://docs.typesafe.ai/api)は `POST https://api.typesafe.ai/v1/systemone`、Bearer認証、`model/state/questions` です。[モデル仕様](https://docs.typesafe.ai/models)に合わせ `jev-1.13.0` を固定します。`jev-latest` は更新され得るため比較実験では固定モデルを推奨します。

商品種別の5択と、6つの意味評価軸のyes/noを、独立した7つのChoice質問として送ります。各軸はyesの確率を使用し、provider confidenceを別に保存します。[Noulには独立したconfidenceがありません](https://docs.typesafe.ai/confidence)。そのため今回はChoiceを採用しました。Noul/Score応答の検証・変換もテストしますが既定質問はChoiceです。

6軸は research_value / problem_specificity / comparison_value / longtail_potential / content_value / commodity_risk。longtailは検索需要の実測や予測ではなく、商品に根拠のある具体的な検討テーマの余地です。

Goで `clamp(0.22*research + 0.18*problem + 0.20*comparison + 0.18*longtail + 0.22*content - 0.25*commodity + role補正, 0, 1)` を計算します。補正は本体+0.05、交換品0、付属品0、セット-0.01、不明-0.03。confidenceを機会スコアに混ぜません。重みは `-weights path.json`（`internal/jev/score.go` のScoreConfig形式）で変更できます。

### 認証・実行

ローカル `.env` の `JEV_API_KEY` に設定済みのキーを置いてください。`.env.example` の `JEV_BASE_URL` / `JEV_MODEL` は既定値です。秘密情報をGitに登録しないでください。`-env-file` で既存ファイルを選べます。キーがない場合はAPIを呼ばず、計画を表示して終了します。

```powershell
# APIを呼ばず計画・質問・sample state・入力サイズを確認
 go run ./cmd/evaluate -dry-run
# Stage A: 10件
 go run ./cmd/evaluate -limit 10
# Stage B: 追加100件（累計100件にする場合は -limit 90）
 go run ./cmd/evaluate -limit 100
# Stage C: 残りすべて
 go run ./cmd/evaluate
# 保存済みの上位20件だけ表示。API呼び出しなし
 go run ./cmd/evaluate -top 20 -min-score 0
# 保存済み評価からCSV出力。API呼び出しなし
 go run ./cmd/evaluate -export-review
# 同一設定・同一version内の明示的再評価
 go run ./cmd/evaluate -limit 10 -force
# 質問やモデルや重み等を変更した実験は新version
 go run ./cmd/evaluate -version v2 -dry-run
```

`go build -o evaluate.exe ./cmd/evaluate` でWindows実行ファイルを作れます。配布バイナリは同じオプションです。DB移行前のバックアップを保持し、schema 2では新しいcollect/inspectも使用してください。旧Day 1バイナリは新schemaを開けません。

### 再評価防止・状態・耐障害性

主キーは `(source, source_id, evaluation_version)`。既存評価はskip、`-force` は同じ設定で更新します。モデル、質問、重み、一次フィルタ、state前処理、対象商品とstate hashをversionのfingerprintとして固定し、設定変更には新versionを要求します。単に同じversionに `-force` を付けても設定変更は許しません。dry-runは評価version・結果・runを登録しません（DB schema移行は行います）。

Day 1のstate出力v1は維持し、評価専用state v2を別に作りました。商品名・説明をuntrusted dataとして質問から分離し、説明は既定3000 Unicode文字、名前300文字、shop120文字まで。隣接する同一文をまとめ、URLやraw JSONを送りません。DBの元の商品は変更しません。数値比較はGo、価格はAPI値を使い、商品名の割引価格を確定価格として扱いません。

説明中の採点要求や宣伝を無視する質問を設定しています。ただし[公式のモデル制約](https://docs.typesafe.ai/model-jaggedness/jev-1.13)も踏まえ、プロンプト注入への完全な耐性や日本語の判定精度は保証せず、実際の人間レビューで確認します。

HTTP既定timeout20秒、間隔250ms、追加retry3回、商品あたり最大2分。429/529、一時的5xx、timeout等に待機・backoff・Retry-Afterを使用。認証・schema・model等の致命的エラーは停止し、その他は商品単位で記録して続行します。成功結果は逐次保存します。壊れた成功応答は課金重複を避けて自動再送せず、再実行で未評価を再試行します。

### usage・ランキング・人間レビュー

dry-runの入力サイズはUTF-8/JSONのバイト数でありtoken数ではありません。実応答のinput/output tokensだけを記録し、欠損値はunknown。retryでusage不明の試行を別途数えます。既定モデルの推定入力料金は$0.042/百万token、output無料（上記モデル仕様）。`-input-usd-per-million` で変更できます。別モデルの既定料金はunknown。表示額は既知usage分の推定額で、実請求額ではありません。

保存済み評価のcoverage、種別分布、scoreのmin/mean/median/p90/p95/p99/max、軸とconfidenceのmean/median、上位10/5/3/1%を集計します。割合は切り上げ、同点はsource/source_id順に固定件数を選びます。395件すべて評価済みなら40/20/12/4件です。

40件以上評価済みでCSVを作成できます。Top20と、それを除いた評価済み母集団からRandom20を重複なしに選びます。途中段階のCSVはその評価済み部分集合の比較に留まります。`-review-seed 42` で再現可能です。

- `data/day2-review.csv`: score/role/confidenceを含む通常CSV、human_good_candidate / human_score / human_notesは空欄。
- `data/day2-review-blind.csv`: group/score/role/confidenceを隠し、順序を混ぜた商品情報とreview ID。まずこちらを人間が評価してください。
- `data/day2-review-key.csv`: IDとgroupの対応表。人間評価終了まで開かずローカルで保持してください。

UTF-8 BOMとCSV quotingに対応し、表計算ソフトの数式解釈を避けます。人間レビュー後にTop群とRandom群を比較するまで、濃縮性能が良いとは結論しません。

### Day 2の検証状況

`go test ./...` / `go vet ./...` が成功。HTTP mock、応答・confidence検証、score、schema 1→2、upsert/version/force、部分失敗からの再開、CSVの40件・重複なし・blind化を検証しました。

実DBは1020商品・eligible395件。移行前後のproducts全列を双方向比較し差分0、評価テーブルは0件。dry-runで395件、平均state3837.54 bytes、最大9101 bytes、全request合計4556537 bytesを確認しました。Jev認証情報がローカルに存在しないため、ご依頼の例外に従って実API評価は未実施です。実ランキング、誤判定、confidence傾向、usage、実レビューCSVはまだありません。
# Day 3: recall gateとLLM reranker

Day 3は追加調査候補の選別実験です。記事生成・公開・商品検索・Jev再評価は行いません。
Jev v1のquestions、model、6軸、role、state、重み、一次フィルタは固定し、保存済みv1評価だけを使います。

## Gateを先に固定

Day2.6の120件だけで、AI平均4以上のpositive recallが95%以上となるthresholdの中から、最もreject件数が多いものを選びます。同じreject件数なら低いthresholdを採用します。

```powershell
go run ./cmd/gate -db data/money.db -calibration data/day2-6-analysis.csv -out data/day3
```

出力は `gate-v1.json`、全thresholdの `gate-thresholds.csv`、取りこぼしたpositiveの `gate-false-negatives.csv`。
既存ファイルは上書きしません。実験済みのGateを再利用してください。LLM結果からthresholdを変更しません。
rawはv1と同じ式のclamp前の値です。9桁CSVの丸め誤差を境界比較に限り5e-10まで許容します。
校正データは層化・role調整済みの標本なので、校正上95%以上でも母集団Recallの保証にはなりません。

## モデルと評価

固定モデルは `claude-sonnet-5-5`、versionは `llm-reranker-claude-v1`。
公式仕様上、日付なしIDも固定snapshotです。OpenAI版は実評価0件のまま、ユーザー指定によりClaude用の新versionへ切り替えました。
[公式モデル仕様・料金](https://platform.claude.com/docs/en/models/sonnet-5-5/overview)と
[Structured Outputs仕様](https://platform.claude.com/docs/en/build-with-claude/structured-outputs)を2026-10-04に確認しています。
Messages APIのJSON Schemaで6軸とoverallを取得し、0〜100整数制約はGo側でも検証します。
adaptive thinking、effort=medium、max_tokens=4096。Web検索や他のツールは有効にしません。
商品データは保存済みJev stateから転記し、追加のdescription加工は行いません。
Jev product_roleだけを付加し、raw・Opportunity Score・AI judge score・過去group・source_idはLLMに渡しません。

## Dry-run

```powershell
go run ./cmd/rerank -dry-run -gate data/day3/gate-v1.json -limit 0
```

API呼び出しとrerankerテーブルへの書き込みはありません。商品数、既評価数、pending、model、version、入力例、rubric、UTF-8 request byte数を表示します。byte数をtoken数や費用とは扱いません。
APIキーは `.env` の `ANTHROPIC_API_KEY`。既存のローカル設定は `-env-file <path>` でも指定できます。
キー未設定で通常実行した場合も、dry-runを表示して実API評価前に停止します。

## 段階的実行と再開

```powershell
# Stage A: 10件。JSON構造・6軸範囲・model一致・DB保存を確認
go run ./cmd/rerank -limit 10
# Stage B: pendingを追加40件、累計約50件。分布と理由を確認
go run ./cmd/rerank -limit 40
# Stage C: 残り全件
go run ./cmd/rerank -limit 0
```

実API評価前の応答構造確認はテスト用fixtureによるもので、実サービスとの接続確認ではありません。
失敗商品は保存せず、成功ごとに即時保存します。同じversionは通常再評価しません。
versionはmodel・rubric・Schema・Gate・cohort/input hash・tie-break・reasoning設定を拘束し、変更を検出したら停止します。
HTTP timeout 60秒、商品ごとのbudget 3分、最大2 retry、指数backoff、429/5xx対応。認証エラーは即時停止し、provider error bodyはログに出しません。

実行時だけ追加するテーブルは `reranker_versions`、`llm_product_evaluations`、`llm_rerank_runs`。
既存のproducts/product_evaluationsを更新しません。Day 2のschema versionは保持し、Day 3追加テーブルは冪等に作成します。
各軸・理由・request/raw response・入力hash・評価日時・Gate/version/modelを保存します。
usageが返った呼び出しのinput/output/cached tokensと推定費用をrun単位で記録します。
不明なusageや課金され得る失敗retryも別集計し、推定費用の既知小計を実請求額とは扱いません。
料金は `-prices config/llm-prices.example.json` で切り替え可能です。料金変更は採点に影響しません。

## 全件完了後のTop20とblind review

```powershell
go run ./cmd/rerank -export-review data/day3
```

全Gate通過商品が評価済みの場合だけ出力します。
overall、investigation、comparison、independent value、buyer problem、wrong choice、audienceの降順で固定sortし、全軸同点の最終tieのみsource/IDで安定化します。
Random20は同じGate通過集合からTop20を除いてseed=20261004で抽出。40件のID・順序もランダム化します。
blind/key、Top20/Random20一覧、事前判定基準を含むplanを保存し、既存sampleは上書きしません。
URLのquery/fragmentを除去し、blind CSVには商品情報だけを含めます。

3つの新しい独立文脈のAI judgeには `day3-review-blind.csv` だけを与えてください。
Jev/LLM結果、key、過去履歴、他judgeの回答は共有しません。評価は商品の優劣でなく購入判断コンテンツの価値（1〜5）。
`day3-judge-1.json`〜`day3-judge-3.json` に40件ずつの `review_id`、整数 `score`、短い日本語 `notes` を保存します。

```powershell
go run ./cmd/analyze-day3
```

120採点の完了を検証した後だけkeyを読み、40件一対一・重複なし・20+20を検証します。
不一致なら停止。元CSVは変更せず、`day3-ai-judge.csv`、`day3-analysis.csv`、`day3-comparison.json` を保存します。
平均/中央値/SD、4以上・4.5以上率、Hedges g、Cliff、bootstrap 10,000回95% CI、judge別差、overall/6軸のSpearmanを出力します。
Strong GOは差>=0.5、全judge正、g>=0.5、CI下限>0。Weak GOは差>=0.2で全judge正。差<0.2はNO-GO、その他はinconclusive。
独立AI judgeとの一致の検証であり、人間ground truth・収益性の証明ではありません。Day 4へ自動で進みません。

```powershell
go test ./...
go vet ./...
```

DB・環境ファイル・実API raw data・review/AI judge結果はGitに含めず、`data/` またはリポジトリ外の出力先に保存してください。

Claude料金は100万tokenあたり通常入力USD2、出力USD10、cache read USD0.20、5分cache write USD2.50。明示的cache_controlは使いません。InputTokensは通常入力+cache read+cache writeの合計です。

## Day 3.5: product family / diversification

保存済み329件を再利用し、API呼出しを行わず商品レベル順位とテーマ選択を分離します。
元Day3 DBのコピーを指定してください。以下はリポジトリから実行する例です。

```powershell
Copy-Item ../day3-money.db ../day3-5-money.db
go run ./cmd/diversify -db ../day3-5-money.db -out ../../outputs/day3-5-results/data -family-cap 2 -top 20 -judge ../../outputs/day3-live-results/data/day3-analysis.csv -holdout ../../outputs/day2-live-results/data/day2-6-key.csv
```

同時にnone/cap1/cap2/cap3/penalty/MMRを出力します。主方式cap2は固定です。
NFKC・広告除去・型番抽出でLevel Aを識別し、brand+seriesでLevel Bを作成。
Aは最大1代表、Bは最大2代表。互換フィルターは適合先のbrandと物理メーカーを混同しません。
型番・ブランド不明は保守的に分け、brand既知でモデル不明のfallbackはJaccard>=0.90のcomplete link。
Level Cは商品名に明示された用途キーワードによる暫定テーマです。

元のoverall/6軸は変更せず、別テーブルにversion付きassignmentとrankを保存。
同じversionの設定・入力・algorithm fingerprintが変わった場合は上書きを拒否します。
`-dry-run` はCSV/JSON/reportを生成し、新テーブルを保存しません。
`diversified_rank=0` はcapや同一商品代表選択により除外された商品です。
生成物はproduct-families/diversified-ranking/top20/day4-candidates CSV、statistics JSON、
Day3-5-diversification-report.md、evaluation-manifests/day3-claude.json。
Manifestは保存済みrequest/responseとrunの料金snapshotから生成し、secretを保存しません。

Day3.5のscore lossは元Claudeの0〜100点単位。指定された0.20/0.30点基準を適用し、
参考の0〜1換算値をGO判定に用いません。既存judge被覆部分だけで新Top20全体の品質は証明できません。
Day4調査・コンテンツ生成・公開には進みません。

```powershell
go test ./...
go vet ./...
```

## Day 3.6: diversification blind quality validation

Day3 original Top20と固定済みDay3.5 cap2 Top20のunionを、同一rubricの新しい独立文脈3 judgeで採点します。
アプリのrubric/model/weights/Gate/rank/family/cap/5テーマは変更しません。旧judgeは再利用しません。

```powershell
go run ./cmd/validate-diversification -mode prepare -data ../../outputs/day3-6-results/data
```

`prepare` は新しい空ディレクトリだけに書き、既存成果物を上書きしません。DBはmode=roで読むだけです。
固定seed=20261004で匿名ID/順序を作り、全文の商品説明・価格・レビュー・shop/categoryだけをblind CSVへ保存。
keyは別ファイル。各judgeへblind CSVと`validation.Rubric`の共通基準だけを渡し、
`day3-6-judge-1/2/3.json`にreview_id/1〜5整数score/日本語notesのarrayを保存します。

```powershell
go run ./cmd/validate-diversification -mode analyze -data ../../outputs/day3-6-results/data
```

全judgeのID/件数/score/notesを検証した後だけkeyを読み、20+20、一対一join、固定diversity/Claude平均を照合。
Overlapを共有するunion listing-unit bootstrap（10,000回、random membership denominators）と、
disjoint changed-items群内bootstrapを両方出力します。SDはsample、Spearmanはtie平均rank。
入替pairはremoved原順位/added多様化順位による説明用対応であり、因果的な1:1ではありません。

判断の曖昧な部分は採点前protocolで固定: >=4.5率の大幅悪化は10ポイント超、新規平均<=2.5はcatastrophic、
平均<3の新規商品3件以上は多数、changed差<=−0.50かつCI上限<0は明確な低下。
ユーザー指定の「3 judge全て悪化」はstrictly negativeとしてNO-GOを優先します。
Day3.5 NO-GOを変更せず、Day3.6 GOはAI judge品質維持に限定。Day4へ自動で進みません。

再計算はblind/key/protocol/3judge JSONだけを別の新規dataディレクトリへコピーし、同じanalyzeコマンドで行います。
入力CSV/DB、judge結果、key、実験出力・環境ファイルはGitに含めません。

## Day 4: external market evidence export

固定5テーマの外部市場調査を、手動確認済みの構造化inputから出力します。
評価モデル・Gate・weights・rank・familyは変更せず、API再評価は行いません。
このCLIはWeb検索・価格収集を自動実行せず、観測済み根拠と固定rubricの再計算だけを行います。

```powershell
go run ./cmd/validate-day4 -data ../../outputs/day4-market-validation/data -out ../../outputs/day4-reexport/data -commit 5facdb2
go test ./...
go vet ./...
```

入力は `day4-research-input.json`、`day4-search-results.json`、検索前に凍結した
`day4-query-set.csv`、hash付き `day4-protocol.json`。新規の出力先を指定してください。
既存出力は上書きせず、query hash改変・テーマ変更・score/risk範囲外・欠落/重複query・出典参照不一致は停止します。
同じinputならCSV/JSONが再現します。調査CSV/価格/全文外部コンテンツ/API keyはGitに含めません。

Market Scoreは7軸0〜5、risk3軸は別表示。Strong GO/GO/HOLD/NO-GOの事前gateを保持します。
検索ツール返却順とGoogle organic rankは区別し、domain分類はheuristicと明記。
検索量・SKU別報酬率は未確認ならunknown、注文UIと在庫確約は区別します。
レポートは調査主体の別artifactで、CLIは記事生成・公開・次工程を実行しません。

## Day 5: evidence-backed content validation

固定3テーマのローカルprototypeと匿名化した競合要約を比較します。公開や評価API呼び出しは行いません。

```powershell
go run ./cmd/validate-content -mode prepare -data ../../outputs/day5-content-validation/data
# blind CSVだけを独立3 judgeに渡し、別reviewerのclaim auditを保存した後:
go run ./cmd/validate-content -mode analyze -data ../../outputs/day5-content-validation/data
go test ./...
go vet ./...
git diff --check
```

prepareにはday5-input.jsonと隣のcontentディレクトリの3記事が必要です。claim/source/questionの参照と重複を検証し、計算・固定seedの匿名ID・CSV・SHA256 manifestを生成します。
analyzeはday5-judge-1/2/3.jsonとday5-audit.jsonを必要とし、全12候補の整数score/ID/notes検証後にだけprivate keyを読みます。全29 claimのaudit、input/blind/prototype hashも照合します。
既存出力への上書きは拒否します。再実行は新規ディレクトリに入力とcontentをコピーしてください。judge再集計にはprepared/key/blind/manifestもコピーし、集計出力を含めないでください。

主差はprototype平均−競合3ページ中の最大平均。judge別の非劣性は各judge自身の最良競合に対して判定します。Strong GO/GOには重大監査違反0が必要です。
比較は構造化要約上の独立AI評価であり、人間ground truth、全文ページの優越性、SEOや収益を証明しません。価格の旧snapshotと現在確認のunknownを区別し、調査データ・記事・API keyはGitに含めません。

## Day 6: local canary artifact

```powershell
go run ./cmd/build-canary -out ../../outputs/day6-canary-publication/site/new-preview.html
# 正規公開先が確定した場合だけ -canonical に実際のHTTPS URLを指定する
```

入力はGit管理した `content/published/compact-air-purifier-placement.md`。ディレクトリ名に関係なく現段階では未公開です。
小さな限定Markdown rendererが見出し・表・HTTPS資料リンクを生成します。広告表示、H1、禁止表現、canonical、affiliate sponsoredを検査し、既存出力の上書きを拒否します。一般的なCMS/Markdown engineではありません。
公開先が見つからない場合はPUBLICATION_TARGET_MISSINGで停止します。canonicalを省略した出力はローカル準備稿で、live SEOチェック成功を意味しません。
新しいホスティング・analytics・Search Console設定を採用せず、計測CSV/manifestの未取得値はunknownとして扱います。CLIはdeploy、index申請、tracking、2本目公開を実行しません。

## ChoiceLen production canary (Day 6.1)

The static site at https://choicelen.page/ contains one frozen Day 6 article:
https://choicelen.page/articles/compact-air-purifier-placement/ . Claim-bearing Markdown is preserved; the build changes publication metadata and navigation/CSS only. Workers Static Assets serves the committed `site/`, with custom-domain routing in `wrangler.jsonc`. Wrangler is pinned to 4.147.0 in the workflow.

To reproduce the publication build, choose a **new** output directory (the CLI refuses an existing one):

```sh
go run ./cmd/build-choiceLen -date 2026-10-06 -out outputs/choicelen-rebuild
go run ./cmd/preview-choiceLen -dir site
```

Preview: http://127.0.0.1:8787/ . Production verification uses normal certificate validation and writes new D+0 evidence without overwriting it:

```sh
go run ./cmd/verify-choiceLen -out outputs/choicelen-new-check
go test ./...
go vet ./...
git diff --check
```

`.github/workflows/choiceLen.yml` uses repository Actions secrets `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID`. Never place these values in config, HTML, commits or issue/PR text. A deploy occurs only with `workflow_dispatch`, or an explicit `[deploy-choicelen]` commit marker on `day6-1-production-publication`. Until the workflow is on the default branch, marked branch pushes are the explicit trigger. Ordinary pushes do not deploy. No existing PR stack is merged by the deployment. `workers.dev` is disabled; no extra subdomain is required.

Public publication metadata and the anchored measurement template are versioned in `data/day6-publication-manifest.json` and `data/day6-measurement-template.csv`. Original blocked Day 6 artifacts are retained separately. Current status: **GO_WITH_AFFILIATE_PENDING**. No real affiliate ID was configured, so there are zero affiliate links. Advertising disclosure remains visible. Search Console, analytics, indexing and conversion metrics are unavailable/unknown; no third-party tracking or measurement automation was enabled.
