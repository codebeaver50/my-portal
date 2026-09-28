package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// ErrTransactionNotFound は指定のIDの取引が存在しない場合に返される。
var ErrTransactionNotFound = errors.New("transaction not found")

// ErrUnbalancedEntries は仕訳明細が複式簿記のルール（借方合計=貸方合計、
// 借方・貸方それぞれ最低1件）を満たさない場合に返される。
var ErrUnbalancedEntries = errors.New("debit total must equal credit total, and at least one debit and one credit entry are required")

// TransactionService は取引の作成・更新・削除・一覧取得と複式簿記のバリデーションを実装する。
type TransactionService interface {
	Create(ctx context.Context, req *dto.TransactionRequest) (*dto.TransactionResponse, error)
	Update(ctx context.Context, id uint, req *dto.TransactionRequest) (*dto.TransactionResponse, error)
	Delete(ctx context.Context, id uint) error
	ListAll(ctx context.Context) (*dto.ListTransactionsResponse, error)
	ListPaginated(ctx context.Context, page, pageSize int, keyword string) (*dto.PaginatedTransactionsResponse, error)
}

type transactionService struct {
	db               *gorm.DB
	transactionRepo  *repository.TransactionRepository
	journalEntryRepo *repository.JournalEntryRepository
}

// NewTransactionService は指定のrepositoryを使う TransactionService を作成する。
func NewTransactionService(
	db *gorm.DB,
	transactionRepo *repository.TransactionRepository,
	journalEntryRepo *repository.JournalEntryRepository,
) TransactionService {
	return &transactionService{db: db, transactionRepo: transactionRepo, journalEntryRepo: journalEntryRepo}
}

// validateEntries は複式簿記のルール（借方・貸方それぞれ最低1件、借方合計=貸方合計）を検証する。
func validateEntries(entries []dto.JournalEntryRequest) error {
	hasDebit, hasCredit := false, false
	debitTotal, creditTotal := 0, 0

	for _, entry := range entries {
		switch entry.Type {
		case models.DebitEntry:
			hasDebit = true
			debitTotal += entry.Amount
		case models.CreditEntry:
			hasCredit = true
			creditTotal += entry.Amount
		}
	}

	if !hasDebit || !hasCredit || debitTotal != creditTotal {
		return ErrUnbalancedEntries
	}
	return nil
}

func toJournalEntries(transactionID uint, entries []dto.JournalEntryRequest) []models.JournalEntry {
	result := make([]models.JournalEntry, len(entries))
	for i, entry := range entries {
		result[i] = models.JournalEntry{
			TransactionID:     transactionID,
			ChartOfAccountsID: entry.ChartOfAccountsID,
			Type:              entry.Type,
			Amount:            entry.Amount,
			Description:       entry.Description,
		}
	}
	return result
}

func (s *transactionService) Create(ctx context.Context, req *dto.TransactionRequest) (*dto.TransactionResponse, error) {
	if err := validateEntries(req.JournalEntries); err != nil {
		return nil, err
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	var transactionID uint
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transactionRepo := s.transactionRepo.WithTx(tx)
		journalEntryRepo := s.journalEntryRepo.WithTx(tx)

		transaction := models.Transaction{Date: date, Description: req.Description}
		if err := transactionRepo.Create(ctx, &transaction); err != nil {
			return err
		}
		transactionID = transaction.ID

		return journalEntryRepo.CreateBatch(ctx, toJournalEntries(transaction.ID, req.JournalEntries))
	})
	if err != nil {
		return nil, err
	}

	return s.getResponse(ctx, transactionID)
}

func (s *transactionService) Update(ctx context.Context, id uint, req *dto.TransactionRequest) (*dto.TransactionResponse, error) {
	existing, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrTransactionNotFound
	}

	if err := validateEntries(req.JournalEntries); err != nil {
		return nil, err
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	// 訂正理由が指定された場合は、既存の取引を書き換えず新しい訂正取引として記録する。
	if req.CorrectionNote != "" {
		var correctionID uint
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			transactionRepo := s.transactionRepo.WithTx(tx)
			journalEntryRepo := s.journalEntryRepo.WithTx(tx)

			correction := models.Transaction{
				Date:            date,
				Description:     req.Description,
				IsCorrection:    true,
				CorrectedFromID: &id,
				CorrectionNote:  req.CorrectionNote,
			}
			if err := transactionRepo.Create(ctx, &correction); err != nil {
				return err
			}
			correctionID = correction.ID

			return journalEntryRepo.CreateBatch(ctx, toJournalEntries(correction.ID, req.JournalEntries))
		})
		if err != nil {
			return nil, err
		}
		return s.getResponse(ctx, correctionID)
	}

	// 訂正理由なし：既存取引を仕訳ごと入れ替えて更新する。
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transactionRepo := s.transactionRepo.WithTx(tx)
		journalEntryRepo := s.journalEntryRepo.WithTx(tx)

		existing.Date = date
		existing.Description = req.Description
		if err := transactionRepo.Update(ctx, existing); err != nil {
			return err
		}

		if err := journalEntryRepo.DeleteByTransactionID(ctx, id); err != nil {
			return err
		}
		return journalEntryRepo.CreateBatch(ctx, toJournalEntries(id, req.JournalEntries))
	})
	if err != nil {
		return nil, err
	}

	return s.getResponse(ctx, id)
}

func (s *transactionService) Delete(ctx context.Context, id uint) error {
	existing, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrTransactionNotFound
	}
	// 仕訳エントリーはFKのON DELETE CASCADEで一緒に削除される。
	return s.transactionRepo.Delete(ctx, id)
}

func (s *transactionService) ListAll(ctx context.Context) (*dto.ListTransactionsResponse, error) {
	transactions, err := s.transactionRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = dto.ToTransactionResponse(transaction)
	}

	return &dto.ListTransactionsResponse{Transactions: responses, Total: len(responses)}, nil
}

func (s *transactionService) ListPaginated(ctx context.Context, page, pageSize int, keyword string) (*dto.PaginatedTransactionsResponse, error) {
	transactions, total, err := s.transactionRepo.GetPaginated(ctx, page, pageSize, keyword)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = dto.ToTransactionResponse(transaction)
	}

	return &dto.PaginatedTransactionsResponse{
		Transactions: responses,
		Total:        int(total),
		Page:         page,
		PageSize:     pageSize,
		HasNextPage:  int64(page*pageSize) < total,
	}, nil
}

func (s *transactionService) getResponse(ctx context.Context, id uint) (*dto.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, ErrTransactionNotFound
	}
	response := dto.ToTransactionResponse(*transaction)
	return &response, nil
}
