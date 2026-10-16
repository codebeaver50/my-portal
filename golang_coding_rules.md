# Go Coding Rules（Go API / `server/`）

このドキュメントには、一般的なスタイルガイドに加えて **このプロジェクトで決めたこと** と **避けたいアンチパターン** を書く。
ディレクトリ構成・レイヤーの役割・設計判断は `ARCHITECTURE.md` を正とし、ここではコードの書き方を扱う。

## 前提とするガイド

ここに書いていないことは、以下に従う。

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)（Go Code Review Comments と食い違う場合は Go Code Review Comments を優先する）

機械的にチェックできるものは [golangci-lint](https://golangci-lint.run/)（設定は `server/.golangci.yml`）に任せる。CI（`.github/workflows/ci.yml`）でも `go build` / `go vet` / `go test` とあわせて golangci-lint を実行し、違反があるとCIが失敗する。

```bash
cd server
golangci-lint run ./...   # チェック
golangci-lint fmt ./...   # gofmt / goimports で整形
```

### Lint で強制しているルール

`server/.golangci.yml` で、標準のlinter（`errcheck` / `govet` / `ineffassign` / `staticcheck` / `unused`）に加えて、このドキュメントのうち以下のルールをチェックしている。ルールを追加・変更した場合は、このドキュメントと `server/.golangci.yml` の両方を揃える。

| ルール | linter |
|---|---|
| `gofmt` で整形し、import を3つのグループに分ける | `gofmt` / `goimports`（formatter） |
| エラーの判別は `errors.Is` / `errors.As` で行う | `errorlint` |
| 書式の指定がない固定のメッセージは `errors.New` を使う | `perfsprint`（`errorf` のみ） |
| 外部への HTTP リクエストに context を渡す | `noctx`（テストを除く） |
| `resp.Body` を閉じる | `bodyclose` |
| `panic` / `http.DefaultClient` を使わない。`internal/` で `os.Getenv` しない | `forbidigo` |
| feature から別の feature を import しない | `depguard`（featureを追加したらルールを追記する） |

`//nolint` で止める場合は、`//nolint:<linter> // 理由` のように対象のlinterと理由を書く。

## パッケージ・ファイル

- パッケージ名は小文字の1単語にする。feature-slug のハイフンは詰める（`book-database` → `bookdatabase`）
- 別のパッケージと同じ名前になる import には、何のパッケージか分かる別名を付ける
  ```go
  acmiddleware "github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
  simpleledgerrouter "github.com/codebeaver50/my-portal/server/internal/features/simpleledger/router"
  ```
- import は「標準ライブラリ」「外部モジュール」「このモジュール（`github.com/codebeaver50/my-portal/server/...`）」の3つのグループに空行で分ける
- ファイル名は snake_case にし、レイヤー名を接尾辞にする（`transaction_controller.go` / `transaction_service.go` / `transaction_repository.go`）。`internal/models/` はテーブル名と同じファイル名にする（`simple_ledger_transactions.go`）
- feature から別の feature のパッケージを import しない。複数の feature で使う処理は `internal/db` や `internal/accesscontrol` のような feature 非依存の場所に置く

## レイヤーごとの書き方

レイヤーの役割（controller → service → repository、dto、router）は `ARCHITECTURE.md` を参照。ここでは書き方の決まりを扱う。

### 型の公開のしかた

| レイヤー | 公開する型 | 理由 |
|---|---|---|
| controller | `interface`（`TransactionController`）と、それを返す `NewXxx`。実装の struct は非公開 | router から使うメソッドだけを見せる |
| service | `interface`（`TransactionService`）と `NewXxx`。実装の struct は非公開 | controller のテストで service を差し替えられるようにする |
| repository | 具体的な struct（`*TransactionRepository`）と `NewXxx` | service のテストはインメモリの SQLite で本物の repository を使うため、interface を作らない |
| 外部API連携 | 使う側（service）に interface を定義する（`bookdatabase/service/clients.go` の `NDLClient`） | テストでネットワークに出ないフェイクに差し替えるため |

- 使う側が必要としない interface を先回りして作らない（Go Code Review Comments の「Interfaces」）
- 依存はコンストラクタ（`NewXxx`）の引数で受け取り、パッケージ変数・`init()` で組み立てない。組み立ては feature の `router` で行う

### controller

- ハンドラは `gin.HandlerFunc` を返すメソッドにする（`func (ctrl *transactionController) Create() gin.HandlerFunc`）
- リクエストの読み取り（パスパラメータ・クエリ・`ShouldBindJSON`）と、service の結果・エラーを HTTP のステータスとレスポンスに変換することだけを行う。業務上の判断（存在チェック・重複チェック等）は service に置く
- service には `c.Request.Context()` を渡す
- エラーとステータスの対応は「エラー処理」の節を参照

### service

- 第1引数に `ctx context.Context` を取る
- 複数の repository にまたがる書き込みは、service でトランザクションを張り、repository の `WithTx(tx)` で同じトランザクションを使わせる
  ```go
  err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
      transactionRepo := s.transactionRepo.WithTx(tx)
      journalEntryRepo := s.journalEntryRepo.WithTx(tx)
      // ...
  })
  ```
- `internal/models` のエンティティと `dto` の変換は service（または dto の `ToXxxResponse` 関数）で行う

### repository

- GORM のクエリを1つ（または取得＋件数のように対になるもの）発行するだけにする。条件による分岐・find-or-create のような判断は service に置く
  ```go
  // ❌ repository が「なければ作る」を判断している
  func (r *PublisherRepository) Resolve(ctx context.Context, name string) (uint, error)
  // ✅ repository は FindByName / Create だけを持ち、分岐は service で行う
  ```
- `r.db.WithContext(ctx)` を必ず通す（リクエストのキャンセル・タイムアウトをクエリに伝えるため）
- 1件取得で行が見つからない場合は `(nil, nil)` を返し、service が `ErrXxxNotFound` に変換する。`gorm.ErrRecordNotFound` を repository の外に出さない
  ```go
  err := r.db.WithContext(ctx).First(&transaction, id).Error
  if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, nil
  }
  ```
- トランザクション用に `WithTx(tx *gorm.DB) *XxxRepository` を持たせる

### dto / models

- API のレスポンスは必ず `dto` の型で返す。`internal/models` のエンティティを `c.JSON` に直接渡さない（テーブルの列を足しただけで API の形が変わってしまうため）
- JSON のキーは camelCase にする（`json:"journalEntries"`）。フロントエンドの `_lib/types.ts` と揃える
- `internal/models` の struct を変更したら、対応するマイグレーションSQLを必ず追加する（`ARCHITECTURE.md` の「DBマイグレーション」）。`TableName()` でテーブル名を明示する
- 取りうる値が決まっている文字列は、名前付きの型と定数にする（TypeScript の文字列リテラルのユニオン型に相当する）
  ```go
  type EntryType string

  const (
      DebitEntry  EntryType = "debit"
      CreditEntry EntryType = "credit"
  )
  ```

## エラー処理

### エラーの作り方・返し方

- エラーは呼び出し元に返す。ログに出して握りつぶさない・ログに出してから返す（二重にログが出る）こともしない。ログに出すのは controller（とミドルウェア・`cmd/`）だけにする
- 呼び出し元が判別する必要があるエラーは、service にパッケージ変数（センチネルエラー）として定義する
  ```go
  // ErrTransactionNotFound は指定のIDの取引が存在しない場合に返される。
  var ErrTransactionNotFound = errors.New("transaction not found")
  ```
- ユーザーに見せる入力エラーは、feature の service の `ValidationError`（`Message`、項目ごとのエラーがあれば `FieldErrors`）で返す。`Message` は画面にそのまま出せる日本語にする
- 下位のエラーに情報を足して返すときは `fmt.Errorf("...: %w", err)` で包む。書式の指定がない固定のメッセージは `errors.New` を使う
- エラーの判別は `errors.Is` / `errors.As` を使い、`==` や型アサーションで比べない（`%w` で包まれたエラーを判別できなくなるため）
  ```go
  // ✅
  if errors.Is(err, gorm.ErrRecordNotFound) { /* ... */ }
  // ❌
  if err == gorm.ErrRecordNotFound { /* ... */ }
  ```
- `panic` は使わない。プログラムを止めてよいのは `cmd/` の `main` での起動時の設定エラー（`log.Fatal`）だけ

### controller でのレスポンスへの変換

| エラー | ステータス | レスポンスの本文 |
|---|---|---|
| パスパラメータ・クエリが不正 | `400` | `{"error": "invalid xxx id"}` 等 |
| `ShouldBindJSON` の失敗 | `400` | `{"error": "入力内容を確認してください"}` 等の日本語の固定メッセージ |
| `ValidationError` | `400` | `{"error": Message, "fieldErrors": {...}}`（`formbuilder/controller/errors.go` の `writeValidationError`） |
| `ErrXxxNotFound` | `404` | なし |
| 重複等の競合（`ErrPostSlugTaken`） | `409` | `{"error": "..."}` |
| 上記以外（DBエラー等の想定外のエラー） | `500` | **なし**。`log.Printf("<feature-slug>: <操作> failed: %v", err)` でログに出す |

- **想定外のエラーの `err.Error()` をレスポンスに入れない**。SQL・テーブル名・内部の構成がクライアントに漏れるうえ、サーバー側の障害がクライアントの入力ミス（`400`）に見えてしまうため
  ```go
  // ❌ DBエラーまで400でクライアントに返してしまう
  if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
  }
  // ✅ 判別できるエラーを先に処理し、残りは500にしてログに出す
  if writeValidationError(c, err) {
      return
  }
  if err != nil {
      log.Printf("form-builder: create form failed: %v", err)
      c.Status(http.StatusInternalServerError)
      return
  }
  ```
- `ShouldBindJSON` のエラー（validator のメッセージ）もそのまま返さない。`Key: 'TransactionRequest.Date' Error:Field validation for 'Date' failed on the 'required' tag` のような英語の内部向けのメッセージが画面に出てしまうため。項目ごとの分かりやすいエラーは、フロントエンドの zod による事前チェックと service の `ValidationError` で出す
- 認証・権限のエラー（`401` / `403`）は `accesscontrol/middleware` の `Require` が返す。feature の controller で権限をチェックしない

## 入力の検証

- **検証の正は Go 側**。フロントエンドの zod スキーマは事前チェックで、文字数上限等の制約値を変える場合は両方を揃える（`fe_coding_rules.md` の「入力の検証（zod）」）
- 型・必須・文字数のような形の検証は dto の `binding` タグ（`binding:"required,max=255"`）で行い、DBの状態に依存する検証や業務ルール（借方合計=貸方合計等）は service で行う
- 定数（上限値等）は意味の分かる名前の定数にする（`maxOptions` / `maxTextLength`）

## context・外部API

- DB・外部API・ほかの service を呼ぶ関数は、第1引数に `ctx context.Context` を取り、下へそのまま渡す。`context.Background()` を途中で作り直さない（`cmd/` とテストを除く）
- 外部への HTTP リクエストは `http.NewRequestWithContext` で作り、`Timeout` を設定した `http.Client` で送る。`http.Get` や `http.DefaultClient` はタイムアウトがなく、相手が応答しないとリクエスト（cron の同期処理等）が止まり続けるため使わない
- `resp.Body` は `defer resp.Body.Close()` で必ず閉じ、ステータスコードを確認してから読む

## 設定・環境変数

- 環境変数は `cmd/` の `main` で読み、必要な値を引数で下へ渡す。`internal/` の中で `os.Getenv` しない（テストで差し替えにくく、どの設定に依存しているかが分かりにくくなるため）
  - 例外: `internal/contact` は問い合わせメールの SMTP 設定を読んでいる（現状のまま）
- 必須の環境変数がない場合は、起動時に `log.Fatal` で止める。既定値がある場合（`PORT` 等）は `main` で補う

## ログ

- `log` パッケージを使い、`<feature-slug>: <操作> failed: %v` の形式で、どこで何が失敗したかが分かるようにする
- パスワード・セッショントークン・Cookie の値をログに出さない

## テスト

- 標準の `testing` パッケージを使う（testify 等のアサーションライブラリは使わない）。失敗時のメッセージは `got` と `want` が分かるようにする
  ```go
  t.Errorf("err = %v, want ErrFormNotFound", err)
  ```
- ケースごとに `t.Run` でサブテストに分け、名前には何を確かめるのかを書く（`"rejects unknown field keys"`）。入力と期待値の組が多い場合はテーブル駆動テストにする
- service・repository のテストは、テスト用のインメモリ SQLite（feature ごとの `testdb_test.go` の `newTestDB`）を使う。テスト用DBに限り `AutoMigrate` を使ってよい（本番のスキーマは golang-migrate で管理する）
- controller のテストは `httptest` で HTTP のステータス・レスポンスを確かめる
- 外部APIはフェイク（interface を満たすテスト用の型）に差し替え、テストからネットワークに出ない
- テストの準備用の関数は `t.Helper()` を呼び、後始末は `t.Cleanup` で登録する

## 命名

| 対象 | 規則 | 例 |
|---|---|---|
| パッケージ | 小文字の1単語 | `bookdatabase`, `controller` |
| 公開する型・関数 | PascalCase | `TransactionService`, `NewTransactionService` |
| 非公開の型・関数 | camelCase | `transactionService`, `validateEntries` |
| コンストラクタ | `New` + 型名 | `NewFormController` |
| センチネルエラー | `Err` + 内容 | `ErrFormNotFound` |
| 頭字語 | すべて大文字かすべて小文字 | `ID`, `URL`, `userID`（`Id` / `Url` にしない） |
| レシーバ | 型を表す短い名前で統一する | controller は `ctrl`、service は `s`、repository は `r` |
| ルート登録の関数 | `Setup` + feature名 + `Routes` | `SetupSimpleLedgerRoutes` |

## コメント

- コメントは日本語で書き、**なぜそうしているのか**（理由・制約・避けている問題）を書く。コードを読めば分かる「何をしているか」は書かない
- 公開する型・関数・定数には、その名前で始まるドキュメントコメントを付ける（`// TransactionService は取引の作成・...を実装する。`）
- パッケージには `// Package xxx は ...` のコメントを、パッケージ内のどれか1ファイルに付ける
- HTTP ハンドラのドキュメントコメントには、処理するメソッドとパスを書く（`// Create は POST /api/simple-ledger/transactions を処理する。`）
