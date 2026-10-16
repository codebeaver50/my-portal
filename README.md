# my-portal

ポートフォリオ用のアプリケーション（feature）をまとめたポータルサイト。フロントエンドは Next.js、API は Go（Gin / GORM）、DB は MySQL。

- 仕様: [SPEC.md](./SPEC.md)
- 構成・設計判断・本番デプロイ（AWS Lightsail / Docker Compose）: [ARCHITECTURE.md](./ARCHITECTURE.md)
- コーディングルール: [CODING_RULES.md](./CODING_RULES.md)

## 開発環境のセットアップ

macOS（Apple Silicon）と [Homebrew](https://brew.sh/) を前提にする。ツールは次の2つに分けて管理している。

| 対象 | 管理するファイル | 入れ方 |
|---|---|---|
| Mac に1つあればよいもの（proto・Docker Desktop） | `Brewfile` | `brew bundle` |
| バージョンを揃えるもの（Node.js・Go・golangci-lint） | `.nvmrc`（Node.js）・`.prototools`（Go・golangci-lint） | `proto install` |

以下の手順 1〜3 は、`./setup.sh` で対話的に進めることもできる。各手順の前に確認し、済んでいる手順は飛ばす（何度実行してもよい。`.env` 等の既存のファイルは上書きしない）。`./setup.sh -y` で確認せずに進める。

```bash
git clone https://github.com/codebeaver50/my-portal.git
cd my-portal
./setup.sh
```

### 1. ツールを入れる

```bash
git clone https://github.com/codebeaver50/my-portal.git
cd my-portal

brew bundle      # proto と Docker Desktop
proto setup      # 初回のみ。シェルの設定（PATH 等）を追加する。終わったらターミナルを開き直す
proto install    # .nvmrc の Node.js と、.prototools の Go・golangci-lint
npm ci
```

`node --version`・`go version`・`golangci-lint --version` が、`.nvmrc`・`.prototools` のバージョンになっていることを確認する。

### 2. 環境変数のファイルを作る

```bash
cp .env.example .env.local               # Next.js 用
cp server/.env.example server/.env       # Go API 用（ローカルでは値を変えずに使える）
```

どちらも git の管理外。本番用の `.env.prod` は Lightsail Instance 上にだけ置く。

### 3. DB を起動してマイグレーションを適用する

Docker Desktop を起動してから実行する。

```bash
docker compose up -d                     # MySQL と mailpit（問い合わせメールの確認用）

cd server
set -a; source .env; set +a
go run ./cmd/migrate up
```

### 4. 起動する

```bash
# ターミナル1: Go API（http://localhost:8080）
cd server
set -a; source .env; set +a
go run ./cmd/api

# ターミナル2: Next.js（http://localhost:3000）
npm run dev
```

Go API はホットリロードしないため、Go のコードを変更したら再起動する（`ARCHITECTURE.md` の「ローカル開発」を参照）。

デモポータルには、デモ用アカウント（ログインID `demo-guest` / パスワード `demo-pass`）でログインできる。管理者アカウントは `server/cmd/createuser` で作成する（`ARCHITECTURE.md` の「アクセス管理（RBAC）」を参照）。

## チェック

CI（`.github/workflows/ci.yml`）と同じチェックを手元で実行できる。

```bash
npm run lint
npm run build

cd server
go vet ./...
golangci-lint run ./...
go test ./...
```

## ツールのバージョンを上げるとき

| ツール | 変更するファイル |
|---|---|
| Node.js | `.nvmrc`（CI もここから読む） |
| Go | `.prototools` の `go` と `server/go.mod` の `go` の**両方**（揃っていないと CI が失敗する） |
| golangci-lint | `.prototools` の `golangci-lint`（CI もここから読む） |

変更したら `proto install` を実行して、手元のツールを入れ直す。
