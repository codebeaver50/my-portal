# Architecture

## 技術スタック

- フロントエンド: Next.js 16 (App Router) / TypeScript / React 19 / Tailwind CSS v4
- API: Go / Gin（HTTPルーティング） / GORM（DBアクセス）
- DB: MySQL（Dockerコンテナ）
- デプロイ: AWS Lightsail Instance（VPS）1台の上で、Next.js・Go API・MySQL・リバースプロキシ（Caddy）をDocker Composeでまとめて動かす。MySQLはコンテナの外に公開しない

## ディレクトリ構造（案）

```
app/
  page.tsx                    # ポータルトップ（feature一覧）
  (features)/                 # route group。URLには出現しない
    <feature-slug>/           # 各ポートフォリオアプリケーション（URL: /<feature-slug>）
      page.tsx
      _components/            # そのfeature専用のコンポーネント（非ルーティング）
      _lib/                   # そのfeature専用のロジック（非ルーティング）
      ...                     # ページを複数持つfeatureはさらにネストしたルートを配置
components/
  ui/                         # プロジェクト共通のUIコンポーネントライブラリ（Button, Card等）
server/
  go.mod
  cmd/api/main.go             # エントリポイント（Goの標準的なcmd/レイアウト）
  cmd/migrate/main.go         # DBマイグレーション適用コマンド
  migrations/                 # SQLマイグレーションファイル（golang-migrate形式）
  internal/
    db/                       # feature非依存の共通処理（GORMのDB接続等）
    models/                   # DBエンティティ（GORM struct）。feature間で共有し、テーブルと1対1対応させる
      book.go
      publisher.go
      ...
    features/
      <feature-slug>/         # featureごとのドメインロジック（Next.js側のfeature-slugと対応）
        controller/           # HTTPハンドラ（Ginのgin.HandlerFuncを返す）
        service/               # ビジネスロジック。repositoryを呼び出し、dtoを組み立てる
        repository/             # DBアクセス（GORMクエリ）。internal/models のエンティティを読み書きする
        dto/                     # APIリクエスト/レスポンス専用の型（modelsのエンティティとは別）
        router/                  # repository→service→controllerを組み立ててルート登録する
        ...                      # 外部API連携等、上記4層に収まらないものは自由なサブパッケージにしてよい（例: ndl/, openbd/）
  ...
```

- `app/page.tsx`: ポータルトップ。`(features)/` 配下の各featureへの導線を表示する
- `app/(features)/<feature-slug>/`: featureごとに1つのポートフォリオアプリケーションを配置する。feature間の依存は持たせない
  - `(features)` はroute group（括弧付きフォルダ）とし、URLには `/features/` を出さない（例: `sample1` → `/sample1`）
  - feature専用のコンポーネント・ロジックは `_components/` `_lib/` 等のprivate folder（アンダースコア接頭辞）としてfeatureフォルダ内にcolocateし、ルーティングに含めない
- `components/ui/`: Tailwind CSS v4でスタイリングした共通UIコンポーネント（Button, Card等）を配置する。ポータルトップ・各featureはTailwindのユーティリティクラスを直接書くのではなく、原則としてここのコンポーネントを利用する。外部UIライブラリ（shadcn/ui等）は導入せず、自前のコンポーネントとして育てていく
- `server/`: Go APIを配置する独立したGoモジュール。Next.js側とは依存を持たない
  - `internal/features/<feature-slug>/`: featureごとのドメインロジックを配置する。Next.js側の`app/(features)/<feature-slug>/`とfeature-slugを揃え、どのAPIコードがどのfeatureに属すか分かるようにする。package名にハイフンは使えないため、`book-database` → `bookdatabase`のように詰めた名前にする
    - `controller` → `service` → `repository` の3層構成とする。`controller`はGinのHTTPハンドラ（リクエストの読み取りとレスポンス整形のみ）、`service`はビジネスロジック（複数repositoryの組み合わせ、トランザクション境界の管理）、`repository`はGORMによるDBアクセスに専念させる
    - `dto`はAPIリクエスト/レスポンス専用の型を置く。`internal/models`のDBエンティティ（GORM struct）とは別物とし、service層が両者を変換する
    - `router`はfeatureごとに`controller`/`service`/`repository`をコンストラクタで組み立て（DI）、ルートを登録する。`cmd/api/main.go`からはfeatureごとの`router.SetupXxxRoutes(...)`を呼ぶだけにする
    - 外部API連携（`ndl/`, `openbd/`等）のように上記4層に収まらないものは自由なサブパッケージとしてよい
  - `internal/models/`: 全featureで共有するDBエンティティ（GORM struct）を置く。テーブルと1対1対応させ、`TableName()`でテーブル名を明示する
  - 複数featureで共有する処理（GORMのDB接続等）は`internal/db/`のようにfeature非依存の場所に置く
  - `PORT` 環境変数をリッスンするGinサーバー。本番では `server/Dockerfile` でビルドしたコンテナとして動かす
  - CaddyがNext.js（web）とGo API（api）を同じ公開ドメインの配下に集約する（`/api/*`・`/health` はapiへ、それ以外はwebへ）ため、Next.js・Go APIは同一オリジンになる。それでもGo側の`ALLOWED_ORIGIN`環境変数でCORS許可オリジンを明示する
  - Next.js側はfeatureのServer Component・Client Componentともに `NEXT_PUBLIC_API_BASE_URL` 環境変数（例: `.env.example`）でGo APIのベースURLを参照する。同一オリジンの公開ドメインを指すため、SSR時のサーバー間通信もブラウザからの呼び出しも同じURLを使う

## DBマイグレーション

- [golang-migrate](https://github.com/golang-migrate/migrate) を使用。SQLファイル（up/down）を `server/migrations/` に置く。GORMの`AutoMigrate`は使わない（本番でのスキーマ変更を明示的なSQL＋レビュー可能な差分にするため）
- `internal/models/` のGORM structは既存スキーマへのマッピング専用。struct定義を変更したときは対応するマイグレーションSQLも必ず追加する
- マイグレーションファイルは `server/migrations/migrations.go` の `go:embed` でバイナリに埋め込む。本番（Railway）でも別途CLIを用意する必要がなく、`server/cmd/migrate` を実行するだけで適用できる
- 実行方法:
  ```bash
  cd server
  DATABASE_URL="mysql://app:app@tcp(127.0.0.1:3306)/my_portal" go run ./cmd/migrate up   # 適用
  DATABASE_URL="mysql://app:app@tcp(127.0.0.1:3306)/my_portal" go run ./cmd/migrate down  # 直前のマイグレーションを取り消し
  ```
- 新しいマイグレーションを追加する場合は `server/migrations/` に連番のup/downファイルを追加する（例: `000002_xxx.up.sql` / `000002_xxx.down.sql`）
- テーブル名はfeature単位で `<feature-slug>_` のprefixを付ける（例: `book-database` featureのテーブルは `book_database_books` のように `book_database_` を付与）。単一のMySQLインスタンス・単一スキーマを複数featureで共有するため、feature間のテーブル名衝突を防ぐ

## 主要な設計判断

- feature一覧はコード内で静的に定義する（例: `features` 配列を1箇所で管理し、ポータルトップがそれを参照する）。ディレクトリの自動スキャンやCMS等の動的管理は行わない
- Next.js・Go API・MySQLは単一のLightsail Instance上でDocker Composeによりまとめて動かす。マネージドサービス（Vercel/RDS等）に分散させるより、VPS1台の運用（Docker/Linux/リバースプロキシ/証明書更新）を経験することを優先
- DBアクセスはGo API経由のみとし、Next.js側から直接DBには接続しない

## ローカル開発

- MySQLは `docker-compose.yml` でコンテナ起動する（`docker compose up -d`）。ホストへのMySQLインストールは不要
- Next.js・Goはどちらもコンテナ化せず、ホスト上でネイティブ実行する（`npm run dev` / `go run ./cmd/api`）。コード変更のたびにイメージを再ビルドする必要がなく開発サイクルが速いため
- 接続情報（ユーザー: `app` / パスワード: `app` / DB名: `my_portal` / ポート: `3306`）は開発用の固定値。本番の値とは別物
- Go APIは`go run`のホットリロードがないため、Go側のコード（ルート追加等）を変更したら必ずAPIを再起動する。古いプロセスが`8080`番ポートを掴んだまま新しい`go run ./cmd/api`を実行すると`bind: address already in use`で起動に失敗し、**古いAPIが応答し続ける**（新しく追加したエンドポイントだけが`404`になるのが症状）。`lsof -i :8080`でプロセスの起動時刻を確認し、古いプロセスを`kill`してから起動し直す

## 本番デプロイ（AWS Lightsail）

- 対象: Lightsail **Instance**（VPS）1台。Lightsail Container Serviceは使わない（MySQLをコンテナで動かし永続化したいため、ステートレス前提のマネージドコンテナサービスと相性が悪い）
- 構成ファイル: リポジトリルートの `docker-compose.prod.yml`（ローカル開発用の`docker-compose.yml`とは別物）。`web`（Next.js）・`api`（Go）・`mysql`・`caddy`・`migrate`（一時実行専用）の5サービス
  - `Dockerfile`（ルート、Next.js用）: `next.config.ts` の `output: "standalone"` を前提にした multi-stage build。`NEXT_PUBLIC_API_BASE_URL` はビルド時にJSへ埋め込まれるため、`docker compose build` 時にbuild argとして渡す
  - `server/Dockerfile`: `api`・`migrate` 両方のバイナリを1つのイメージに含め、`docker-compose.prod.yml`側でどちらを起動するかを切り替える
  - `Caddyfile`: 単一の公開ドメインで、`/api/*`・`/health` を`api`コンテナへ、それ以外を`web`コンテナへリバースプロキシする。Let's Encrypt証明書の取得・更新もCaddyが自動で行う
- 環境変数は `.env.prod`（git管理しない、Lightsail Instance上にのみ置く）にまとめ、`docker compose --env-file .env.prod -f docker-compose.prod.yml ...` で読み込む（`DOMAIN` / `MYSQL_ROOT_PASSWORD` / `MYSQL_PASSWORD` / `CRON_SECRET` / `READ_ONLY`）。`READ_ONLY`の詳細は後述の「READ_ONLYモード」を参照
- マイグレーションは常時起動サービスにせず、`docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm migrate up` で都度実行する
- Lightsail側の設定: Networkingタブで80/443番ポートを開放し、インスタンスの静的IPをドメインのAレコードに割り当てる（CaddyのHTTP-01検証に必要）
- Lightsailのインスタンスはメモリが小さく、MySQL等の起動時にOOMが発生しやすいため、スワップを設定する（インスタンス初回セットアップ時に1回実行すればよい）:
  ```bash
  sudo fallocate -l 2G /swapfile
  sudo chmod 600 /swapfile
  sudo mkswap /swapfile
  sudo swapon /swapfile
  echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
  ```

## 管理者ログイン（ポートフォリオ共通）

- 目的: 各featureは基本的に「誰でも編集・削除できる公開デモ」として作る方針だが、一部のfeature（例: `simple-cms`）は特定の操作を管理者（＝サイト運営者本人）限定にしたい。featureごとに認証を作らず、ポートフォリオ全体で共通の管理者ログインを1つ持ち、各featureはそれを参照するだけにする
- 実装場所: `server/internal/admin/`（`internal/features/<feature-slug>/` の配下ではなく、`internal/db/` と同様のfeature非依存の共通処理として置く）
  - `ADMIN_PASSWORD` 環境変数と照合し、一致すれば署名付きセッションCookie（`portal_admin_session`、HMAC署名・有効期限つき、DBには保存しないステートレスな方式）を発行する
  - `POST /api/admin/login` / `POST /api/admin/logout` / `GET /api/admin/session` の3エンドポイントを `admin.RegisterRoutes` で登録する（`cmd/api/main.go` から呼ぶ）
  - `admin.AuthMiddleware()` を各featureの管理者限定エンドポイントに付与する（例: `simple-cms` の記事・カテゴリの作成/更新/削除）
- フロントエンド側: `app/login/`にポートフォリオ共通のログインページを置く（`app/(features)/`配下ではない。特定のfeatureに属さないため）
  - `app/_lib/adminAuth.ts` の `checkIsAdmin()`（Server Component専用、Cookieを中継して`GET /api/admin/session`を叩く）を各featureのページ・レイアウトから呼び、管理者向けUIを出すかどうかを判断する
  - `app/_lib/adminApi.ts` の `adminLogin` / `adminLogout`（Client Component用）
  - ログインページは `?redirect=/<feature-slug>` を受け取り、ログイン成功後に元のfeatureへ戻す
  - ポータル共通フッター（`app/_components/SiteFooter.tsx`）に `/login` への導線を常設する
- 新しいfeatureで管理者限定の操作を追加する場合は、featureごとに認証を作らず、この共通の `admin.AuthMiddleware()` / `checkIsAdmin()` を再利用すること

## CI/CD（GitHub Actions）

- `.github/workflows/ci.yml`: `main`へのpush・PRで lint / build / go vet / go test を実行する
- `.github/workflows/deploy.yml`: CIが`main`で成功したことを`workflow_run`で検知し、以下を自動実行する
  1. ルートの`Dockerfile`（web）・`server/Dockerfile`（api）をビルドし、GHCR（`ghcr.io/codebeaver50/my-portal-web` / `ghcr.io/codebeaver50/my-portal-api`）に`latest`とコミットSHAタグでpush
  2. GitHub ActionsからSSHでLightsailインスタンスに接続し、`git pull` → `docker compose pull` → `run --rm migrate up` → `up -d` を実行
  - Lightsailインスタンス自身ではビルドを行わない（メモリが小さいVPSでのビルドによるOOM・デプロイ時間の増加を避けるため）。`docker-compose.prod.yml`の`api`/`web`は`image:`でGHCRのイメージを参照しつつ、`build:`も残しているため手動でのローカルビルドも可能
  - GHCRのパッケージ（`my-portal-web` / `my-portal-api`）は初回push後に **Public** に設定しておく。Privateのままだと、Lightsail側で`docker login ghcr.io`が必要になる
  - 必要なGitHub Secrets: `DOMAIN`（`NEXT_PUBLIC_API_BASE_URL`のビルド用。`.env.prod`の`DOMAIN`と同じ値）、`LIGHTSAIL_HOST` / `LIGHTSAIL_USER` / `LIGHTSAIL_SSH_KEY`（デプロイ用SSH接続情報）、`LIGHTSAIL_APP_DIR`（インスタンス上のリポジトリのパス）
  - Lightsailインスタンス上には事前にリポジトリをclone済みで、GitHub Actionsの公開鍵に対応する秘密鍵を`LIGHTSAIL_SSH_KEY`に登録しておく必要がある（デプロイ専用ユーザー・専用鍵を推奨）
- セットアップ時に詰まりやすい点:
  - `.github/workflows/`配下を変更するコミットをHTTPSのPersonal Access Token（PAT）でpushすると、`refusing to allow a Personal Access Token to create or update workflow ... without \`workflow\` scope`で拒否される。PAT（classic）に`workflow`スコープを追加するか、fine-grained tokenなら`Workflows: Read and write`権限を付与して再発行する
  - デプロイ用SSH鍵は専用に新規生成する（`ssh-keygen -t ed25519 -f ~/.ssh/my_portal_deploy -N ""`）。`LIGHTSAIL_SSH_KEY`には**秘密鍵**の中身（`.pub`ではない方）を設定する
  - 公開鍵をLightsail側の`~/.ssh/authorized_keys`に追記する際、`echo "<公開鍵>" >> ~/.ssh/authorized_keys`を末尾に改行のないファイルに対して実行すると、既存の鍵と1行に連結されて両方とも壊れる（`ssh-keygen -lf ~/.ssh/authorized_keys`のフィンガープリントが想定と一致しない/型が化けるのが症状）。追記後は必ず`cat -A ~/.ssh/authorized_keys`等で1鍵1行になっているか確認する
  - `~/.ssh`は`700`、`authorized_keys`は`600`、ホームディレクトリはグループ/他者に書き込み権限がないこと（sshdの`StrictModes`により、権限が緩いと鍵が一致していても黙って`Permission denied (publickey)`になる）
  - 接続の切り分けには`ssh -v -i <秘密鍵> <user>@<host>`でローカルから直接SSHしてみて、GitHub Actions側の問題かサーバー側の鍵登録の問題かを先に特定するとよい

## READ_ONLYモード（公開デモの書き込み保護）

- 目的: 各featureは基本的に「誰でも編集・削除できる公開デモ」（例: `simple-ledger`）として作る方針のため、本番でシードデータやデモデータをいたずらに書き換えられたくない場合に、書き込み系エンドポイントだけを丸ごと止められるようにしている
- 挙動: `server/cmd/api/main.go` の `readOnlyMiddleware` がGinのグローバルミドルウェアとして登録されている。`READ_ONLY=true`（`.env.prod`）のとき、`GET`/`HEAD`/`OPTIONS`以外の全リクエスト（`POST`/`PUT`/`PATCH`/`DELETE`）を、除外リストに載っているルートを除いて`403`＋`{"error": "..."}`で拒否する。未設定・`false`のときは通常どおり書き込みを許可する
- **除外ルール**: `cmd/api/main.go` の `readOnlyMiddleware(readOnly, "/api/books/sync")` の第2引数以降が除外パスのリスト。あるエンドポイントを除外してよいかどうかは以下で判断する
  - ✅ 除外してよい（認可あり）: ブラウザ経由でユーザーが直接叩けない（cronや管理者のみが`Authorization: Bearer <secret>`等の別の認可を使って呼ぶ）更新系エンドポイント。例: `/api/books/sync`（`CRON_SECRET`で保護された`CronAuthMiddleware`付き）、`/api/admin/login`・`/api/admin/logout`と`simple-cms`の記事・カテゴリの作成/更新/削除（前述「管理者ログイン（ポートフォリオ共通）」の`admin.AuthMiddleware`付き）
  - ✅ 除外してよい（認可なし、DB/デモデータ非変更）: 認可のない公開エンドポイントでも、DB上のシードデータ・デモデータを一切変更せず、外部への副作用（メール送信等）のみを行うもの。例: `/api/contact`（`internal/contact`、問い合わせ内容をメール転送するだけでDBに書き込まない）。READ_ONLYの目的は「公開デモのシードデータ保護」であり、こうしたエンドポイントはその対象外になる
  - ❌ 除外してはいけない: フロントエンドの画面操作（ボタンクリック等）から到達する、認可のない**DB書き込みを伴う**公開エンドポイント。例: `simple-ledger`の取引作成・更新・削除
  - 新しいfeatureでcron専用・管理者専用の更新エンドポイントを追加した場合は、必ずそのエンドポイント自身にも認可（bearer tokenチェック等）を実装したうえで、`main.go`の除外リストに追記すること。DB/デモデータを変更する認可なしエンドポイントを除外リストに入れてはいけない
- ローカルでの動作確認:
  ```bash
  cd server
  DATABASE_URL="mysql://app:app@tcp(127.0.0.1:3306)/my_portal" READ_ONLY=true go run ./cmd/api
  ```

## カスタムフォームビルダー（form-builder）

- 目的: 項目を自由に組み合わせてフォームを作成し、集まった回答をデータベースとして一覧できるfeature。他の公開デモと同様、誰でも作成・編集・削除できる（`READ_ONLY=true`の本番では書き込みがすべて止まる。管理者限定にはしていない）
- テーブル（`server/migrations/000004_create_form_builder_tables.*.sql`、シードデータとしてフォーム3件・回答17件を含む）
  - `form_builder_forms`: フォーム本体
  - `form_builder_form_fields`: 項目定義。`type`は`text` / `textarea` / `number` / `email` / `date` / `select` / `radio` / `checkbox`。`options`（JSON配列）は`select`/`radio`/`checkbox`の選択肢で、それ以外は`[]`
    - レイアウト: `layout_row`（0始まり）が同じ項目を同じ行に横並びで表示し、行の中の並びとフォーム全体の表示順は`sort_order`で決める（行ごとに左から右）。1行あたり最大3項目
    - 幅: `layout_width`は1行を12等分した単位での幅（3〜12）。同じ行の合計は12以内で、右側に余白が残ってもよい（kintoneと同様に項目ごとに幅を持つ。行を常に埋める比率方式にはしない）
  - `form_builder_form_records`: 回答。`data`（JSON）に`{ "<項目ID>": 値 }`の形で保存する。項目ごとに列を持たないため、フォーム定義を変更してもスキーマ変更は不要
- API（`/api/form-builder`）
  ```
  GET    /forms                       フォーム一覧（項目数・回答数つき）
  POST   /forms                       フォームと項目を作成
  GET    /forms/:id                   フォーム詳細（項目つき）
  PUT    /forms/:id                   フォームと項目を更新
  DELETE /forms/:id                   フォームを削除（項目・回答はFKのON DELETE CASCADEで削除）
  GET    /forms/:id/records           回答一覧（page / pageSize、新しい順）
  POST   /forms/:id/records           回答を投稿
  DELETE /forms/:id/records/:recordId 回答を削除
  ```
- 設計判断
  - **項目IDの維持**: 回答は項目IDをキーに保存するため、フォーム更新時は項目を丸ごと入れ替えず、リクエストの`fields[].id`で既存項目を識別する（IDあり＝更新、IDなし＝新規作成、リクエストに含まれない既存項目＝削除）。項目名の変更・並べ替えをしても既存の回答との対応が崩れない
  - 削除した項目の回答値は`data`内に残るが、項目定義がないため回答データ一覧には表示しない。入力形式・選択肢を変更しても既存の回答は書き換えない（編集画面に注意書きを表示）
  - **回答の検証はGo側（`service/record_service.go`）を正とする**。必須チェック・型・メール形式・`YYYY-MM-DD`・選択肢に含まれるか・未知の項目IDの混入を検証し、エラー時は`400`＋`{"error": "...", "fieldErrors": {"<項目ID>": "..."}}`を返す。フロント側（`_lib/schema.ts`）も項目定義からzodスキーマを動的に組み立てて同じ制約で事前チェックするが、制約値（文字数上限等）を変更する場合は両方を揃えること
  - checkboxの値は回答の選択順に依らず、選択肢の定義順で保存する
  - **レイアウトのAPI表現**: 項目は表示順の平らな配列`fields[]`のまま、各項目に`row`を持たせる（行ごとの入れ子配列にはしない）。各項目は`width`も持つ。Go側（`normalizeLayout`）は`row`が配列の先頭から昇順か・1行3項目以内か・幅が3〜12で行の合計が12以内かを検証し、行番号を0からの連番に詰めて保存する。`width`が未指定（0）の項目には行を均等に割った幅を割り当てる。フロント側は`_lib/layout.ts`で行ごとの配列（`FieldDraft[][]`）との変換・挿入・移動を純粋関数として扱う
- フロントエンド（`app/(features)/form-builder/`）
  - `/form-builder`（一覧）、`/form-builder/new`・`/[formId]/edit`（フォーム構築画面）、`/[formId]`（回答画面。行ごとに横並び、スマホ幅では縦に積む）、`/[formId]/records`（回答データ一覧）
  - フォーム構築画面（`_components/builder/`）はkintoneのようなドラッグ＆ドロップ式。左のパーツ（入力形式）を右のキャンバス（回答画面と同じ見た目）へドラッグして項目を追加し、配置済みの項目もドラッグで並べ替える。行と行のあいだに落とすと新しい行、行の中の項目の左右に落とすと横並びになる。項目のクリックで設定ダイアログ（項目名・入力形式・選択肢・必須、ドラッグの代替となる「前へ」「後ろへ」）を開く
    - ドラッグ＆ドロップには**React Aria**（Adobe、`react-aria`パッケージ）の`useDrag` / `useDrop`フックのみを使う（見た目は`components/ui`のまま）。採用理由: ①キーボード（Enterで掴む→Tabで落とし先を選ぶ→Enterで置く）・スクリーンリーダーでのドラッグ操作が組み込みで用意されている、②継続的にメンテナンスされている（`@dnd-kit/core`は2024年12月以降更新が止まり後継は0.x、Atlassianのpragmatic-drag-and-dropはキーボード操作のドラッグを自前で用意する必要がある）
    - 落とし先（`DropZone`）は全ての挿入位置に常に描画しておき、ドラッグ中かつ置ける位置だけを有効にする（無効な間は`aria-hidden`・クリック不可）。各落とし先には「2行目の「氏名」の右に挿入」のような読み上げ用ラベルを付ける。満員（3項目）の行や、置いても配置が変わらない位置は無効にする
    - 幅の変更: 各項目の右端のつまみ（`WidthHandle`、`role="separator"`）を左右にドラッグすると1/12単位で変わり、キーボードではフォーカスして ←/→ で変えられる（React Aria の`useMove`）。最小幅〜行に収まる最大幅の範囲に制限する。項目を既存の行へ移動・追加して幅が足りない場合は、その項目を残りの幅に縮め、それも最小幅未満なら行を均等割りにする（`_lib/layout.ts`の`fitRow`）。回答画面はsm以上で12列グリッドに幅どおり並べ、スマホ幅では縦に積む
    - 注意: React Aria はドラッグ開始時の処理を`requestAnimationFrame`で行うため、タブが非表示（`document.visibilityState === "hidden"`）だとキーボードでのドラッグが始まらない。ブラウザ自動操作でテストする際はタブを前面に出すこと
  - `[formId]`配下のページはServer Componentで`_lib/getForm.ts`（Reactの`cache`で`generateMetadata`とページ本体の取得をまとめる）からフォームを取得し、不正なID・存在しないフォームは`notFound()`にする。一覧・回答・削除等の操作はTanStack Query（`_lib/useForms.ts`）
  - 共通UIとして`components/ui/`に`Checkbox` / `Radio`と、`Button`の`danger`バリアント（削除確認用）を追加した。`Modal`には`role="dialog"`・`aria-modal`・タイトルとの関連付けと、Escキーで閉じる処理を追加した
