package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// newTestDB は simple cms 関連のテーブルをマイグレーション済みの、テスト
// 専用のインメモリSQLite DBを返す。単一コネクションに固定し、複数コネクションが
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

	if err := db.AutoMigrate(&models.Category{}, &models.Post{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

// newServices はテスト用DBに紐づく CategoryService と PostService を返す。
func newServices(t *testing.T) (CategoryService, PostService) {
	t.Helper()

	db := newTestDB(t)
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db)
	return NewCategoryService(categoryRepo, postRepo), NewPostService(postRepo, categoryRepo)
}
