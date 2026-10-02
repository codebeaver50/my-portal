package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// fakeRecordService は service.RecordService をDBアクセスなしで差し替える。
type fakeRecordService struct {
	createFormID uint
	createResp   *dto.RecordResponse
	createErr    error

	deleteFormID   uint
	deleteRecordID uint
	deleteErr      error

	listPage     int
	listPageSize int
	listResp     *dto.PaginatedRecordsResponse
	listErr      error
}

func (f *fakeRecordService) Create(ctx context.Context, formID uint, req *dto.RecordRequest) (*dto.RecordResponse, error) {
	f.createFormID = formID
	return f.createResp, f.createErr
}

func (f *fakeRecordService) Delete(ctx context.Context, formID, recordID uint) error {
	f.deleteFormID = formID
	f.deleteRecordID = recordID
	return f.deleteErr
}

func (f *fakeRecordService) ListPaginated(ctx context.Context, formID uint, page, pageSize int) (*dto.PaginatedRecordsResponse, error) {
	f.listPage = page
	f.listPageSize = pageSize
	return f.listResp, f.listErr
}

func newRecordRouter(svc service.RecordService) *gin.Engine {
	ctrl := NewRecordController(svc)
	engine := gin.New()
	engine.GET("/forms/:id/records", ctrl.List())
	engine.POST("/forms/:id/records", ctrl.Create())
	engine.DELETE("/forms/:id/records/:recordId", ctrl.Delete())
	return engine
}

func TestRecordController_Create(t *testing.T) {
	t.Run("returns 201 with the created record", func(t *testing.T) {
		svc := &fakeRecordService{createResp: &dto.RecordResponse{ID: 1, FormID: 5}}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/forms/5/records", bytes.NewBufferString(`{"values":{"1":"山田"}}`))
		newRecordRouter(svc).ServeHTTP(w, req)

		if w.Code != http.StatusCreated || svc.createFormID != 5 {
			t.Errorf("code = %d, formID = %d", w.Code, svc.createFormID)
		}
	})

	t.Run("returns 400 with fieldErrors on validation error", func(t *testing.T) {
		svc := &fakeRecordService{createErr: &service.ValidationError{
			Message:     "入力内容に誤りがあります",
			FieldErrors: map[string]string{"1": "この項目は必須です"},
		}}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/forms/5/records", bytes.NewBufferString(`{"values":{}}`))
		newRecordRouter(svc).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("code = %d, want 400", w.Code)
		}
		var body struct {
			Error       string            `json:"error"`
			FieldErrors map[string]string `json:"fieldErrors"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Error == "" || body.FieldErrors["1"] != "この項目は必須です" {
			t.Errorf("body = %+v", body)
		}
	})

	t.Run("returns 404 when the form does not exist", func(t *testing.T) {
		svc := &fakeRecordService{createErr: service.ErrFormNotFound}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/forms/5/records", bytes.NewBufferString(`{"values":{}}`))
		newRecordRouter(svc).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("code = %d, want 404", w.Code)
		}
	})

	t.Run("returns 400 when values is missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/forms/5/records", bytes.NewBufferString(`{}`))
		newRecordRouter(&fakeRecordService{}).ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("code = %d, want 400", w.Code)
		}
	})
}

func TestRecordController_List(t *testing.T) {
	t.Run("passes pagination params", func(t *testing.T) {
		svc := &fakeRecordService{listResp: &dto.PaginatedRecordsResponse{}}
		w := httptest.NewRecorder()
		newRecordRouter(svc).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/forms/5/records?page=2&pageSize=50", nil))

		if w.Code != http.StatusOK || svc.listPage != 2 || svc.listPageSize != 50 {
			t.Errorf("code = %d, page = %d, pageSize = %d", w.Code, svc.listPage, svc.listPageSize)
		}
	})

	t.Run("rejects too large pageSize", func(t *testing.T) {
		w := httptest.NewRecorder()
		newRecordRouter(&fakeRecordService{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/forms/5/records?pageSize=101", nil))

		if w.Code != http.StatusBadRequest {
			t.Errorf("code = %d, want 400", w.Code)
		}
	})
}

func TestRecordController_Delete(t *testing.T) {
	svc := &fakeRecordService{deleteErr: service.ErrRecordNotFound}
	w := httptest.NewRecorder()
	newRecordRouter(svc).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/forms/5/records/9", nil))

	if w.Code != http.StatusNotFound || svc.deleteFormID != 5 || svc.deleteRecordID != 9 {
		t.Errorf("code = %d, formID = %d, recordID = %d", w.Code, svc.deleteFormID, svc.deleteRecordID)
	}
}
