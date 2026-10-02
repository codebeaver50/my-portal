// Package dto は form builder feature のHTTP APIにおけるリクエスト/レスポンス型を定義する。
// service層がこれを組み立ててcontrollerに返す。
package dto

import (
	"time"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// FormFieldRequest はフォーム項目1件分の作成/更新リクエスト。
// ID が指定された場合は既存項目の更新、未指定の場合は新規項目として扱う。
// 既存項目のIDを保つことで、回答データ（項目IDをキーに保存）との対応を維持する。
type FormFieldRequest struct {
	ID       *uint            `json:"id"`
	Label    string           `json:"label" binding:"required,max=100"`
	Type     models.FieldType `json:"type" binding:"required,oneof=text textarea number email date select radio checkbox"`
	Required bool             `json:"required"`
	Options  []string         `json:"options"`
}

// FormRequest はフォームの作成/更新リクエスト。Fields の並び順がそのまま表示順になる。
type FormRequest struct {
	Title       string             `json:"title" binding:"required,max=100"`
	Description string             `json:"description" binding:"max=500"`
	Fields      []FormFieldRequest `json:"fields" binding:"required,min=1,max=50,dive"`
}

// FormFieldResponse はフォーム項目1件分。
type FormFieldResponse struct {
	ID       uint             `json:"id"`
	Label    string           `json:"label"`
	Type     models.FieldType `json:"type"`
	Required bool             `json:"required"`
	Options  []string         `json:"options"`
}

// FormResponse はフォーム1件分（項目定義つき）。
type FormResponse struct {
	ID          uint                `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Fields      []FormFieldResponse `json:"fields"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

// FormSummaryResponse はフォーム一覧の1件分。
type FormSummaryResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FieldCount  int       `json:"fieldCount"`
	RecordCount int64     `json:"recordCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ListFormsResponse は GET /api/form-builder/forms のレスポンス。
type ListFormsResponse struct {
	Forms []FormSummaryResponse `json:"forms"`
}

// RecordRequest は回答の投稿リクエスト。Values のキーは項目IDの文字列。
type RecordRequest struct {
	Values map[string]any `json:"values" binding:"required"`
}

// RecordResponse は回答1件分。Values のキーは項目IDの文字列。
type RecordResponse struct {
	ID        uint           `json:"id"`
	FormID    uint           `json:"formId"`
	Values    map[string]any `json:"values"`
	CreatedAt time.Time      `json:"createdAt"`
}

// PaginatedRecordsResponse は GET /api/form-builder/forms/:id/records のレスポンス。
type PaginatedRecordsResponse struct {
	Records     []RecordResponse `json:"records"`
	Total       int              `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	HasNextPage bool             `json:"hasNextPage"`
}
