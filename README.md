# jev-money-engine

個人向けの商品候補収集システム。Day 1では楽天市場の商品を取得し、共通モデルに正規化してSQLiteに保存、Goの一次フィルタで候補を選別します。翌日にJevへ渡すJSON stateを生成できます。Jev/LLM API、Web UI、記事・SNS生成は実装していません。

## 必要な環境とセットアップ

- Go 1.26以上、ネットワーク接続
- 楽天Web ServiceのApplication IDとAccess Key
- SQLiteはpure Goドライバ `modernc.org/sqlite` を使用。Python、Cコンパイラ、SQLiteの別途インストールは不要

```sh
git clone https://github.com/AkiraShimizu0/rakuten-product-search.git jev-money-engine
cd jev-money-engine
go mod download
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

DBはプロジェクトルートからの相対パス `data/money.db`。`-db` で変更可能。初回に自動作成し、`PRAGMA user_version=1` の簡易スキーマ管理を使用します。`data/` 全体はGitから除外されます。

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

## 構成

```text
cmd/collect/       楽天収集CLI
cmd/inspect/       DB集計・サンプル・state出力CLI
internal/product/ データソース非依存Product
internal/rakuten/ HTTP clientと楽天専用正規化
internal/store/   SQLite保存
internal/filter/  決定的な一次フィルタ
internal/jev/     Jev state変換のみ
internal/report/  集計とサンプリング
internal/config/  ローカル環境設定
internal/cli/     共通フラグ
data/             実データ・DB（Git対象外）
```

将来のcollectorは専用のレスポンス型と正規化を追加し、共通Productに変換してstore/filter/reportを再利用できます。Day 1では他ECサイトや汎用collector frameworkを追加しません。
