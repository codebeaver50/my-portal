package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// FormFieldRepository は form_builder_form_fields テーブルにアクセスする。
type FormFieldRepository struct {
	db *gorm.DB
}

// NewFormFieldRepository は指定のコネクションに紐づく FormFieldRepository を作成する。
func NewFormFieldRepository(db *gorm.DB) *FormFieldRepository {
	return &FormFieldRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく FormFieldRepository を返す。
func (r *FormFieldRepository) WithTx(tx *gorm.DB) *FormFieldRepository {
	return &FormFieldRepository{db: tx}
}

// CreateBatch は複数の項目の行をまとめて挿入する。
func (r *FormFieldRepository) CreateBatch(ctx context.Context, fields []models.FormField) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&fields).Error
}

// Update は既存の項目の行を上書きする。
func (r *FormFieldRepository) Update(ctx context.Context, field *models.FormField) error {
	return r.db.WithContext(ctx).Save(field).Error
}

// DeleteByIDs は指定のIDの項目をまとめて削除する。
func (r *FormFieldRepository) DeleteByIDs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&models.FormField{}, ids).Error
}
