#!/usr/bin/env bash
# ローカルの開発環境を対話的にセットアップする（README.md の「開発環境のセットアップ」と同じ手順）。
# 済んでいる手順は確認して飛ばすため、何度実行してもよい。既存のファイル（.env 等）は上書きしない。
#
#   ./setup.sh       各手順の前に確認する
#   ./setup.sh -y    確認せずにすべての手順を進める（管理者アカウントの作成だけは行わない）
#
# macOS 標準の bash 3.2 でも動くように書く（連想配列や ${var,,} 等は使わない）。
set -euo pipefail

cd "$(dirname "$0")"

ASSUME_YES=false
if [ "${1:-}" = "-y" ] || [ "${1:-}" = "--yes" ]; then
  ASSUME_YES=true
fi

step() { printf '\n\033[1;34m==> %s\033[0m\n' "$1"; }
info() { printf '    %s\n' "$1"; }
ok() { printf '    \033[32m✓\033[0m %s\n' "$1"; }
warn() { printf '    \033[33m!\033[0m %s\n' "$1"; }
fail() {
  printf '    \033[31m✗\033[0m %s\n' "$1" >&2
  exit 1
}

# confirm "質問" [y|n] : 既定値は y。-y のときは確認せずに既定値を使う
confirm() {
  local default="${2:-y}" answer hint="[Y/n]"
  [ "$default" = "n" ] && hint="[y/N]"
  if [ "$ASSUME_YES" = true ]; then
    [ "$default" = "y" ]
    return
  fi
  read -r -p "    $1 $hint " answer
  answer="${answer:-$default}"
  case "$answer" in
    [Yy]*) return 0 ;;
    *) return 1 ;;
  esac
}

# 前提の確認 ---------------------------------------------------------------

step "前提の確認"
[ "$(uname -s)" = "Darwin" ] || fail "macOS 以外には対応していません。README.md の手順を参考に手動でセットアップしてください"
# Homebrew を入れた直後で、シェルの設定（~/.zprofile）がまだ読み込まれていない場合に備える
if ! command -v brew >/dev/null 2>&1 && [ -x /opt/homebrew/bin/brew ]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
fi
if ! command -v brew >/dev/null 2>&1; then
  fail "Homebrew がありません。https://brew.sh/ の手順で入れてから、もう一度実行してください"
fi
ok "macOS / Homebrew（$(brew --prefix)）"

# 1. proto・Docker Desktop -------------------------------------------------

step "proto と Docker Desktop（Brewfile）"
# brew bundle だと、Homebrew 以外の方法で入れた proto・Docker Desktop も「ない」と判断して
# 二重に入れようとするため、コマンド・アプリがあるかで判断する
if command -v proto >/dev/null 2>&1 || [ -x "$HOME/.proto/bin/proto" ]; then
  ok "proto は入っています"
elif confirm "proto を Homebrew で入れますか？"; then
  brew install proto
else
  fail "proto がないと続けられません"
fi

if [ -d /Applications/Docker.app ]; then
  ok "Docker Desktop は入っています"
elif confirm "Docker Desktop を Homebrew で入れますか？"; then
  brew install --cask docker-desktop
else
  warn "Docker Desktop を入れずに進めます（DB の起動とマイグレーションは飛ばします）"
fi

# 2. proto のシェル設定 -----------------------------------------------------

step "proto のシェル設定"
export PROTO_HOME="${PROTO_HOME:-$HOME/.proto}"
case ":$PATH:" in
  *":$PROTO_HOME/shims:"*)
    ok "PATH に proto の shims が入っています"
    ;;
  *)
    if confirm "proto setup でシェルの設定（PATH 等）を追加しますか？"; then
      proto setup
      warn "このスクリプトの後でターミナルを開き直してください"
    else
      warn "シェルの設定を飛ばしました。node・go 等のコマンドが proto のバージョンにならない場合があります"
    fi
    ;;
esac
# このスクリプトの中では、Homebrew 等で入った同名のコマンド（go 等）より proto のものを優先する
export PATH="$PROTO_HOME/shims:$PROTO_HOME/bin:$PATH"

# 3. Node.js・Go・golangci-lint --------------------------------------------

step "Node.js・Go・golangci-lint（.nvmrc / .prototools）"
if confirm "proto install でバージョンを揃えますか？（入っているものは飛ばされます）"; then
  proto install
fi
# 別の方法で入れた同名のコマンドが PATH で先にあると、proto のバージョンにならないため確かめる
# （.prototools の [settings] / [plugins] より前の行がバージョンの指定）
tool_version() { awk -F'"' -v name="$1 = " '/^\[/ { exit } $1 == name { print $2 }' .prototools; }
check_version() {
  local name="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    ok "$name $actual"
  else
    warn "$name が $actual です（期待するバージョン: $expected）。$(command -v "$name") が使われています。PATH で ~/.proto/shims が先になるようにしてください"
  fi
}
node_major="$(node --version | sed 's/^v\([0-9]*\).*/\1/')"
check_version node "$(cat .nvmrc)" "$node_major"
check_version go "$(tool_version go)" "$(go env GOVERSION | sed 's/^go//')"
check_version golangci-lint "$(tool_version golangci-lint)" "$(golangci-lint --version | awk '{print $4}')"

# 4. npm の依存パッケージ ---------------------------------------------------

step "npm の依存パッケージ"
if [ -d node_modules ]; then
  if confirm "node_modules があります。npm ci で入れ直しますか？" n; then
    npm ci
  fi
elif confirm "npm ci を実行しますか？"; then
  npm ci
fi

# 5. 環境変数のファイル -----------------------------------------------------

step "環境変数のファイル"
copy_env() {
  local src="$1" dest="$2"
  if [ -f "$dest" ]; then
    ok "$dest はあります（上書きしません）"
  elif confirm "$src から $dest を作りますか？"; then
    cp "$src" "$dest"
    ok "$dest を作りました"
  fi
}
copy_env .env.example .env.local
copy_env server/.env.example server/.env

# 6. DB の起動とマイグレーション ---------------------------------------------

step "DB の起動とマイグレーション"
db_ready=false
if [ ! -d /Applications/Docker.app ]; then
  warn "Docker Desktop がないため飛ばします"
elif [ ! -f server/.env ]; then
  warn "server/.env がないため飛ばします"
elif confirm "MySQL を起動して、マイグレーションを適用しますか？"; then
  if ! docker info >/dev/null 2>&1; then
    info "Docker Desktop を起動しています..."
    open -a Docker
    for _ in $(seq 1 60); do
      docker info >/dev/null 2>&1 && break
      sleep 2
    done
    docker info >/dev/null 2>&1 || fail "Docker Desktop が起動しませんでした。起動してからもう一度実行してください"
  fi
  # healthcheck が通る（MySQL が接続を受け付ける）まで待つ
  docker compose up -d --wait
  (
    cd server
    set -a
    # shellcheck disable=SC1091
    source .env
    set +a
    go run ./cmd/migrate up
  )
  db_ready=true
  ok "MySQL とマイグレーションの準備ができました"
fi

# 7. 管理者アカウント（任意） -----------------------------------------------

if [ "$db_ready" = true ] && [ "$ASSUME_YES" = false ]; then
  step "管理者アカウント（任意）"
  info "デモ用アカウント demo-guest / demo-pass は最初から使えます"
  if confirm "管理者アカウントを作成しますか？" n; then
    read -r -p "    ログインID: " login_id
    read -r -p "    表示名: " display_name
    (
      cd server
      set -a
      # shellcheck disable=SC1091
      source .env
      set +a
      # パスワードは createuser が端末から入力を表示せずに読む
      go run ./cmd/createuser -login "$login_id" -name "$display_name" -role admin
    )
  fi
fi

# 完了 -----------------------------------------------------------------------

step "完了"
cat <<'EOF'
    起動するには、2つのターミナルで次を実行してください。

      # Go API（http://localhost:8080）
      cd server && set -a && source .env && set +a && go run ./cmd/api

      # Next.js（http://localhost:3000）
      npm run dev
EOF
