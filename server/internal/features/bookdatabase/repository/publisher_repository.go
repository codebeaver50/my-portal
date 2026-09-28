package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// PublisherRepository は book_database_publishers テーブルにアクセスする。
type PublisherRepository struct {
	db *gorm.DB
}

// NewPublisherRepository は指定のコネクションに紐づく PublisherRepository を作成する。
func NewPublisherRepository(db *gorm.DB) *PublisherRepository {
	return &PublisherRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく PublisherRepository を返す。
func (r *PublisherRepository) WithTx(tx *gorm.DB) *PublisherRepository {
	return &PublisherRepository{db: tx}
}

// FindByName は指定の名前の出版社を返す。存在しない場合は nil を返す。
func (r *PublisherRepository) FindByName(ctx context.Context, name string) (*models.Publisher, error) {
	var publisher models.Publisher
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&publisher).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &publisher, nil
}

// Create は新しい出版社の行を挿入する。
func (r *PublisherRepository) Create(ctx context.Context, publisher *models.Publisher) error {
	return r.db.WithContext(ctx).Create(publisher).Error
}

// FindByNames は指定の名前のうち存在する出版社をまとめて返す。
func (r *PublisherRepository) FindByNames(ctx context.Context, names []string) ([]models.Publisher, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var publishers []models.Publisher
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&publishers).Error
	return publishers, err
}

// CreateBatch は新しい出版社の行をまとめて挿入する。
func (r *PublisherRepository) CreateBatch(ctx context.Context, publishers []*models.Publisher) error {
	if len(publishers) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(publishers).Error
}
