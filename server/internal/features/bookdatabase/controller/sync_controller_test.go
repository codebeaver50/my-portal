package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/dto"
)

// fakeSyncService は service.SyncService をネットワーク/DBアクセスなしで差し替える。
type fakeSyncService struct {
	resp *dto.SyncSummaryResponse
	err  error
}

func (f *fakeSyncService) Sync(ctx context.Context) (*dto.SyncSummaryResponse, error) {
	return f.resp, f.err
}

func TestSyncController_Sync(t *testing.T) {
	t.Run("no CRON_SECRET set allows the request and returns the summary", func(t *testing.T) {
		svc := &fakeSyncService{resp: &dto.SyncSummaryResponse{Created: 3}}
		r := gin.New()
		r.GET("/books/sync", CronAuthMiddleware(), NewSyncController(svc).Sync())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/sync", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var got dto.SyncSummaryResponse
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}
		if got.Created != 3 {
			t.Errorf("Created = %d, want 3", got.Created)
		}
	})

	t.Run("CRON_SECRET set rejects a request without a matching bearer token", func(t *testing.T) {
		t.Setenv("CRON_SECRET", "s3cret")
		r := gin.New()
		r.GET("/books/sync", CronAuthMiddleware(), NewSyncController(&fakeSyncService{}).Sync())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/sync", nil))

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("CRON_SECRET set accepts a request with a matching bearer token", func(t *testing.T) {
		t.Setenv("CRON_SECRET", "s3cret")
		svc := &fakeSyncService{resp: &dto.SyncSummaryResponse{}}
		r := gin.New()
		r.GET("/books/sync", CronAuthMiddleware(), NewSyncController(svc).Sync())

		req := httptest.NewRequest(http.MethodGet, "/books/sync", nil)
		req.Header.Set("Authorization", "Bearer s3cret")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &fakeSyncService{err: errors.New("sync failed")}
		r := gin.New()
		r.GET("/books/sync", CronAuthMiddleware(), NewSyncController(svc).Sync())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/sync", nil))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", w.Code)
		}
	})
}
