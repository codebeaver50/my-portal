// Package router は simple ledger feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	acmiddleware "github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	acservice "github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/service"
)

// SetupSimpleLedgerRoutes は simple ledger feature のルートをapiGroup配下に
// 登録する（例: GET /api/simple-ledger/transactions）。取引の作成・更新・削除には
// アクセス管理（RBAC）で simple-ledger の write 権限が必要。authSvc は accesscontrol の
// router.SetupAuthRoutes の戻り値を渡す。
func SetupSimpleLedgerRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, authSvc acservice.AuthService) {
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

			writes := transactions.Group("", acmiddleware.Authenticate(authSvc), acmiddleware.Require(permission.SimpleLedger, permission.Write))
			writes.POST("", transactionCtrl.Create())
			writes.PUT("/:id", transactionCtrl.Update())
			writes.DELETE("/:id", transactionCtrl.Delete())
		}
	}
}
