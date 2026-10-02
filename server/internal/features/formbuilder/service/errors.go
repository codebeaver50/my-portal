package service

import "errors"

// ErrFormNotFound は指定のIDのフォームが存在しない場合に返される。
var ErrFormNotFound = errors.New("form not found")

// ErrRecordNotFound は指定のIDの回答が存在しない（または別フォームの回答である）場合に返される。
var ErrRecordNotFound = errors.New("record not found")

// ValidationError はフォーム定義・回答内容が不正な場合に返される。FieldErrors は
// 回答の検証時のみ設定され、キーは項目IDの文字列、値はその項目のエラーメッセージ。
type ValidationError struct {
	Message     string
	FieldErrors map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}
