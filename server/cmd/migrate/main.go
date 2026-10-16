// Command migrate は server/migrations に埋め込んだSQLマイグレーションを、
// DATABASE_URL のDBに適用する（または取り消す）。ローカルでも本番でも同じように
// 実行でき、マイグレーション用のCLIを別途インストールする必要がない。
package main

import (
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/codebeaver50/my-portal/server/migrations"
)

func main() {
	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("failed to load migrations: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}

	switch direction {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	default:
		log.Fatalf("unknown direction %q (use \"up\" or \"down\")", direction)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration complete")
}
