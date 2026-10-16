// Command createuser はポータル共通のアクセス管理（RBAC）のユーザーを作成する。
// サイト運営者本人の管理者アカウントのように、パスワードをリポジトリ（マイグレーションのシード）に
// 残したくないユーザーを作るために使う。
//
// パスワードはコマンドライン引数ではなく標準入力から読む（シェルの履歴やプロセス一覧に残さないため）。
// 端末から実行した場合は入力を画面に表示せず、確認のためにもう一度入力させる。
//
//	DATABASE_URL="mysql://app:app@tcp(127.0.0.1:3306)/my_portal" go run ./cmd/createuser -login owner -name "サイト管理者" -role admin
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
	"github.com/codebeaver50/my-portal/server/internal/db"
)

func main() {
	loginID := flag.String("login", "", "ログインID（必須）")
	displayName := flag.String("name", "", "表示名（必須）")
	roleCode := flag.String("role", service.AdminRoleCode, "割り当てるロールのcode")
	flag.Parse()

	if *loginID == "" || *displayName == "" {
		flag.Usage()
		os.Exit(2)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	password, err := readPassword()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := db.Open(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	users := service.NewUserService(conn, repository.NewUserRepository(conn), repository.NewRoleRepository(conn))
	created, err := users.Create(context.Background(), &dto.CreateUserRequest{
		LoginID:     *loginID,
		DisplayName: *displayName,
		Password:    password,
		RoleCode:    *roleCode,
	})
	if err != nil {
		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			log.Fatal(validationErr.Message)
		}
		log.Fatalf("failed to create user: %v", err)
	}

	log.Printf("created user %q (id=%d, role=%s)", created.LoginID, created.ID, created.RoleCode)
}

// readPassword は標準入力からパスワードを読む。端末なら入力を表示せずに2回入力させ、
// パイプ等の場合は1行目をそのまま使う。
func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("failed to read password from stdin: %w", err)
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	fmt.Fprint(os.Stderr, "パスワード: ")
	first, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Fprint(os.Stderr, "パスワード（確認）: ")
	second, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	if string(first) != string(second) {
		return "", errors.New("パスワードが一致しません")
	}
	return string(first), nil
}
