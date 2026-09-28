package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// TransactionRepository は simple_ledger_transactions テーブルにアクセスする。
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository は指定のコネクションに紐づく TransactionRepository を作成する。
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく TransactionRepository を返す。
func (r *TransactionRepository) WithTx(tx *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: tx}
}

func (r *TransactionRepository) preload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("JournalEntries").
		Preload("JournalEntries.ChartOfAccounts")
}

// Create は新しい取引の行を挿入する。成功時、transaction.ID が採番される。
func (r *TransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetByID はIDで取引を返す。存在しない場合は nil を返す。
func (r *TransactionRepository) GetByID(ctx context.Context, id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.preload(r.db.WithContext(ctx)).First(&transaction, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetAll は取引日が新しい順にすべての取引を返す。
func (r *TransactionRepository) GetAll(ctx context.Context) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.preload(r.db.WithContext(ctx)).
		Order("date DESC, created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// GetPaginated は取引日が新しい順にページネーション付きで取引を返す。keywordが
// 指定された場合はdescriptionの部分一致で絞り込む。
func (r *TransactionRepository) GetPaginated(ctx context.Context, page, pageSize int, keyword string) ([]models.Transaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Transaction{})
	if keyword != "" {
		query = query.Where("description LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var transactions []models.Transaction
	offset := (page - 1) * pageSize
	err := r.preload(query).
		Order("date DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// Update は既存の取引の行を上書きする。
func (r *TransactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// Delete は取引を削除する。
func (r *TransactionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Transaction{}, id).Error
}
