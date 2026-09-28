package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// PostRepository は simple_cms_posts テーブルにアクセスする。
type PostRepository struct {
	db *gorm.DB
}

// NewPostRepository は指定のコネクションに紐づく PostRepository を作成する。
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) preload(db *gorm.DB) *gorm.DB {
	return db.Preload("Category")
}

// Create は新しい記事の行を挿入する。成功時、post.ID が採番される。
func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// GetByID はIDで記事を返す。存在しない場合は nil を返す。
func (r *PostRepository) GetByID(ctx context.Context, id uint) (*models.Post, error) {
	var post models.Post
	err := r.preload(r.db.WithContext(ctx)).First(&post, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetBySlug はslugで記事を返す。存在しない場合は nil を返す。
func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*models.Post, error) {
	var post models.Post
	err := r.preload(r.db.WithContext(ctx)).Where("slug = ?", slug).First(&post).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetAll は作成日時が新しい順にすべての記事を返す。
func (r *PostRepository) GetAll(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post
	err := r.preload(r.db.WithContext(ctx)).Order("created_at DESC").Find(&posts).Error
	return posts, err
}

// Update は既存の記事の行を上書きする。
func (r *PostRepository) Update(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// Delete は記事を削除する。
func (r *PostRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Post{}, id).Error
}
