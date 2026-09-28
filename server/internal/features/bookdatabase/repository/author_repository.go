package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// AuthorRepository は book_database_authors テーブルにアクセスする。
type AuthorRepository struct {
	db *gorm.DB
}

// NewAuthorRepository は指定のコネクションに紐づく AuthorRepository を作成する。
func NewAuthorRepository(db *gorm.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく AuthorRepository を返す。
func (r *AuthorRepository) WithTx(tx *gorm.DB) *AuthorRepository {
	return &AuthorRepository{db: tx}
}

// FindByName は指定の名前の著者を返す。存在しない場合は nil を返す。
func (r *AuthorRepository) FindByName(ctx context.Context, name string) (*models.Author, error) {
	var author models.Author
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&author).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &author, nil
}

// Create は新しい著者の行を挿入する。
func (r *AuthorRepository) Create(ctx context.Context, author *models.Author) error {
	return r.db.WithContext(ctx).Create(author).Error
}

// FindByNames は指定の名前のうち存在する著者をまとめて返す。
func (r *AuthorRepository) FindByNames(ctx context.Context, names []string) ([]models.Author, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var authors []models.Author
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&authors).Error
	return authors, err
}

// CreateBatch は新しい著者の行をまとめて挿入する。
func (r *AuthorRepository) CreateBatch(ctx context.Context, authors []*models.Author) error {
	if len(authors) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(authors).Error
}
