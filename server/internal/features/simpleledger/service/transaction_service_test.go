package service

import (
	"context"
	"errors"
	"testing"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

func newTransactionService(t *testing.T) (TransactionService, uint, uint) {
	t.Helper()
	db := newTestDB(t)
	cashID, revenueID := seedAccounts(t, db)
	svc := NewTransactionService(db, repository.NewTransactionRepository(db), repository.NewJournalEntryRepository(db))
	return svc, cashID, revenueID
}

func balancedEntries(cashID, revenueID uint, amount int) []dto.JournalEntryRequest {
	return []dto.JournalEntryRequest{
		{ChartOfAccountsID: cashID, Type: models.DebitEntry, Amount: amount},
		{ChartOfAccountsID: revenueID, Type: models.CreditEntry, Amount: amount},
	}
}

func TestTransactionService_Create(t *testing.T) {
	t.Run("creates a balanced transaction", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)

		result, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date:           "2026-01-15",
			Description:    "売上入金",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Date != "2026-01-15" || len(result.JournalEntries) != 2 {
			t.Errorf("result = %+v", result)
		}
	})

	t.Run("rejects unbalanced entries", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)

		_, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date: "2026-01-15",
			JournalEntries: []dto.JournalEntryRequest{
				{ChartOfAccountsID: cashID, Type: models.DebitEntry, Amount: 1000},
				{ChartOfAccountsID: revenueID, Type: models.CreditEntry, Amount: 500},
			},
		})
		if !errors.Is(err, ErrUnbalancedEntries) {
			t.Errorf("err = %v, want ErrUnbalancedEntries", err)
		}
	})

	t.Run("rejects entries missing a debit or credit side", func(t *testing.T) {
		svc, cashID, _ := newTransactionService(t)

		_, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date: "2026-01-15",
			JournalEntries: []dto.JournalEntryRequest{
				{ChartOfAccountsID: cashID, Type: models.DebitEntry, Amount: 1000},
				{ChartOfAccountsID: cashID, Type: models.DebitEntry, Amount: 1000},
			},
		})
		if !errors.Is(err, ErrUnbalancedEntries) {
			t.Errorf("err = %v, want ErrUnbalancedEntries", err)
		}
	})

	t.Run("rejects invalid date format", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)

		_, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date:           "15-01-2026",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		})
		if err == nil {
			t.Error("expected an error for invalid date format")
		}
	})
}

func TestTransactionService_Update(t *testing.T) {
	t.Run("in-place edit replaces journal entries", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)
		created, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date:           "2026-01-15",
			Description:    "元の取引",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		updated, err := svc.Update(context.Background(), created.ID, &dto.TransactionRequest{
			Date:           "2026-01-16",
			Description:    "編集後",
			JournalEntries: balancedEntries(cashID, revenueID, 2000),
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.ID != created.ID {
			t.Errorf("updated.ID = %d, want %d (in-place edit keeps the same id)", updated.ID, created.ID)
		}
		if updated.Description != "編集後" || updated.JournalEntries[0].Amount != 2000 {
			t.Errorf("updated = %+v", updated)
		}
	})

	t.Run("correction note creates a new correction transaction", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)
		created, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date:           "2026-01-15",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		corrected, err := svc.Update(context.Background(), created.ID, &dto.TransactionRequest{
			Date:           "2026-01-15",
			Description:    "金額の訂正",
			JournalEntries: balancedEntries(cashID, revenueID, 1500),
			CorrectionNote: "入力ミスのため",
		})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if corrected.ID == created.ID {
			t.Error("correction should create a new transaction, not overwrite the original")
		}
		if !corrected.IsCorrection || corrected.CorrectedFromID == nil || *corrected.CorrectedFromID != created.ID {
			t.Errorf("corrected = %+v", corrected)
		}
	})

	t.Run("unknown id returns ErrTransactionNotFound", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)

		_, err := svc.Update(context.Background(), 999, &dto.TransactionRequest{
			Date:           "2026-01-15",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		})
		if !errors.Is(err, ErrTransactionNotFound) {
			t.Errorf("err = %v, want ErrTransactionNotFound", err)
		}
	})
}

func TestTransactionService_Delete(t *testing.T) {
	t.Run("deletes an existing transaction", func(t *testing.T) {
		svc, cashID, revenueID := newTransactionService(t)
		created, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date:           "2026-01-15",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if err := svc.Delete(context.Background(), created.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		all, err := svc.ListAll(context.Background())
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}
		if all.Total != 0 {
			t.Errorf("Total = %d, want 0 after delete", all.Total)
		}
	})

	t.Run("unknown id returns ErrTransactionNotFound", func(t *testing.T) {
		svc, _, _ := newTransactionService(t)

		if err := svc.Delete(context.Background(), 999); !errors.Is(err, ErrTransactionNotFound) {
			t.Errorf("err = %v, want ErrTransactionNotFound", err)
		}
	})
}

func TestTransactionService_ListPaginated(t *testing.T) {
	svc, cashID, revenueID := newTransactionService(t)
	for i := 0; i < 3; i++ {
		if _, err := svc.Create(context.Background(), &dto.TransactionRequest{
			Date:           "2026-01-15",
			Description:    "取引",
			JournalEntries: balancedEntries(cashID, revenueID, 1000),
		}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	page, err := svc.ListPaginated(context.Background(), 1, 2, "")
	if err != nil {
		t.Fatalf("ListPaginated() error = %v", err)
	}
	if page.Total != 3 || len(page.Transactions) != 2 || !page.HasNextPage {
		t.Errorf("page = %+v", page)
	}

	page2, err := svc.ListPaginated(context.Background(), 2, 2, "")
	if err != nil {
		t.Fatalf("ListPaginated() error = %v", err)
	}
	if len(page2.Transactions) != 1 || page2.HasNextPage {
		t.Errorf("page2 = %+v", page2)
	}
}
