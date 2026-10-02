package models

import "time"

// FormRecord: フォームへの回答1件
type FormRecord struct {
	ID     uint `gorm:"column:id;primaryKey" json:"id"`
	FormID uint `gorm:"column:form_id;not null;index" json:"formId"`
	// Data: { "<field_id>": 値 } 形式のJSONオブジェクト
	Data      string    `gorm:"column:data;type:json;not null" json:"data"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

// FormRecord 構造体は form_builder_form_records テーブルにマッピングされる
func (FormRecord) TableName() string {
	return "form_builder_form_records"
}
