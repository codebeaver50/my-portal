package models

import "time"

// Form: フォームビルダーで作成されたフォーム
type Form struct {
	ID          uint      `gorm:"column:id;primaryKey" json:"id"`
	Title       string    `gorm:"column:title;size:100;not null" json:"title"`
	Description string    `gorm:"column:description;size:500;not null;default:''" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`

	Fields []FormField `gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE" json:"fields,omitempty"`
}

// Form 構造体は form_builder_forms テーブルにマッピングされる
func (Form) TableName() string {
	return "form_builder_forms"
}
