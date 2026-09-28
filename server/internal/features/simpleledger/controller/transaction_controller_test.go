package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// fakeTransactionService は service.TransactionService をDBアクセスなしで
// 差し替える。渡された引数を記録し、事前に設定した戻り値をそのまま返す。
type fakeTransactionService struct {
	createReq  *dto.TransactionRequest
	createResp *dto.TransactionResponse
	createErr  error

	updateID   uint
	updateReq  *dto.TransactionRequest
	updateResp *dto.TransactionResponse
	updateErr  error

	deleteID  uint
	deleteErr error

	listResp *dto.ListTransactionsResponse
	listErr  error

	paginatedPage     int
	paginatedPageSize int
	paginatedKeyword  string
	paginatedResp     *dto.PaginatedTransactionsResponse
	paginatedErr      error
}

func (f *fakeTransactionService) Create(ctx context.Context, req *dto.TransactionRequest) (*dto.TransactionResponse, error) {
	f.createReq = req
	return f.createResp, f.createErr
}

func (f *fakeTransactionService) Update(ctx context.Context, id uint, req *dto.TransactionRequest) (*dto.TransactionResponse, error) {
	f.updateID = id
	f.updateReq = req
	return f.updateResp, f.updateErr
}

func (f *fakeTransactionService) Delete(ctx context.Context, id uint) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeTransactionService) ListAll(ctx context.Context) (*dto.ListTransactionsResponse, error) {
	return f.listResp, f.listErr
}

func (f *fakeTransactionService) ListPaginated(ctx context.Context, page, pageSize int, keyword string) (*dto.PaginatedTransactionsResponse, error) {
	f.paginatedPage = page
	f.paginatedPageSize = pageSize
	f.paginatedKeyword = keyword
	return f.paginatedResp, f.paginatedErr
}

var _ service.TransactionService = (*fakeTransactionService)(nil)

func TestTransactionController_Create(t *testing.T) {
	t.Run("returns 201 with the created transaction", func(t *testing.T) {
		svc := &fakeTransactionService{createResp: &dto.TransactionResponse{ID: 1, Description: "取引"}}
		r := gin.New()
		r.POST("/transactions", NewTransactionController(svc).Create())

		body := `{"date":"2026-01-15","journalEntries":[{"chartOfAccountsId":1,"type":"debit","amount":100},{"chartOfAccountsId":2,"type":"credit","amount":100}]}`
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body)))

		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201, body = %s", w.Code, w.Body.String())
		}
		var got dto.TransactionResponse
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}
		if got.ID != 1 {
			t.Errorf("got.ID = %d, want 1", got.ID)
		}
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		svc := &fakeTransactionService{}
		r := gin.New()
		r.POST("/transactions", NewTransactionController(svc).Create())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(`{}`)))

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("service validation error returns 400", func(t *testing.T) {
		svc := &fakeTransactionService{createErr: service.ErrUnbalancedEntries}
		r := gin.New()
		r.POST("/transactions", NewTransactionController(svc).Create())

		body := `{"date":"2026-01-15","journalEntries":[{"chartOfAccountsId":1,"type":"debit","amount":100},{"chartOfAccountsId":2,"type":"credit","amount":50}]}`
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body)))

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})
}

func TestTransactionController_Update(t *testing.T) {
	t.Run("unknown id returns 404", func(t *testing.T) {
		svc := &fakeTransactionService{updateErr: service.ErrTransactionNotFound}
		r := gin.New()
		r.PUT("/transactions/:id", NewTransactionController(svc).Update())

		body := `{"date":"2026-01-15","journalEntries":[{"chartOfAccountsId":1,"type":"debit","amount":100},{"chartOfAccountsId":2,"type":"credit","amount":100}]}`
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/transactions/999", bytes.NewBufferString(body)))

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
		if svc.updateID != 999 {
			t.Errorf("id passed to service = %d, want 999", svc.updateID)
		}
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		svc := &fakeTransactionService{}
		r := gin.New()
		r.PUT("/transactions/:id", NewTransactionController(svc).Update())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/transactions/abc", bytes.NewBufferString(`{}`)))

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})
}

func TestTransactionController_Delete(t *testing.T) {
	t.Run("returns 204 on success", func(t *testing.T) {
		svc := &fakeTransactionService{}
		r := gin.New()
		r.DELETE("/transactions/:id", NewTransactionController(svc).Delete())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/transactions/7", nil))

		if w.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", w.Code)
		}
		if svc.deleteID != 7 {
			t.Errorf("id passed to service = %d, want 7", svc.deleteID)
		}
	})

	t.Run("unknown id returns 404", func(t *testing.T) {
		svc := &fakeTransactionService{deleteErr: service.ErrTransactionNotFound}
		r := gin.New()
		r.DELETE("/transactions/:id", NewTransactionController(svc).Delete())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/transactions/7", nil))

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &fakeTransactionService{deleteErr: errors.New("db down")}
		r := gin.New()
		r.DELETE("/transactions/:id", NewTransactionController(svc).Delete())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/transactions/7", nil))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", w.Code)
		}
	})
}

func TestTransactionController_ListPaginated(t *testing.T) {
	t.Run("passes query params to service", func(t *testing.T) {
		svc := &fakeTransactionService{paginatedResp: &dto.PaginatedTransactionsResponse{}}
		r := gin.New()
		r.GET("/transactions/paginated", NewTransactionController(svc).ListPaginated())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/transactions/paginated?page=2&pageSize=10&keyword=家賃", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if svc.paginatedPage != 2 || svc.paginatedPageSize != 10 || svc.paginatedKeyword != "家賃" {
			t.Errorf("page=%d pageSize=%d keyword=%q", svc.paginatedPage, svc.paginatedPageSize, svc.paginatedKeyword)
		}
	})

	t.Run("invalid pageSize returns 400", func(t *testing.T) {
		svc := &fakeTransactionService{}
		r := gin.New()
		r.GET("/transactions/paginated", NewTransactionController(svc).ListPaginated())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/transactions/paginated?pageSize=1000", nil))

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})
}
