package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// newTestDB は書籍関連のテーブルをマイグレーション済みの、テスト専用の
// インメモリSQLite DBを返す。単一コネクションに固定し、複数コネクションが
// それぞれ別の空DBを見てしまうのを防ぐ。本番のMySQLスキーマ管理（golang-migrate）
// とは無関係の、テスト内でのみ使い捨てるDBであるため AutoMigrate を使う。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.AutoMigrate(
		&models.Publisher{},
		&models.Author{},
		&models.Book{},
		&models.BookAuthor{},
		&models.BookSource{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}
