package models

import "time"

// FieldType はフォーム項目の入力形式を表す。
type FieldType string

const (
	TextField     FieldType = "text"
	TextareaField FieldType = "textarea"
	NumberField   FieldType = "number"
	EmailField    FieldType = "email"
	DateField     FieldType = "date"
	SelectField   FieldType = "select"
	RadioField    FieldType = "radio"
	CheckboxField FieldType = "checkbox"
)

// HasOptions は選択肢（options）を持つ入力形式かどうかを返す。
func (t FieldType) HasOptions() bool {
	return t == SelectField || t == RadioField || t == CheckboxField
}

// FormField: フォームの項目定義
type FormField struct {
	ID       uint      `gorm:"column:id;primaryKey" json:"id"`
	FormID   uint      `gorm:"column:form_id;not null;index" json:"formId"`
	Label    string    `gorm:"column:label;size:100;not null" json:"label"`
	Type     FieldType `gorm:"column:type;size:20;not null" json:"type"`
	Required bool      `gorm:"column:required;not null;default:false" json:"required"`
	// Options: select/radio/checkbox の選択肢（文字列のJSON配列）。それ以外の形式では "[]"
	Options string `gorm:"column:options;type:json;not null" json:"options"`
	// LayoutRow: 表示する行（0始まり）。同じ行の項目は横並びで表示する
	LayoutRow int `gorm:"column:layout_row;not null;default:0" json:"layoutRow"`
	// SortOrder: フォーム全体での表示順（行ごとに左から右）
	SortOrder int       `gorm:"column:sort_order;not null;default:0" json:"sortOrder"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// FormField 構造体は form_builder_form_fields テーブルにマッピングされる
func (FormField) TableName() string {
	return "form_builder_form_fields"
}
