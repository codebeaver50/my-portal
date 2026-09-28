// Package service は simple ledger feature のビジネスロジック（複式簿記の
// バリデーション・訂正フローの組み立て）を実装する。
package service

import (
	"context"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/repository"
)

// ChartOfAccountsService は GET /api/simple-ledger/chart-of-accounts を実装する。
type ChartOfAccountsService interface {
	ListActive(ctx context.Context) (*dto.ListChartOfAccountsResponse, error)
}

type chartOfAccountsService struct {
	repo *repository.ChartOfAccountsRepository
}

// NewChartOfAccountsService は指定のrepositoryを使う ChartOfAccountsService を作成する。
func NewChartOfAccountsService(repo *repository.ChartOfAccountsRepository) ChartOfAccountsService {
	return &chartOfAccountsService{repo: repo}
}

func (s *chartOfAccountsService) ListActive(ctx context.Context) (*dto.ListChartOfAccountsResponse, error) {
	accounts, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChartOfAccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = dto.ToChartOfAccountResponse(account)
	}

	return &dto.ListChartOfAccountsResponse{Accounts: responses}, nil
}
