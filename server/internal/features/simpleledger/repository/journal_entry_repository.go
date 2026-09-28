package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// JournalEntryRepository は simple_ledger_journal_entries テーブルにアクセスする。
type JournalEntryRepository struct {
	db *gorm.DB
}

// NewJournalEntryRepository は指定のコネクションに紐づく JournalEntryRepository を作成する。
func NewJournalEntryRepository(db *gorm.DB) *JournalEntryRepository {
	return &JournalEntryRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく JournalEntryRepository を返す。
func (r *JournalEntryRepository) WithTx(tx *gorm.DB) *JournalEntryRepository {
	return &JournalEntryRepository{db: tx}
}

// CreateBatch は複数の仕訳エントリーをまとめて挿入する。
func (r *JournalEntryRepository) CreateBatch(ctx context.Context, entries []models.JournalEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&entries).Error
}

// DeleteByTransactionID は指定の取引に紐づく仕訳エントリーをすべて削除する。
func (r *JournalEntryRepository) DeleteByTransactionID(ctx context.Context, transactionID uint) error {
	return r.db.WithContext(ctx).
		Where("transaction_id = ?", transactionID).
		Delete(&models.JournalEntry{}).Error
}
