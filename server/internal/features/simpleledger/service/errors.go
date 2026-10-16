package service

// ValidationError はリクエストの内容が不正な場合に返される。Message は画面にそのまま出せる日本語にする。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
