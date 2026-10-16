package service

import (
	"context"
	"testing"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/dto"
)

func TestPostService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects an unknown category with a ValidationError", func(t *testing.T) {
		_, postSvc := newServices(t)

		_, err := postSvc.Create(ctx, &dto.PostRequest{CategoryID: 999, Title: "公開", Slug: "hello", Content: "本文"})
		assertValidationError(t, err)
	})
}
