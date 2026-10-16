package service

import (
	"context"
	"errors"
	"testing"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/dto"
)

func assertValidationError(t *testing.T, err error) {
	t.Helper()
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
}

func TestCategoryService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects an invalid slug with a ValidationError", func(t *testing.T) {
		categorySvc, _ := newServices(t)

		_, err := categorySvc.Create(ctx, &dto.CategoryRequest{Name: "お知らせ", Slug: "News"})
		assertValidationError(t, err)
	})

	t.Run("returns ErrCategorySlugTaken for a duplicate slug", func(t *testing.T) {
		categorySvc, _ := newServices(t)
		if _, err := categorySvc.Create(ctx, &dto.CategoryRequest{Name: "お知らせ", Slug: "news"}); err != nil {
			t.Fatalf("Create() err = %v", err)
		}

		_, err := categorySvc.Create(ctx, &dto.CategoryRequest{Name: "ニュース", Slug: "news"})
		if !errors.Is(err, ErrCategorySlugTaken) {
			t.Errorf("err = %v, want ErrCategorySlugTaken", err)
		}
	})
}

func TestCategoryService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ErrCategoryInUse when posts belong to the category", func(t *testing.T) {
		categorySvc, postSvc := newServices(t)
		category, err := categorySvc.Create(ctx, &dto.CategoryRequest{Name: "お知らせ", Slug: "news"})
		if err != nil {
			t.Fatalf("Create() err = %v", err)
		}
		if _, err := postSvc.Create(ctx, &dto.PostRequest{CategoryID: category.ID, Title: "公開", Slug: "hello", Content: "本文"}); err != nil {
			t.Fatalf("Create() post err = %v", err)
		}

		if err := categorySvc.Delete(ctx, category.ID); !errors.Is(err, ErrCategoryInUse) {
			t.Errorf("err = %v, want ErrCategoryInUse", err)
		}
	})

	t.Run("deletes a category without posts", func(t *testing.T) {
		categorySvc, _ := newServices(t)
		category, err := categorySvc.Create(ctx, &dto.CategoryRequest{Name: "お知らせ", Slug: "news"})
		if err != nil {
			t.Fatalf("Create() err = %v", err)
		}

		if err := categorySvc.Delete(ctx, category.ID); err != nil {
			t.Errorf("err = %v, want nil", err)
		}
	})
}
