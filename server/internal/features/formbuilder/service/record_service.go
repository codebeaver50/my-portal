package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/mail"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

const (
	maxTextLength     = 255
	maxTextareaLength = 2000
	maxEmailLength    = 254
	maxNumberAbs      = 1e15
)

// RecordService はフォームへの回答の投稿・削除・一覧取得と、項目定義に基づく回答内容の検証を実装する。
type RecordService interface {
	Create(ctx context.Context, formID uint, req *dto.RecordRequest) (*dto.RecordResponse, error)
	Delete(ctx context.Context, formID, recordID uint) error
	ListPaginated(ctx context.Context, formID uint, page, pageSize int) (*dto.PaginatedRecordsResponse, error)
}

type recordService struct {
	formRepo   *repository.FormRepository
	recordRepo *repository.FormRecordRepository
}

// NewRecordService は指定のrepositoryを使う RecordService を作成する。
func NewRecordService(formRepo *repository.FormRepository, recordRepo *repository.FormRecordRepository) RecordService {
	return &recordService{formRepo: formRepo, recordRepo: recordRepo}
}

// validateValues は回答値をフォームの項目定義に照らして検証し、保存用に正規化した
// 値を返す。未入力の任意項目は結果に含めない。
func validateValues(fields []models.FormField, values map[string]any) (map[string]any, error) {
	known := make(map[string]bool, len(fields))
	for _, field := range fields {
		known[strconv.FormatUint(uint64(field.ID), 10)] = true
	}
	for key := range values {
		if !known[key] {
			return nil, &ValidationError{Message: fmt.Sprintf("フォームに存在しない項目（ID: %s）が含まれています", key)}
		}
	}

	normalized := make(map[string]any, len(fields))
	fieldErrors := make(map[string]string)
	for _, field := range fields {
		key := strconv.FormatUint(uint64(field.ID), 10)

		value, err := normalizeValue(field, values[key])
		if err != nil {
			fieldErrors[key] = err.Error()
			continue
		}
		if value == nil {
			if field.Required {
				fieldErrors[key] = "この項目は必須です"
			}
			continue
		}
		normalized[key] = value
	}

	if len(fieldErrors) > 0 {
		return nil, &ValidationError{Message: "入力内容に誤りがあります", FieldErrors: fieldErrors}
	}
	return normalized, nil
}

// normalizeValue は1項目分の回答値を入力形式に応じて検証・正規化する。
// 未入力（nil・空文字・空配列）の場合は (nil, nil) を返す。
func normalizeValue(field models.FormField, raw any) (any, error) {
	if raw == nil {
		return nil, nil
	}

	var options []string
	if field.Type.HasOptions() {
		if err := json.Unmarshal([]byte(field.Options), &options); err != nil {
			return nil, fmt.Errorf("選択肢の読み込みに失敗しました")
		}
	}

	switch field.Type {
	case models.NumberField:
		number, ok := raw.(float64)
		if !ok {
			return nil, fmt.Errorf("数値を入力してください")
		}
		if math.IsNaN(number) || math.Abs(number) > maxNumberAbs {
			return nil, fmt.Errorf("入力できる範囲を超えています")
		}
		return number, nil

	case models.CheckboxField:
		items, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("選択肢の形式が不正です")
		}
		selected := make([]string, 0, len(items))
		for _, item := range items {
			option, ok := item.(string)
			if !ok || !slices.Contains(options, option) {
				return nil, fmt.Errorf("選択肢にない値が含まれています")
			}
			if !slices.Contains(selected, option) {
				selected = append(selected, option)
			}
		}
		if len(selected) == 0 {
			return nil, nil
		}
		// 回答の並び順に依らず、選択肢の定義順で保存する。
		slices.SortFunc(selected, func(a, b string) int {
			return slices.Index(options, a) - slices.Index(options, b)
		})
		return selected, nil
	}

	text, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("文字列を入力してください")
	}
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	switch field.Type {
	case models.TextField:
		text = strings.TrimSpace(text)
		if utf8.RuneCountInString(text) > maxTextLength {
			return nil, fmt.Errorf("%d文字以内で入力してください", maxTextLength)
		}
	case models.TextareaField:
		if utf8.RuneCountInString(text) > maxTextareaLength {
			return nil, fmt.Errorf("%d文字以内で入力してください", maxTextareaLength)
		}
	case models.EmailField:
		text = strings.TrimSpace(text)
		address, err := mail.ParseAddress(text)
		if err != nil || address.Address != text || len(text) > maxEmailLength {
			return nil, fmt.Errorf("メールアドレスの形式が正しくありません")
		}
	case models.DateField:
		if _, err := time.Parse("2006-01-02", text); err != nil {
			return nil, fmt.Errorf("日付はYYYY-MM-DD形式で入力してください")
		}
	case models.SelectField, models.RadioField:
		if !slices.Contains(options, text) {
			return nil, fmt.Errorf("選択肢にない値です")
		}
	default:
		return nil, fmt.Errorf("未対応の入力形式です")
	}
	return text, nil
}

func (s *recordService) Create(ctx context.Context, formID uint, req *dto.RecordRequest) (*dto.RecordResponse, error) {
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}

	values, err := validateValues(form.Fields, req.Values)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	record := models.FormRecord{FormID: formID, Data: string(data)}
	if err := s.recordRepo.Create(ctx, &record); err != nil {
		return nil, err
	}

	created, err := s.recordRepo.GetByID(ctx, record.ID)
	if err != nil {
		return nil, err
	}
	if created == nil {
		return nil, ErrRecordNotFound
	}
	return toRecordResponse(*created)
}

func (s *recordService) Delete(ctx context.Context, formID, recordID uint) error {
	record, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}
	if record == nil || record.FormID != formID {
		return ErrRecordNotFound
	}
	return s.recordRepo.Delete(ctx, recordID)
}

func (s *recordService) ListPaginated(ctx context.Context, formID uint, page, pageSize int) (*dto.PaginatedRecordsResponse, error) {
	form, err := s.formRepo.GetByID(ctx, formID)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}

	records, total, err := s.recordRepo.GetPaginatedByFormID(ctx, formID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.RecordResponse, len(records))
	for i, record := range records {
		response, err := toRecordResponse(record)
		if err != nil {
			return nil, err
		}
		responses[i] = *response
	}

	return &dto.PaginatedRecordsResponse{
		Records:     responses,
		Total:       int(total),
		Page:        page,
		PageSize:    pageSize,
		HasNextPage: int64(page*pageSize) < total,
	}, nil
}

func toRecordResponse(record models.FormRecord) (*dto.RecordResponse, error) {
	values := map[string]any{}
	if err := json.Unmarshal([]byte(record.Data), &values); err != nil {
		return nil, fmt.Errorf("decode data of record %d: %w", record.ID, err)
	}
	return &dto.RecordResponse{
		ID:        record.ID,
		FormID:    record.FormID,
		Values:    values,
		CreatedAt: record.CreatedAt,
	}, nil
}
