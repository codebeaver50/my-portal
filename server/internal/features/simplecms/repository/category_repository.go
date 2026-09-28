// Package repository は simple cms feature のDBアクセス（読み書き）を実装する。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// CategoryRepository は simple_cms_categories テーブルにアクセスする。
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository は指定のコネクションに紐づく CategoryRepository を作成する。
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create は新しいカテゴリの行を挿入する。成功時、category.ID が採番される。
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID はIDでカテゴリを返す。存在しない場合は nil を返す。
func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetBySlug はslugでカテゴリを返す。存在しない場合は nil を返す。
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetAll はカテゴリ名の昇順ですべてのカテゴリを返す。
func (r *CategoryRepository) GetAll(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	return categories, err
}

// Update は既存のカテゴリの行を上書きする。
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete はカテゴリを削除する。紐づく記事が存在する場合はFKのON DELETE
// RESTRICTによりDBエラーが返る。
func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error
}
