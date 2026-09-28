// Package controller は simple ledger feature の /api/simple-ledger エンドポイント
// のHTTPハンドラを実装する。
package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/service"
)

// ChartOfAccountsController は GET /api/simple-ledger/chart-of-accounts を実装する。
type ChartOfAccountsController interface {
	List() gin.HandlerFunc
}

type chartOfAccountsController struct {
	service service.ChartOfAccountsService
}

// NewChartOfAccountsController は指定のserviceを使う ChartOfAccountsController を作成する。
func NewChartOfAccountsController(service service.ChartOfAccountsService) ChartOfAccountsController {
	return &chartOfAccountsController{service: service}
}

func (ctrl *chartOfAccountsController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.ListActive(c.Request.Context())
		if err != nil {
			log.Printf("simple-ledger: list chart of accounts failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
