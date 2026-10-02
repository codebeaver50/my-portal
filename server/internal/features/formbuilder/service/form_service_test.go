package service

import (
	"context"
	"errors"
	"testing"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

func newServices(t *testing.T) (FormService, RecordService) {
	t.Helper()
	db := newTestDB(t)
	formRepo := repository.NewFormRepository(db)
	recordRepo := repository.NewFormRecordRepository(db)
	formSvc := NewFormService(db, formRepo, repository.NewFormFieldRepository(db), recordRepo)
	recordSvc := NewRecordService(formRepo, recordRepo)
	return formSvc, recordSvc
}

func uintPtr(v uint) *uint {
	return &v
}

func sampleFormRequest() *dto.FormRequest {
	return &dto.FormRequest{
		Title:       "  参加申込  ",
		Description: "説明",
		Fields: []dto.FormFieldRequest{
			{Label: "氏名", Type: models.TextField, Required: true},
			{Label: "参加方法", Type: models.RadioField, Required: true, Options: []string{" 会場 ", "オンライン"}},
			{Label: "メモ", Type: models.TextareaField, Options: []string{"無視される"}},
		},
	}
}

func assertValidationError(t *testing.T, err error) *ValidationError {
	t.Helper()
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
	return validationErr
}

func TestFormService_Create(t *testing.T) {
	t.Run("creates a form with normalized fields in order", func(t *testing.T) {
		svc, _ := newServices(t)

		result, err := svc.Create(context.Background(), sampleFormRequest())
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Title != "参加申込" || len(result.Fields) != 3 {
			t.Fatalf("result = %+v", result)
		}
		if result.Fields[0].Label != "氏名" || result.Fields[1].Label != "参加方法" || result.Fields[2].Label != "メモ" {
			t.Errorf("fields out of order: %+v", result.Fields)
		}
		if got := result.Fields[1].Options; len(got) != 2 || got[0] != "会場" {
			t.Errorf("radio options = %v, want trimmed options", got)
		}
		if got := result.Fields[2].Options; len(got) != 0 {
			t.Errorf("textarea options = %v, want empty", got)
		}
	})

	t.Run("rejects choice fields without options", func(t *testing.T) {
		svc, _ := newServices(t)

		_, err := svc.Create(context.Background(), &dto.FormRequest{
			Title:  "フォーム",
			Fields: []dto.FormFieldRequest{{Label: "部署", Type: models.SelectField}},
		})
		assertValidationError(t, err)
	})

	t.Run("rejects duplicated options", func(t *testing.T) {
		svc, _ := newServices(t)

		_, err := svc.Create(context.Background(), &dto.FormRequest{
			Title:  "フォーム",
			Fields: []dto.FormFieldRequest{{Label: "部署", Type: models.CheckboxField, Options: []string{"A", " A"}}},
		})
		assertValidationError(t, err)
	})

	t.Run("rejects blank title", func(t *testing.T) {
		svc, _ := newServices(t)

		req := sampleFormRequest()
		req.Title = "   "
		_, err := svc.Create(context.Background(), req)
		assertValidationError(t, err)
	})

	t.Run("rejects field ids on create", func(t *testing.T) {
		svc, _ := newServices(t)

		req := sampleFormRequest()
		req.Fields[0].ID = uintPtr(1)
		_, err := svc.Create(context.Background(), req)
		assertValidationError(t, err)
	})
}

func TestFormService_Update(t *testing.T) {
	t.Run("keeps, adds, removes and reorders fields", func(t *testing.T) {
		svc, _ := newServices(t)
		ctx := context.Background()

		created, err := svc.Create(ctx, sampleFormRequest())
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		nameID := created.Fields[0].ID
		methodID := created.Fields[1].ID
		memoID := created.Fields[2].ID

		updated, err := svc.Update(ctx, created.ID, &dto.FormRequest{
			Title: "参加申込（改）",
			Fields: []dto.FormFieldRequest{
				{ID: uintPtr(methodID), Label: "参加方法", Type: models.SelectField, Required: true, Options: []string{"会場", "オンライン", "未定"}},
				{ID: uintPtr(nameID), Label: "お名前", Type: models.TextField, Required: true},
				{Label: "メールアドレス", Type: models.EmailField},
			},
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Title != "参加申込（改）" || len(updated.Fields) != 3 {
			t.Fatalf("updated = %+v", updated)
		}
		if updated.Fields[0].ID != methodID || updated.Fields[0].Type != models.SelectField || len(updated.Fields[0].Options) != 3 {
			t.Errorf("first field = %+v, want updated method field", updated.Fields[0])
		}
		if updated.Fields[1].ID != nameID || updated.Fields[1].Label != "お名前" {
			t.Errorf("second field = %+v, want renamed name field", updated.Fields[1])
		}
		if updated.Fields[2].Label != "メールアドレス" {
			t.Errorf("third field = %+v, want new email field", updated.Fields[2])
		}
		for _, field := range updated.Fields {
			if field.ID == memoID {
				t.Errorf("removed field %d still exists", memoID)
			}
		}
	})

	t.Run("rejects field ids that belong to another form", func(t *testing.T) {
		svc, _ := newServices(t)
		ctx := context.Background()

		first, _ := svc.Create(ctx, sampleFormRequest())
		second, _ := svc.Create(ctx, sampleFormRequest())

		_, err := svc.Update(ctx, second.ID, &dto.FormRequest{
			Title:  "フォーム",
			Fields: []dto.FormFieldRequest{{ID: uintPtr(first.Fields[0].ID), Label: "氏名", Type: models.TextField}},
		})
		assertValidationError(t, err)
	})

	t.Run("returns ErrFormNotFound for unknown form", func(t *testing.T) {
		svc, _ := newServices(t)

		_, err := svc.Update(context.Background(), 999, sampleFormRequest())
		if !errors.Is(err, ErrFormNotFound) {
			t.Errorf("err = %v, want ErrFormNotFound", err)
		}
	})
}

func TestFormService_ListAll(t *testing.T) {
	svc, recordSvc := newServices(t)
	ctx := context.Background()

	withRecords, _ := svc.Create(ctx, sampleFormRequest())
	if _, err := svc.Create(ctx, sampleFormRequest()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	for range 2 {
		_, err := recordSvc.Create(ctx, withRecords.ID, &dto.RecordRequest{Values: map[string]any{
			key(withRecords.Fields[0].ID): "山田",
			key(withRecords.Fields[1].ID): "会場",
		}})
		if err != nil {
			t.Fatalf("record Create() error = %v", err)
		}
	}

	result, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(result.Forms) != 2 {
		t.Fatalf("len(forms) = %d, want 2", len(result.Forms))
	}
	for _, form := range result.Forms {
		want := int64(0)
		if form.ID == withRecords.ID {
			want = 2
		}
		if form.RecordCount != want || form.FieldCount != 3 {
			t.Errorf("form %d: recordCount = %d, fieldCount = %d", form.ID, form.RecordCount, form.FieldCount)
		}
	}
}

func TestFormService_Delete(t *testing.T) {
	svc, _ := newServices(t)
	ctx := context.Background()

	created, _ := svc.Create(ctx, sampleFormRequest())
	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := svc.Get(ctx, created.ID); !errors.Is(err, ErrFormNotFound) {
		t.Errorf("Get() err = %v, want ErrFormNotFound", err)
	}
	if err := svc.Delete(ctx, created.ID); !errors.Is(err, ErrFormNotFound) {
		t.Errorf("second Delete() err = %v, want ErrFormNotFound", err)
	}
}
