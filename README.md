# my-portal

ポートフォリオ用のアプリケーション（feature）をまとめたポータルサイト。フロントエンドは Next.js、API は Go（Gin / GORM）、DB は MySQL。

- 仕様: [SPEC.md](./SPEC.md)
- 構成・設計判断・本番デプロイ（AWS Lightsail / Docker Compose）: [ARCHITECTURE.md](./ARCHITECTURE.md)
- コーディングルール: [CODING_RULES.md](./CODING_RULES.md)

## 開発環境のセットアップ

macOS（Apple Silicon）と [Homebrew](https://brew.sh/) を前提にする。ツールは次のように分けて管理している。

| 対象 | 管理するファイル | 入れ方 |
|---|---|---|
| Mac に1つあればよいもの（proto） | `Brewfile` | `brew bundle` |
| コンテナの実行環境（ローカルの MySQL 用） | なし（Docker Desktop と Rancher Desktop のどちらでもよい） | 各自で入れる |
| バージョンを揃えるもの（Node.js・Go・golangci-lint） | `.nvmrc`（Node.js）・`.prototools`（Go・golangci-lint） | `proto install` |

以下の手順 1〜3 は、`./setup.sh` で対話的に進めることもできる。各手順の前に確認し、済んでいる手順は飛ばす（何度実行してもよい。既存の `.env`・`server/.env` は上書きしない）。`./setup.sh -y` で確認せずに進める。

```bash
git clone https://github.com/codebeaver50/my-portal.git
cd my-portal
./setup.sh
```

### 1. ツールを入れる

コンテナの実行環境として、[Docker Desktop](https://www.docker.com/products/docker-desktop/) か [Rancher Desktop](https://rancherdesktop.io/) を入れておく（`brew install --cask docker-desktop` / `brew install --cask rancher-desktop`）。Rancher Desktop の場合は、コンテナエンジンに「dockerd (moby)」を選ぶ（`docker` コマンドを使うため）。

```bash
git clone https://github.com/codebeaver50/my-portal.git
cd my-portal

brew bundle      # proto
proto setup      # 初回のみ。シェルの設定（PATH 等）を追加する。終わったらターミナルを開き直す
proto install    # .nvmrc の Node.js と、.prototools の Go・golangci-lint
npm ci
```

`node --version`・`go version`・`golangci-lint --version` が、`.nvmrc`・`.prototools` のバージョンになっていることを確認する。

### 2. 環境変数のファイルを作る

```bash
cp .env.example .env                     # Next.js 用
cp server/.env.example server/.env       # Go API 用
```

| ファイル | 読むもの | 中身 |
|---|---|---|
| `.env` | Next.js（自動で読み込む） | `.env.example` のコピー。「本番のみ」の節は空のままでよい |
| `server/.env` | Go API（`go run` の前に `set -a; source .env; set +a` で環境変数にする） | `server/.env.example` のコピー。ローカルでは値を変えずに使える |

問い合わせフォームのメールを mailpit で確認したい場合は、`server/.env` の `CONTACT_SMTP_*` に `server/.env.example` のコメントにある値を入れる（`./setup.sh` で作るとこの値が入る）。

どちらも git の管理外。本番では、ルートの `.env.example`（本番のキーも含む）をコピーして Lightsail Instance 上に `.env.prod` を作る。

### 3. DB を起動してマイグレーションを適用する

Docker Desktop か Rancher Desktop を起動してから実行する。

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
