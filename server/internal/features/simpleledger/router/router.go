// Package router は simple ledger feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/service"
)

// SetupSimpleLedgerRoutes は simple ledger feature のルートをapiGroup配下に
// 登録する（例: GET /api/simple-ledger/transactions）。
func SetupSimpleLedgerRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	chartOfAccountsRepo := repository.NewChartOfAccountsRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	journalEntryRepo := repository.NewJournalEntryRepository(db)

	chartOfAccountsSvc := service.NewChartOfAccountsService(chartOfAccountsRepo)
	transactionSvc := service.NewTransactionService(db, transactionRepo, journalEntryRepo)

	chartOfAccountsCtrl := controller.NewChartOfAccountsController(chartOfAccountsSvc)
	transactionCtrl := controller.NewTransactionController(transactionSvc)

	ledger := apiGroup.Group("/simple-ledger")
	{
		ledger.GET("/chart-of-accounts", chartOfAccountsCtrl.List())

		transactions := ledger.Group("/transactions")
		{
			transactions.GET("", transactionCtrl.List())
			transactions.GET("/paginated", transactionCtrl.ListPaginated())
			transactions.POST("", transactionCtrl.Create())
			transactions.PUT("/:id", transactionCtrl.Update())
			transactions.DELETE("/:id", transactionCtrl.Delete())
		}
	}
}
