package service

import "errors"

// ErrCategoryInUse は記事が紐づいているカテゴリを削除しようとした場合に返される。
var ErrCategoryInUse = errors.New("category has posts")

// ValidationError はリクエストの内容が不正な場合に返される。Message は画面にそのまま出せる日本語にする。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
