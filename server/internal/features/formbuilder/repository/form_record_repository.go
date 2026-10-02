package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// FormRecordRepository は form_builder_form_records テーブルにアクセスする。
type FormRecordRepository struct {
	db *gorm.DB
}

// NewFormRecordRepository は指定のコネクションに紐づく FormRecordRepository を作成する。
func NewFormRecordRepository(db *gorm.DB) *FormRecordRepository {
	return &FormRecordRepository{db: db}
}

// Create は新しい回答の行を挿入する。成功時、record.ID が採番される。
func (r *FormRecordRepository) Create(ctx context.Context, record *models.FormRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByID はIDで回答を返す。存在しない場合は nil を返す。
func (r *FormRecordRepository) GetByID(ctx context.Context, id uint) (*models.FormRecord, error) {
	var record models.FormRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetPaginatedByFormID は指定のフォームの回答を、投稿日時が新しい順にページネーション付きで返す。
func (r *FormRecordRepository) GetPaginatedByFormID(ctx context.Context, formID uint, page, pageSize int) ([]models.FormRecord, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.FormRecord{}).Where("form_id = ?", formID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.FormRecord
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC, id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// CountByFormID はフォームIDごとの回答件数を返す。回答が0件のフォームは含まれない。
func (r *FormRecordRepository) CountByFormID(ctx context.Context) (map[uint]int64, error) {
	var rows []struct {
		FormID uint
		Count  int64
	}
	err := r.db.WithContext(ctx).
		Model(&models.FormRecord{}).
		Select("form_id, COUNT(*) AS count").
		Group("form_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		counts[row.FormID] = row.Count
	}
	return counts, nil
}

// Delete は回答を削除する。
func (r *FormRecordRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.FormRecord{}, id).Error
}
