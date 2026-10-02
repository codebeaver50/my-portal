package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

const (
	maxOptions      = 50
	maxOptionLength = 100
	maxFieldsPerRow = 3
)

// FormService はフォーム（項目定義を含む）の作成・更新・削除・取得を実装する。
type FormService interface {
	Create(ctx context.Context, req *dto.FormRequest) (*dto.FormResponse, error)
	Update(ctx context.Context, id uint, req *dto.FormRequest) (*dto.FormResponse, error)
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*dto.FormResponse, error)
	ListAll(ctx context.Context) (*dto.ListFormsResponse, error)
}

type formService struct {
	db         *gorm.DB
	formRepo   *repository.FormRepository
	fieldRepo  *repository.FormFieldRepository
	recordRepo *repository.FormRecordRepository
}

// NewFormService は指定のrepositoryを使う FormService を作成する。
func NewFormService(
	db *gorm.DB,
	formRepo *repository.FormRepository,
	fieldRepo *repository.FormFieldRepository,
	recordRepo *repository.FormRecordRepository,
) FormService {
	return &formService{db: db, formRepo: formRepo, fieldRepo: fieldRepo, recordRepo: recordRepo}
}

// normalizeFields は項目定義を検証し、前後の空白を除いたラベル・選択肢に正規化した
// コピーを返す。選択肢を持たない入力形式の options は空にする。
// 行番号は 0 から詰めた連番に振り直す（例: 0, 2, 2, 5 → 0, 1, 1, 2）。
func normalizeFields(fields []dto.FormFieldRequest) ([]dto.FormFieldRequest, error) {
	normalized := make([]dto.FormFieldRequest, len(fields))
	seenIDs := make(map[uint]bool)
	row, fieldsInRow := -1, 0

	for i, field := range fields {
		position := i + 1

		if i > 0 && field.Row < fields[i-1].Row {
			return nil, &ValidationError{Message: "項目は行の順に並べてください"}
		}
		if i == 0 || field.Row != fields[i-1].Row {
			row++
			fieldsInRow = 0
		}
		fieldsInRow++
		if fieldsInRow > maxFieldsPerRow {
			return nil, &ValidationError{Message: fmt.Sprintf("1行に並べられる項目は%d個までです", maxFieldsPerRow)}
		}

		if field.ID != nil {
			if seenIDs[*field.ID] {
				return nil, &ValidationError{Message: fmt.Sprintf("%d番目の項目のIDが重複しています", position)}
			}
			seenIDs[*field.ID] = true
		}

		label := strings.TrimSpace(field.Label)
		if label == "" {
			return nil, &ValidationError{Message: fmt.Sprintf("%d番目の項目のラベルを入力してください", position)}
		}

		options := []string{}
		if field.Type.HasOptions() {
			if len(field.Options) == 0 {
				return nil, &ValidationError{Message: fmt.Sprintf("「%s」の選択肢を1つ以上入力してください", label)}
			}
			if len(field.Options) > maxOptions {
				return nil, &ValidationError{Message: fmt.Sprintf("「%s」の選択肢は%d個以内にしてください", label, maxOptions)}
			}
			seenOptions := make(map[string]bool, len(field.Options))
			for _, option := range field.Options {
				option = strings.TrimSpace(option)
				if option == "" {
					return nil, &ValidationError{Message: fmt.Sprintf("「%s」に空の選択肢があります", label)}
				}
				if utf8.RuneCountInString(option) > maxOptionLength {
					return nil, &ValidationError{Message: fmt.Sprintf("「%s」の選択肢は%d文字以内にしてください", label, maxOptionLength)}
				}
				if seenOptions[option] {
					return nil, &ValidationError{Message: fmt.Sprintf("「%s」の選択肢「%s」が重複しています", label, option)}
				}
				seenOptions[option] = true
				options = append(options, option)
			}
		}

		normalized[i] = dto.FormFieldRequest{
			ID:       field.ID,
			Label:    label,
			Type:     field.Type,
			Required: field.Required,
			Options:  options,
			Row:      row,
		}
	}

	return normalized, nil
}

// normalizeTitle は前後の空白を除いたタイトルを返す。空白のみの場合はエラーにする。
func normalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", &ValidationError{Message: "フォーム名を入力してください"}
	}
	return title, nil
}

func toFormField(formID uint, sortOrder int, field dto.FormFieldRequest) (models.FormField, error) {
	options, err := json.Marshal(field.Options)
	if err != nil {
		return models.FormField{}, err
	}
	return models.FormField{
		FormID:    formID,
		Label:     field.Label,
		Type:      field.Type,
		Required:  field.Required,
		Options:   string(options),
		LayoutRow: field.Row,
		SortOrder: sortOrder,
	}, nil
}

func (s *formService) Create(ctx context.Context, req *dto.FormRequest) (*dto.FormResponse, error) {
	title, err := normalizeTitle(req.Title)
	if err != nil {
		return nil, err
	}
	fields, err := normalizeFields(req.Fields)
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		if field.ID != nil {
			return nil, &ValidationError{Message: "新規作成時に項目IDは指定できません"}
		}
	}

	var formID uint
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		formRepo := s.formRepo.WithTx(tx)
		fieldRepo := s.fieldRepo.WithTx(tx)

		form := models.Form{Title: title, Description: strings.TrimSpace(req.Description)}
		if err := formRepo.Create(ctx, &form); err != nil {
			return err
		}
		formID = form.ID

		newFields := make([]models.FormField, len(fields))
		for i, field := range fields {
			newField, err := toFormField(form.ID, i+1, field)
			if err != nil {
				return err
			}
			newFields[i] = newField
		}
		return fieldRepo.CreateBatch(ctx, newFields)
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, formID)
}

// Update はフォームを更新する。IDつきの項目は既存項目を上書きし、IDなしの項目は
// 新規作成し、リクエストに含まれない既存項目は削除する。削除された項目の回答値は
// 回答データ内に残るが、項目定義がないため一覧には表示されない。
func (s *formService) Update(ctx context.Context, id uint, req *dto.FormRequest) (*dto.FormResponse, error) {
	existing, err := s.formRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFormNotFound
	}

	title, err := normalizeTitle(req.Title)
	if err != nil {
		return nil, err
	}
	fields, err := normalizeFields(req.Fields)
	if err != nil {
		return nil, err
	}

	existingFields := make(map[uint]models.FormField, len(existing.Fields))
	for _, field := range existing.Fields {
		existingFields[field.ID] = field
	}
	for _, field := range fields {
		if field.ID != nil {
			if _, ok := existingFields[*field.ID]; !ok {
				return nil, &ValidationError{Message: fmt.Sprintf("項目ID %d はこのフォームに存在しません", *field.ID)}
			}
		}
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		formRepo := s.formRepo.WithTx(tx)
		fieldRepo := s.fieldRepo.WithTx(tx)

		existing.Title = title
		existing.Description = strings.TrimSpace(req.Description)
		if err := formRepo.Update(ctx, existing); err != nil {
			return err
		}

		keptIDs := make(map[uint]bool, len(fields))
		var newFields []models.FormField
		for i, field := range fields {
			next, err := toFormField(id, i+1, field)
			if err != nil {
				return err
			}
			if field.ID == nil {
				newFields = append(newFields, next)
				continue
			}

			keptIDs[*field.ID] = true
			current := existingFields[*field.ID]
			current.Label = next.Label
			current.Type = next.Type
			current.Required = next.Required
			current.Options = next.Options
			current.LayoutRow = next.LayoutRow
			current.SortOrder = next.SortOrder
			if err := fieldRepo.Update(ctx, &current); err != nil {
				return err
			}
		}

		var removedIDs []uint
		for fieldID := range existingFields {
			if !keptIDs[fieldID] {
				removedIDs = append(removedIDs, fieldID)
			}
		}
		if err := fieldRepo.DeleteByIDs(ctx, removedIDs); err != nil {
			return err
		}
		return fieldRepo.CreateBatch(ctx, newFields)
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

func (s *formService) Delete(ctx context.Context, id uint) error {
	existing, err := s.formRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrFormNotFound
	}
	// 項目定義・回答はFKのON DELETE CASCADEで一緒に削除される。
	return s.formRepo.Delete(ctx, id)
}

func (s *formService) Get(ctx context.Context, id uint) (*dto.FormResponse, error) {
	form, err := s.formRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}
	return toFormResponse(*form)
}

func (s *formService) ListAll(ctx context.Context) (*dto.ListFormsResponse, error) {
	forms, err := s.formRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := s.recordRepo.CountByFormID(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]dto.FormSummaryResponse, len(forms))
	for i, form := range forms {
		summaries[i] = dto.FormSummaryResponse{
			ID:          form.ID,
			Title:       form.Title,
			Description: form.Description,
			FieldCount:  len(form.Fields),
			RecordCount: counts[form.ID],
			CreatedAt:   form.CreatedAt,
			UpdatedAt:   form.UpdatedAt,
		}
	}
	return &dto.ListFormsResponse{Forms: summaries}, nil
}

func toFormResponse(form models.Form) (*dto.FormResponse, error) {
	fields := make([]dto.FormFieldResponse, len(form.Fields))
	for i, field := range form.Fields {
		options := []string{}
		if err := json.Unmarshal([]byte(field.Options), &options); err != nil {
			return nil, fmt.Errorf("decode options of field %d: %w", field.ID, err)
		}
		fields[i] = dto.FormFieldResponse{
			ID:       field.ID,
			Label:    field.Label,
			Type:     field.Type,
			Required: field.Required,
			Options:  options,
			Row:      field.LayoutRow,
		}
	}

	return &dto.FormResponse{
		ID:          form.ID,
		Title:       form.Title,
		Description: form.Description,
		Fields:      fields,
		CreatedAt:   form.CreatedAt,
		UpdatedAt:   form.UpdatedAt,
	}, nil
}
