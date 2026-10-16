package service

import (
	"regexp"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validateSlugFormat は slug がURLパスセグメントとして安全な形式
// （小文字英数字とハイフン区切り、例: "my-post"）かを検証する。
func validateSlugFormat(slug string) error {
	if !slugPattern.MatchString(slug) {
		return &ValidationError{Message: "スラッグは半角英小文字・数字とハイフンで入力してください（例: my-post）"}
	}
	return nil
}
