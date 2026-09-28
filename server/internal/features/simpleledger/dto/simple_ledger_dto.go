// Package dto は simple ledger feature のHTTP APIにおけるリクエスト/レスポンス型を定義する。
// service層がこれを組み立ててcontrollerに返す。
package dto

import (
	"time"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// ChartOfAccountResponse は勘定科目1件分。
type ChartOfAccountResponse struct {
	ID            uint   `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	NormalBalance string `json:"normalBalance"`
}

// ListChartOfAccountsResponse は GET /api/simple-ledger/chart-of-accounts のレスポンス。
type ListChartOfAccountsResponse struct {
	Accounts []ChartOfAccountResponse `json:"accounts"`
}

// JournalEntryRequest は取引作成/更新リクエストに含まれる仕訳明細1行。
type JournalEntryRequest struct {
	ChartOfAccountsID uint             `json:"chartOfAccountsId" binding:"required"`
	Type              models.EntryType `json:"type" binding:"required,oneof=debit credit"`
	Amount            int              `json:"amount" binding:"required,gt=0"`
	Description       string           `json:"description"`
}

// TransactionRequest: 取引の作成/更新リクエスト。
type TransactionRequest struct {
	Date           string                `json:"date" binding:"required"`
	Description    string                `json:"description" binding:"max=255"`
	JournalEntries []JournalEntryRequest `json:"journalEntries" binding:"required,min=2"`
	// CorrectionNote: 更新リクエストでのみ使う。指定すると既存取引を書き換えず、
	// 新しい訂正取引として記録する。
	CorrectionNote string `json:"correctionNote" binding:"max=255"`
}

// JournalEntryResponse は取引レスポンスに含まれる仕訳明細1行。
type JournalEntryResponse struct {
	ID                uint                    `json:"id"`
	ChartOfAccountsID uint                    `json:"chartOfAccountsId"`
	ChartOfAccounts   *ChartOfAccountResponse `json:"chartOfAccounts,omitempty"`
	Type              models.EntryType        `json:"type"`
	Amount            int                     `json:"amount"`
	Description       string                  `json:"description"`
}

// TransactionResponse: 取引レスポンス。
type TransactionResponse struct {
	ID              uint                   `json:"id"`
	Date            string                 `json:"date"`
	Description     string                 `json:"description"`
	JournalEntries  []JournalEntryResponse `json:"journalEntries,omitempty"`
	IsCorrection    bool                   `json:"isCorrection"`
	CorrectedFromID *uint                  `json:"correctedFromId,omitempty"`
	CorrectionNote  string                 `json:"correctionNote"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// ListTransactionsResponse は GET /api/simple-ledger/transactions のレスポンス。
type ListTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
}

// PaginatedTransactionsResponse は GET /api/simple-ledger/transactions/paginated のレスポンス。
type PaginatedTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
	Page         int                   `json:"page"`
	PageSize     int                   `json:"pageSize"`
	HasNextPage  bool                  `json:"hasNextPage"`
}

// ToChartOfAccountResponse は ChartOfAccounts モデルを ChartOfAccountResponse に変換する。
func ToChartOfAccountResponse(account models.ChartOfAccounts) ChartOfAccountResponse {
	return ChartOfAccountResponse{
		ID:            account.ID,
		Code:          account.Code,
		Name:          account.Name,
		Type:          string(account.Type),
		NormalBalance: string(account.NormalBalance),
	}
}

// ToTransactionResponse は Transaction モデルを TransactionResponse に変換する。
func ToTransactionResponse(transaction models.Transaction) TransactionResponse {
	response := TransactionResponse{
		ID:              transaction.ID,
		Date:            transaction.Date.Format("2006-01-02"),
		Description:     transaction.Description,
		IsCorrection:    transaction.IsCorrection,
		CorrectedFromID: transaction.CorrectedFromID,
		CorrectionNote:  transaction.CorrectionNote,
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
	}

	if len(transaction.JournalEntries) > 0 {
		response.JournalEntries = make([]JournalEntryResponse, len(transaction.JournalEntries))
		for i, entry := range transaction.JournalEntries {
			entryResponse := JournalEntryResponse{
				ID:                entry.ID,
				ChartOfAccountsID: entry.ChartOfAccountsID,
				Type:              entry.Type,
				Amount:            entry.Amount,
				Description:       entry.Description,
			}
			if entry.ChartOfAccounts != nil {
				accountResponse := ToChartOfAccountResponse(*entry.ChartOfAccounts)
				entryResponse.ChartOfAccounts = &accountResponse
			}
			response.JournalEntries[i] = entryResponse
		}
	}

	return response
}
