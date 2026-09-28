// Package dto は simple cms feature のHTTP APIにおけるリクエスト/レスポンス型を定義する。
// service層がこれを組み立ててcontrollerに返す。
package dto

import (
	"time"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// CategoryRequest はカテゴリの作成/更新リクエスト。
type CategoryRequest struct {
	Name string `json:"name" binding:"required,max=100"`
	Slug string `json:"slug" binding:"required,max=100"`
}

// CategoryResponse はカテゴリ1件分。
type CategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ListCategoriesResponse は GET /api/simple-cms/categories のレスポンス。
type ListCategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// PostRequest は記事の作成/更新リクエスト。
type PostRequest struct {
	CategoryID uint   `json:"categoryId" binding:"required"`
	Title      string `json:"title" binding:"required,max=255"`
	Slug       string `json:"slug" binding:"required,max=255"`
	Content    string `json:"content" binding:"required"`
}

// PostResponse は記事1件分。
type PostResponse struct {
	ID         uint              `json:"id"`
	CategoryID uint              `json:"categoryId"`
	Category   *CategoryResponse `json:"category,omitempty"`
	Title      string            `json:"title"`
	Slug       string            `json:"slug"`
	Content    string            `json:"content"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

// ListPostsResponse は GET /api/simple-cms/posts のレスポンス。
type ListPostsResponse struct {
	Posts []PostResponse `json:"posts"`
}

// ToCategoryResponse は Category モデルを CategoryResponse に変換する。
func ToCategoryResponse(category models.Category) CategoryResponse {
	return CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}
}

// ToPostResponse は Post モデルを PostResponse に変換する。
func ToPostResponse(post models.Post) PostResponse {
	response := PostResponse{
		ID:         post.ID,
		CategoryID: post.CategoryID,
		Title:      post.Title,
		Slug:       post.Slug,
		Content:    post.Content,
		CreatedAt:  post.CreatedAt,
		UpdatedAt:  post.UpdatedAt,
	}
	if post.Category != nil {
		category := ToCategoryResponse(*post.Category)
		response.Category = &category
	}
	return response
}
