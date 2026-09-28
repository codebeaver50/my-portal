package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// newTestDB は simple ledger 関連のテーブルをマイグレーション済みの、テスト
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

	if err := db.AutoMigrate(
		&models.ChartOfAccounts{},
		&models.Transaction{},
		&models.JournalEntry{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

// seedAccounts はテスト用の勘定科目を2件（現金・売上）作成し、それぞれのIDを返す。
func seedAccounts(t *testing.T, db *gorm.DB) (cashID, revenueID uint) {
	t.Helper()

	cash := models.ChartOfAccounts{Code: "1000", Name: "現金", Type: models.AssetAccount, NormalBalance: models.DebitBalance}
	if err := db.Create(&cash).Error; err != nil {
		t.Fatalf("failed to seed cash account: %v", err)
	}
	revenue := models.ChartOfAccounts{Code: "4000", Name: "売上", Type: models.RevenueAccount, NormalBalance: models.CreditBalance}
	if err := db.Create(&revenue).Error; err != nil {
		t.Fatalf("failed to seed revenue account: %v", err)
	}

	return cash.ID, revenue.ID
}
