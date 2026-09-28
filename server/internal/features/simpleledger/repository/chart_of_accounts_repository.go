// Package repository は simple ledger feature のDBアクセス（読み書き）を実装する。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// ChartOfAccountsRepository は simple_ledger_chart_of_accounts テーブルにアクセスする。
type ChartOfAccountsRepository struct {
	db *gorm.DB
}

// NewChartOfAccountsRepository は指定のコネクションに紐づく ChartOfAccountsRepository を作成する。
func NewChartOfAccountsRepository(db *gorm.DB) *ChartOfAccountsRepository {
	return &ChartOfAccountsRepository{db: db}
}

// GetActive は有効な勘定科目を科目コード順に返す。
func (r *ChartOfAccountsRepository) GetActive(ctx context.Context) ([]models.ChartOfAccounts, error) {
	var accounts []models.ChartOfAccounts
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("code ASC").
		Find(&accounts).Error
	return accounts, err
}
