package service

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

func key(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// newAllTypesForm は全入力形式の項目を1つずつ持つフォームを作成し、
// 入力形式から項目IDのキーを引けるmapと一緒に返す。
func newAllTypesForm(t *testing.T) (FormService, RecordService, uint, map[models.FieldType]string) {
	t.Helper()
	formSvc, recordSvc := newServices(t)

	created, err := formSvc.Create(context.Background(), &dto.FormRequest{
		Title: "全項目",
		Fields: []dto.FormFieldRequest{
			{Label: "テキスト", Type: models.TextField, Required: true},
			{Label: "複数行", Type: models.TextareaField},
			{Label: "数値", Type: models.NumberField},
			{Label: "メール", Type: models.EmailField},
			{Label: "日付", Type: models.DateField},
			{Label: "セレクト", Type: models.SelectField, Options: []string{"A", "B"}},
			{Label: "ラジオ", Type: models.RadioField, Options: []string{"はい", "いいえ"}},
			{Label: "チェック", Type: models.CheckboxField, Options: []string{"X", "Y", "Z"}},
		},
	})
	if err != nil {
		t.Fatalf("form Create() error = %v", err)
	}

	keys := make(map[models.FieldType]string, len(created.Fields))
	for _, field := range created.Fields {
		keys[field.Type] = key(field.ID)
	}
	return formSvc, recordSvc, created.ID, keys
}

func TestRecordService_Create(t *testing.T) {
	t.Run("stores normalized values and omits empty optional values", func(t *testing.T) {
		_, svc, formID, keys := newAllTypesForm(t)

		result, err := svc.Create(context.Background(), formID, &dto.RecordRequest{Values: map[string]any{
			keys[models.TextField]:     "  山田  ",
			keys[models.TextareaField]: "",
			keys[models.NumberField]:   float64(3),
			keys[models.EmailField]:    "taro@example.com",
			keys[models.DateField]:     "2026-10-02",
			keys[models.SelectField]:   "B",
			keys[models.RadioField]:    nil,
			keys[models.CheckboxField]: []any{"Z", "X", "Z"},
		}})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		values := result.Values
		if values[keys[models.TextField]] != "山田" {
			t.Errorf("text = %v, want trimmed", values[keys[models.TextField]])
		}
		if values[keys[models.NumberField]] != float64(3) {
			t.Errorf("number = %v", values[keys[models.NumberField]])
		}
		if _, ok := values[keys[models.TextareaField]]; ok {
			t.Errorf("empty textarea should be omitted: %v", values)
		}
		if _, ok := values[keys[models.RadioField]]; ok {
			t.Errorf("nil radio should be omitted: %v", values)
		}
		checked, ok := values[keys[models.CheckboxField]].([]any)
		if !ok || len(checked) != 2 || checked[0] != "X" || checked[1] != "Z" {
			t.Errorf("checkbox = %v, want [X Z] in option order", values[keys[models.CheckboxField]])
		}
	})

	t.Run("reports field errors per field", func(t *testing.T) {
		_, svc, formID, keys := newAllTypesForm(t)

		_, err := svc.Create(context.Background(), formID, &dto.RecordRequest{Values: map[string]any{
			keys[models.TextField]:     "   ",
			keys[models.NumberField]:   "3",
			keys[models.EmailField]:    "Taro <taro@example.com>",
			keys[models.DateField]:     "2026/10/02",
			keys[models.SelectField]:   "C",
			keys[models.RadioField]:    "たぶん",
			keys[models.CheckboxField]: []any{"X", "W"},
		}})
		validationErr := assertValidationError(t, err)

		for _, fieldType := range []models.FieldType{
			models.TextField, models.NumberField, models.EmailField, models.DateField,
			models.SelectField, models.RadioField, models.CheckboxField,
		} {
			if _, ok := validationErr.FieldErrors[keys[fieldType]]; !ok {
				t.Errorf("missing field error for %s: %v", fieldType, validationErr.FieldErrors)
			}
		}
		if _, ok := validationErr.FieldErrors[keys[models.TextareaField]]; ok {
			t.Errorf("unexpected field error for optional textarea")
		}
	})

	t.Run("rejects unknown field keys", func(t *testing.T) {
		_, svc, formID, keys := newAllTypesForm(t)

		_, err := svc.Create(context.Background(), formID, &dto.RecordRequest{Values: map[string]any{
			keys[models.TextField]: "山田",
			"99999":                "?",
		}})
		assertValidationError(t, err)
	})

	t.Run("returns ErrFormNotFound for unknown form", func(t *testing.T) {
		_, svc := newServices(t)

		_, err := svc.Create(context.Background(), 999, &dto.RecordRequest{Values: map[string]any{}})
		if !errors.Is(err, ErrFormNotFound) {
			t.Errorf("err = %v, want ErrFormNotFound", err)
		}
	})
}

func TestRecordService_ListPaginated(t *testing.T) {
	_, svc, formID, keys := newAllTypesForm(t)
	ctx := context.Background()

	for _, name := range []string{"1", "2", "3"} {
		if _, err := svc.Create(ctx, formID, &dto.RecordRequest{Values: map[string]any{keys[models.TextField]: name}}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	first, err := svc.ListPaginated(ctx, formID, 1, 2)
	if err != nil {
		t.Fatalf("ListPaginated() error = %v", err)
	}
	if first.Total != 3 || len(first.Records) != 2 || !first.HasNextPage {
		t.Errorf("first page = %+v", first)
	}
	if first.Records[0].Values[keys[models.TextField]] != "3" {
		t.Errorf("first record = %v, want newest first", first.Records[0].Values)
	}

	second, err := svc.ListPaginated(ctx, formID, 2, 2)
	if err != nil {
		t.Fatalf("ListPaginated() error = %v", err)
	}
	if len(second.Records) != 1 || second.HasNextPage {
		t.Errorf("second page = %+v", second)
	}

	if _, err := svc.ListPaginated(ctx, 999, 1, 20); !errors.Is(err, ErrFormNotFound) {
		t.Errorf("err = %v, want ErrFormNotFound", err)
	}
}

func TestRecordService_Delete(t *testing.T) {
	formSvc, svc, formID, keys := newAllTypesForm(t)
	ctx := context.Background()

	record, err := svc.Create(ctx, formID, &dto.RecordRequest{Values: map[string]any{keys[models.TextField]: "山田"}})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	other, _ := formSvc.Create(ctx, sampleFormRequest())
	if err := svc.Delete(ctx, other.ID, record.ID); !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("Delete() via other form err = %v, want ErrRecordNotFound", err)
	}
	if err := svc.Delete(ctx, formID, record.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := svc.Delete(ctx, formID, record.ID); !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("second Delete() err = %v, want ErrRecordNotFound", err)
	}
}
