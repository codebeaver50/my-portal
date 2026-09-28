package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/dto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// fakeBookService は service.BookService をネットワーク/DBアクセスなしで
// 差し替える。渡された引数を記録し、事前に設定した戻り値をそのまま返す。
type fakeBookService struct {
	listSkip int
	listResp *dto.ListBooksResponse
	listErr  error

	coverBookID int64
	coverURL    string
	coverErr    error

	fetchSourceURL string
	fetchResp      *http.Response
	fetchErr       error
}

func (f *fakeBookService) List(ctx context.Context, skip int) (*dto.ListBooksResponse, error) {
	f.listSkip = skip
	return f.listResp, f.listErr
}

func (f *fakeBookService) GetCoverImageURL(ctx context.Context, bookID int64) (string, error) {
	f.coverBookID = bookID
	return f.coverURL, f.coverErr
}

func (f *fakeBookService) FetchCoverImage(sourceURL string) (*http.Response, error) {
	f.fetchSourceURL = sourceURL
	return f.fetchResp, f.fetchErr
}

func TestBookController_List(t *testing.T) {
	t.Run("returns books from service", func(t *testing.T) {
		subtitle := "副題"
		svc := &fakeBookService{listResp: &dto.ListBooksResponse{
			Books:   []dto.BookResponse{{ID: 1, Title: "テスト本", Subtitle: &subtitle}},
			HasMore: true,
		}}
		r := gin.New()
		r.GET("/books", NewBookController(svc).List())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books?skip=5", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if svc.listSkip != 5 {
			t.Errorf("skip passed to service = %d, want 5", svc.listSkip)
		}

		var got dto.ListBooksResponse
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}
		if !got.HasMore || len(got.Books) != 1 || got.Books[0].Title != "テスト本" {
			t.Errorf("body = %+v", got)
		}
	})

	t.Run("clamps negative skip to zero", func(t *testing.T) {
		svc := &fakeBookService{listResp: &dto.ListBooksResponse{}}
		r := gin.New()
		r.GET("/books", NewBookController(svc).List())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books?skip=-3", nil))

		if svc.listSkip != 0 {
			t.Errorf("skip passed to service = %d, want 0", svc.listSkip)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &fakeBookService{listErr: errors.New("boom")}
		r := gin.New()
		r.GET("/books", NewBookController(svc).List())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books", nil))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", w.Code)
		}
	})
}

func TestBookController_Cover(t *testing.T) {
	t.Run("invalid id returns 400", func(t *testing.T) {
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(&fakeBookService{}).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/abc/cover", nil))

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("cover lookup error returns 500", func(t *testing.T) {
		svc := &fakeBookService{coverErr: errors.New("db down")}
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(svc).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/1/cover", nil))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", w.Code)
		}
	})

	t.Run("no cover url returns 404", func(t *testing.T) {
		svc := &fakeBookService{coverURL: ""}
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(svc).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/1/cover", nil))

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
		if svc.coverBookID != 1 {
			t.Errorf("bookID passed to service = %d, want 1", svc.coverBookID)
		}
	})

	t.Run("fetch error returns 502", func(t *testing.T) {
		svc := &fakeBookService{coverURL: "https://example.com/a.jpg", fetchErr: errors.New("network")}
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(svc).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/1/cover", nil))

		if w.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", w.Code)
		}
	})

	t.Run("upstream non-200 returns 502", func(t *testing.T) {
		svc := &fakeBookService{
			coverURL:  "https://example.com/a.jpg",
			fetchResp: &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))},
		}
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(svc).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/1/cover", nil))

		if w.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", w.Code)
		}
	})

	t.Run("success proxies image bytes and headers", func(t *testing.T) {
		body := "fake-image-bytes"
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"image/png"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		svc := &fakeBookService{coverURL: "https://example.com/a.jpg", fetchResp: resp}
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(svc).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/42/cover", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if svc.coverBookID != 42 {
			t.Errorf("bookID passed to service = %d, want 42", svc.coverBookID)
		}
		if got := w.Header().Get("Content-Type"); got != "image/png" {
			t.Errorf("Content-Type = %q, want image/png", got)
		}
		if got := w.Header().Get("Cache-Control"); got != "public, max-age=86400" {
			t.Errorf("Cache-Control = %q", got)
		}
		if w.Body.String() != body {
			t.Errorf("body = %q, want %q", w.Body.String(), body)
		}
	})

	t.Run("falls back to image/jpeg when upstream sends no content type", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("x"))}
		svc := &fakeBookService{coverURL: "https://example.com/a.jpg", fetchResp: resp}
		r := gin.New()
		r.GET("/books/:id/cover", NewBookController(svc).Cover())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/1/cover", nil))

		if got := w.Header().Get("Content-Type"); got != "image/jpeg" {
			t.Errorf("Content-Type = %q, want image/jpeg", got)
		}
	})
}
