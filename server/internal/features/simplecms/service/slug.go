package service

import (
	"errors"
	"regexp"
)

// ErrInvalidSlug は slug がURLパスセグメントとして安全な形式
// （小文字英数字とハイフン区切り、例: "my-post"）でない場合に返される。
var ErrInvalidSlug = errors.New(`slug must be lowercase alphanumeric characters and hyphens (e.g. "my-post")`)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateSlugFormat(slug string) error {
	if !slugPattern.MatchString(slug) {
		return ErrInvalidSlug
	}
	return nil
}
