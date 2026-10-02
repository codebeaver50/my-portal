package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// FormRepository は form_builder_forms テーブルにアクセスする。
type FormRepository struct {
	db *gorm.DB
}

// NewFormRepository は指定のコネクションに紐づく FormRepository を作成する。
func NewFormRepository(db *gorm.DB) *FormRepository {
	return &FormRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく FormRepository を返す。
func (r *FormRepository) WithTx(tx *gorm.DB) *FormRepository {
	return &FormRepository{db: tx}
}

func (r *FormRepository) preload(db *gorm.DB) *gorm.DB {
	return db.Preload("Fields", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	})
}

// Create は新しいフォームの行を挿入する（項目は含めない）。成功時、form.ID が採番される。
func (r *FormRepository) Create(ctx context.Context, form *models.Form) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(form).Error
}

// GetByID はIDで項目定義つきのフォームを返す。存在しない場合は nil を返す。
func (r *FormRepository) GetByID(ctx context.Context, id uint) (*models.Form, error) {
	var form models.Form
	err := r.preload(r.db.WithContext(ctx)).First(&form, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &form, nil
}

// GetAll は作成日時が新しい順に項目定義つきのすべてのフォームを返す。
func (r *FormRepository) GetAll(ctx context.Context) ([]models.Form, error) {
	var forms []models.Form
	err := r.preload(r.db.WithContext(ctx)).Order("created_at DESC, id DESC").Find(&forms).Error
	return forms, err
}

// Update は既存のフォームの行を上書きする（項目は含めない）。
func (r *FormRepository) Update(ctx context.Context, form *models.Form) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Save(form).Error
}

// Delete はフォームを削除する。
func (r *FormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Form{}, id).Error
}
