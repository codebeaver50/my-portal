package service

import (
	"context"
	"errors"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// ErrPostNotFound は指定のIDの記事が存在しない場合に返される。
var ErrPostNotFound = errors.New("post not found")

// ErrPostSlugTaken は指定のslugの記事が既に存在する場合に返される。
var ErrPostSlugTaken = errors.New("post slug already exists")

// PostService は記事の作成・更新・削除・一覧取得を実装する。
type PostService interface {
	Create(ctx context.Context, req *dto.PostRequest) (*dto.PostResponse, error)
	Update(ctx context.Context, id uint, req *dto.PostRequest) (*dto.PostResponse, error)
	Delete(ctx context.Context, id uint) error
	ListAll(ctx context.Context) (*dto.ListPostsResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.PostResponse, error)
}

type postService struct {
	postRepo     *repository.PostRepository
	categoryRepo *repository.CategoryRepository
}

// NewPostService は指定のrepositoryを使う PostService を作成する。
func NewPostService(postRepo *repository.PostRepository, categoryRepo *repository.CategoryRepository) PostService {
	return &postService{postRepo: postRepo, categoryRepo: categoryRepo}
}

func (s *postService) validate(ctx context.Context, req *dto.PostRequest, excludePostID uint) error {
	if err := validateSlugFormat(req.Slug); err != nil {
		return err
	}

	category, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return ErrCategoryNotFound
	}

	existing, err := s.postRepo.GetBySlug(ctx, req.Slug)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != excludePostID {
		return ErrPostSlugTaken
	}
	return nil
}

func (s *postService) Create(ctx context.Context, req *dto.PostRequest) (*dto.PostResponse, error) {
	if err := s.validate(ctx, req, 0); err != nil {
		return nil, err
	}

	post := models.Post{
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Slug:       req.Slug,
		Content:    req.Content,
	}
	if err := s.postRepo.Create(ctx, &post); err != nil {
		return nil, err
	}

	return s.getResponse(ctx, post.ID)
}

func (s *postService) Update(ctx context.Context, id uint, req *dto.PostRequest) (*dto.PostResponse, error) {
	existing, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrPostNotFound
	}

	if err := s.validate(ctx, req, id); err != nil {
		return nil, err
	}

	existing.CategoryID = req.CategoryID
	existing.Title = req.Title
	existing.Slug = req.Slug
	existing.Content = req.Content
	if err := s.postRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return s.getResponse(ctx, id)
}

func (s *postService) Delete(ctx context.Context, id uint) error {
	existing, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrPostNotFound
	}
	return s.postRepo.Delete(ctx, id)
}

func (s *postService) ListAll(ctx context.Context) (*dto.ListPostsResponse, error) {
	posts, err := s.postRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = dto.ToPostResponse(post)
	}

	return &dto.ListPostsResponse{Posts: responses}, nil
}

func (s *postService) GetBySlug(ctx context.Context, slug string) (*dto.PostResponse, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	response := dto.ToPostResponse(*post)
	return &response, nil
}

func (s *postService) getResponse(ctx context.Context, id uint) (*dto.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	response := dto.ToPostResponse(*post)
	return &response, nil
}
