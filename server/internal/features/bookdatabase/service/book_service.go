// Package service は book database feature のビジネスロジック（一覧取得・
// カバー画像解決・NDL/openBDのsyncバッチ）を実装する。
package service

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// PageSize は元のNext.js実装のページサイズに合わせている。
const PageSize = 20

// BookService は GET /api/books とそのカバー画像プロキシを実装する。
type BookService interface {
	List(ctx context.Context, skip int) (*dto.ListBooksResponse, error)
	GetCoverImageURL(ctx context.Context, bookID int64) (string, error)
	FetchCoverImage(sourceURL string) (*http.Response, error)
}

type bookService struct {
	bookRepo *repository.BookRepository
}

// NewBookService は指定のrepositoryを使う BookService を作成する。
func NewBookService(bookRepo *repository.BookRepository) BookService {
	return &bookService{bookRepo: bookRepo}
}

// List はskipから最大PageSize件の書籍を返し、それ以降にまだ行があるかを
// 報告する。
func (s *bookService) List(ctx context.Context, skip int) (*dto.ListBooksResponse, error) {
	books, err := s.bookRepo.List(ctx, skip, PageSize+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(books) > PageSize
	if hasMore {
		books = books[:PageSize]
	}

	responses := make([]dto.BookResponse, len(books))
	for i, book := range books {
		responses[i] = toBookResponse(book)
	}

	return &dto.ListBooksResponse{Books: responses, HasMore: hasMore}, nil
}

func toBookResponse(book models.Book) dto.BookResponse {
	authors := make([]string, 0, len(book.BookAuthors))
	for _, ba := range book.BookAuthors {
		if ba.Author != nil {
			authors = append(authors, ba.Author.Name)
		}
	}

	var publisherName *string
	if book.Publisher != nil {
		publisherName = &book.Publisher.Name
	}

	return dto.BookResponse{
		ID:            book.ID,
		Title:         book.Title,
		Subtitle:      book.Subtitle,
		Authors:       authors,
		PublisherName: publisherName,
		PublishedDate: book.PublishedDate,
		CoverImageURL: book.CoverImageURL,
	}
}

// GetCoverImageURL は書籍に保存されているカバー画像URLを返す。書籍が存在
// しない、またはカバー画像がない場合は "" を返す。
func (s *bookService) GetCoverImageURL(ctx context.Context, bookID int64) (string, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return "", err
	}
	if book == nil || book.CoverImageURL == nil {
		return "", nil
	}
	return *book.CoverImageURL, nil
}

// coverThrottle は取得元のカバー画像サーバー（バーストで502を返すNDLの
// サムネイルサーバー等）への送信リクエストを、同時プロキシリクエスト数に
// かかわらず最低minFetchInterval間隔になるよう直列化する。
type coverThrottle struct {
	mu          sync.Mutex
	lastFetchAt time.Time
}

const minFetchInterval = 500 * time.Millisecond

var defaultCoverThrottle = &coverThrottle{}

// FetchCoverImage はカバー画像を取得元URLからフェッチする。共有の
// デフォルトレート制限でスロットリングする。NDLのサムネイルサーバーは
// Refererが自ドメインでないリクエストを拒否するため、ここで設定する。
func (s *bookService) FetchCoverImage(sourceURL string) (*http.Response, error) {
	return defaultCoverThrottle.get(sourceURL)
}

func (t *coverThrottle) get(sourceURL string) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if wait := minFetchInterval - time.Since(t.lastFetchAt); wait > 0 {
		time.Sleep(wait)
	}
	t.lastFetchAt = time.Now()

	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	if origin, err := url.Parse(sourceURL); err == nil {
		req.Header.Set("Referer", origin.Scheme+"://"+origin.Host+"/")
	}
	return http.DefaultClient.Do(req)
}
