package service

import (
	"context"
	"errors"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// ErrCategoryNotFound は指定のIDのカテゴリが存在しない場合に返される。
var ErrCategoryNotFound = errors.New("category not found")

// ErrCategorySlugTaken は指定のslugのカテゴリが既に存在する場合に返される。
var ErrCategorySlugTaken = errors.New("category slug already exists")

// CategoryService はカテゴリの作成・更新・削除・一覧取得を実装する。
type CategoryService interface {
	Create(ctx context.Context, req *dto.CategoryRequest) (*dto.CategoryResponse, error)
	Update(ctx context.Context, id uint, req *dto.CategoryRequest) (*dto.CategoryResponse, error)
	Delete(ctx context.Context, id uint) error
	ListAll(ctx context.Context) (*dto.ListCategoriesResponse, error)
}

type categoryService struct {
	repo *repository.CategoryRepository
}

// NewCategoryService は指定のrepositoryを使う CategoryService を作成する。
func NewCategoryService(repo *repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) checkSlugAvailable(ctx context.Context, slug string, excludeID uint) error {
	if err := validateSlugFormat(slug); err != nil {
		return err
	}

	existing, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return ErrCategorySlugTaken
	}
	return nil
}

func (s *categoryService) Create(ctx context.Context, req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	if err := s.checkSlugAvailable(ctx, req.Slug, 0); err != nil {
		return nil, err
	}

	category := models.Category{Name: req.Name, Slug: req.Slug}
	if err := s.repo.Create(ctx, &category); err != nil {
		return nil, err
	}

	response := dto.ToCategoryResponse(category)
	return &response, nil
}

func (s *categoryService) Update(ctx context.Context, id uint, req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCategoryNotFound
	}

	if err := s.checkSlugAvailable(ctx, req.Slug, id); err != nil {
		return nil, err
	}

	existing.Name = req.Name
	existing.Slug = req.Slug
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	response := dto.ToCategoryResponse(*existing)
	return &response, nil
}

func (s *categoryService) Delete(ctx context.Context, id uint) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrCategoryNotFound
	}
	// 記事が紐づいている場合はFKのON DELETE RESTRICTによりDBエラーが返る。
	return s.repo.Delete(ctx, id)
}

func (s *categoryService) ListAll(ctx context.Context) (*dto.ListCategoriesResponse, error) {
	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = dto.ToCategoryResponse(category)
	}

	return &dto.ListCategoriesResponse{Categories: responses}, nil
}
