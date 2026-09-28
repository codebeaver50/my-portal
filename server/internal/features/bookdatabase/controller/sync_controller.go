package controller

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/service"
)

// SyncController は GET/POST /api/books/sync バッチエンドポイントを実装する。
type SyncController interface {
	Sync() gin.HandlerFunc
}

type syncController struct {
	service service.SyncService
}

// NewSyncController は指定のserviceを使う SyncController を作成する。
func NewSyncController(service service.SyncService) SyncController {
	return &syncController{service: service}
}

// Sync は GET/POST /api/books/sync を処理する。認可は CronAuthMiddleware
// で行う。
func (ctrl *syncController) Sync() gin.HandlerFunc {
	return func(c *gin.Context) {
		summary, err := ctrl.service.Sync(c.Request.Context())
		if err != nil {
			log.Printf("books: sync failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, summary)
	}
}

// CronAuthMiddleware は CRON_SECRET ベアラートークンによる認可を行う
// Ginミドルウェア。Vercel Cron job として動いていた頃のチェックを踏襲して
// おり、CRON_SECRET未設定時はローカル開発用にチェックをスキップする。
func CronAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cronSecret := os.Getenv("CRON_SECRET")
		if cronSecret != "" && c.Request.Header.Get("Authorization") != "Bearer "+cronSecret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}
