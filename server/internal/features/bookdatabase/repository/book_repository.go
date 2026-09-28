// Package repository は book database feature のDBアクセス（読み書き）を実装する。
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// BookRepository は book_database_books と book_database_book_sources
// テーブルにアクセスする。
type BookRepository struct {
	db *gorm.DB
}

// NewBookRepository は指定のコネクションに紐づく BookRepository を作成する。
func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく BookRepository を返す。
func (r *BookRepository) WithTx(tx *gorm.DB) *BookRepository {
	return &BookRepository{db: tx}
}

// List は発行日が新しい順にskipから最大take件の書籍を返す。Publisherと著者の
// 関連（表示順で並んだ著者）もpreloadする。
func (r *BookRepository) List(ctx context.Context, skip, take int) ([]models.Book, error) {
	var books []models.Book
	err := r.db.WithContext(ctx).
		Preload("Publisher").
		Preload("BookAuthors", func(db *gorm.DB) *gorm.DB {
			return db.Order("`order` ASC")
		}).
		Preload("BookAuthors.Author").
		Order("published_date DESC, id DESC").
		Offset(skip).
		Limit(take).
		Find(&books).Error
	return books, err
}

// GetByID はIDで書籍を返す。存在しない場合は nil を返す。
func (r *BookRepository) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	var book models.Book
	err := r.db.WithContext(ctx).First(&book, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// FindByISBN13 は指定のISBN-13の書籍を返す。isbn13がnil/空、または該当する
// 書籍がない場合は nil を返す。
func (r *BookRepository) FindByISBN13(ctx context.Context, isbn13 *string) (*models.Book, error) {
	if isbn13 == nil || *isbn13 == "" {
		return nil, nil
	}
	var book models.Book
	err := r.db.WithContext(ctx).Where("isbn13 = ?", *isbn13).First(&book).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// Create は新しい書籍の行を挿入する。成功時、book.ID が採番される。
func (r *BookRepository) Create(ctx context.Context, book *models.Book) error {
	return r.db.WithContext(ctx).Create(book).Error
}

// bookUpdatableColumns はUpdateが上書きする書籍のカラム一覧。created_atは
// 意図的に除外し、updated_atはGORMの自動タイムスタンプ更新に任せる。
var bookUpdatableColumns = []string{
	"isbn13", "title", "subtitle", "publisher_id", "published_date",
	"series_name", "volume", "price", "cover_image_url",
}

// Update は既存の書籍の行のbookUpdatableColumnsを上書きする。
func (r *BookRepository) Update(ctx context.Context, book *models.Book) error {
	return r.db.WithContext(ctx).
		Model(&models.Book{}).
		Where("id = ?", book.ID).
		Select(bookUpdatableColumns).
		Updates(book).Error
}

// FindSource は種別とソースIDから、記録済みのsyncソースを探す。存在しない
// 場合は nil を返す。
func (r *BookRepository) FindSource(ctx context.Context, sourceType models.SourceType, sourceID string) (*models.BookSource, error) {
	var source models.BookSource
	err := r.db.WithContext(ctx).
		Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		First(&source).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// CreateSource は書籍の新しいsyncソースを記録する。
func (r *BookRepository) CreateSource(ctx context.Context, source *models.BookSource) error {
	return r.db.WithContext(ctx).Create(source).Error
}

// UpdateSource は記録済みのsyncソースの生データを更新する。
func (r *BookRepository) UpdateSource(ctx context.Context, source *models.BookSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

// ReplaceAuthors は書籍の著者リンクを指定の行で置き換える。
func (r *BookRepository) ReplaceAuthors(ctx context.Context, bookID int64, bookAuthors []models.BookAuthor) error {
	db := r.db.WithContext(ctx)
	if err := db.Where("book_id = ?", bookID).Delete(&models.BookAuthor{}).Error; err != nil {
		return err
	}
	if len(bookAuthors) == 0 {
		return nil
	}
	return db.Create(&bookAuthors).Error
}
