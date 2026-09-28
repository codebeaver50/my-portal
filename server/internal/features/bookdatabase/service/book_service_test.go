package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

func TestBookService_List_EmptyDB(t *testing.T) {
	db := newTestDB(t)
	svc := NewBookService(repository.NewBookRepository(db))

	got, err := svc.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got.Books) != 0 {
		t.Errorf("len(Books) = %d, want 0", len(got.Books))
	}
	if got.HasMore {
		t.Errorf("HasMore = true, want false")
	}
}

func TestBookService_List_OrdersByPublishedDateDescAndPaginates(t *testing.T) {
	db := newTestDB(t)
	svc := NewBookService(repository.NewBookRepository(db))

	publisher := &models.Publisher{Name: "テスト出版"}
	if err := db.Create(publisher).Error; err != nil {
		t.Fatalf("failed to seed publisher: %v", err)
	}
	author := &models.Author{Name: "山田太郎"}
	if err := db.Create(author).Error; err != nil {
		t.Fatalf("failed to seed author: %v", err)
	}

	const total = PageSize + 1
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	wantIDs := make([]int64, 0, total)
	for i := 0; i < total; i++ {
		// i=0 が最新の発行日になるよう降順で日付をずらしていく。
		date := base.AddDate(0, 0, -i)
		book := &models.Book{
			Title:         fmt.Sprintf("book-%02d", i),
			PublishedDate: &date,
			PublisherID:   &publisher.ID,
		}
		if err := db.Create(book).Error; err != nil {
			t.Fatalf("failed to seed book %d: %v", i, err)
		}
		if err := db.Create(&models.BookAuthor{BookID: book.ID, AuthorID: author.ID, Order: 0}).Error; err != nil {
			t.Fatalf("failed to link author for book %d: %v", i, err)
		}
		wantIDs = append(wantIDs, book.ID)
	}

	page1, err := svc.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("List(0) error = %v", err)
	}
	if !page1.HasMore {
		t.Errorf("page1.HasMore = false, want true")
	}
	if len(page1.Books) != PageSize {
		t.Fatalf("len(page1.Books) = %d, want %d", len(page1.Books), PageSize)
	}
	for i, b := range page1.Books {
		if b.ID != wantIDs[i] {
			t.Errorf("page1.Books[%d].ID = %d, want %d", i, b.ID, wantIDs[i])
		}
		if len(b.Authors) != 1 || b.Authors[0] != "山田太郎" {
			t.Errorf("page1.Books[%d].Authors = %v", i, b.Authors)
		}
		if b.PublisherName == nil || *b.PublisherName != "テスト出版" {
			t.Errorf("page1.Books[%d].PublisherName = %v", i, b.PublisherName)
		}
	}

	page2, err := svc.List(context.Background(), PageSize)
	if err != nil {
		t.Fatalf("List(%d) error = %v", PageSize, err)
	}
	if page2.HasMore {
		t.Errorf("page2.HasMore = true, want false")
	}
	if len(page2.Books) != 1 {
		t.Fatalf("len(page2.Books) = %d, want 1", len(page2.Books))
	}
	if page2.Books[0].ID != wantIDs[PageSize] {
		t.Errorf("page2.Books[0].ID = %d, want %d", page2.Books[0].ID, wantIDs[PageSize])
	}
}

func TestBookService_GetCoverImageURL(t *testing.T) {
	db := newTestDB(t)
	svc := NewBookService(repository.NewBookRepository(db))

	t.Run("book not found", func(t *testing.T) {
		got, err := svc.GetCoverImageURL(context.Background(), 9999)
		if err != nil {
			t.Fatalf("GetCoverImageURL() error = %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	noCover := &models.Book{Title: "カバーなし"}
	if err := db.Create(noCover).Error; err != nil {
		t.Fatalf("failed to seed book: %v", err)
	}
	t.Run("book without cover", func(t *testing.T) {
		got, err := svc.GetCoverImageURL(context.Background(), noCover.ID)
		if err != nil {
			t.Fatalf("GetCoverImageURL() error = %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	coverURL := "https://example.com/cover.jpg"
	withCover := &models.Book{Title: "カバーあり", CoverImageURL: &coverURL}
	if err := db.Create(withCover).Error; err != nil {
		t.Fatalf("failed to seed book: %v", err)
	}
	t.Run("book with cover", func(t *testing.T) {
		got, err := svc.GetCoverImageURL(context.Background(), withCover.ID)
		if err != nil {
			t.Fatalf("GetCoverImageURL() error = %v", err)
		}
		if got != coverURL {
			t.Errorf("got %q, want %q", got, coverURL)
		}
	})
}

func TestBookService_FetchCoverImage_SetsRefererFromSourceOrigin(t *testing.T) {
	var gotReferer string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer = r.Header.Get("Referer")
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-image-bytes"))
	}))
	defer ts.Close()

	svc := NewBookService(nil)
	resp, err := svc.FetchCoverImage(ts.URL + "/cover.jpg")
	if err != nil {
		t.Fatalf("FetchCoverImage() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	wantReferer := ts.URL + "/"
	if gotReferer != wantReferer {
		t.Errorf("Referer = %q, want %q", gotReferer, wantReferer)
	}
}
