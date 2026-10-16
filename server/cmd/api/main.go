package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	accesscontrolrouter "github.com/codebeaver50/my-portal/server/internal/accesscontrol/router"
	"github.com/codebeaver50/my-portal/server/internal/contact"
	"github.com/codebeaver50/my-portal/server/internal/db"
	bookdatabaserouter "github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/router"
	formbuilderrouter "github.com/codebeaver50/my-portal/server/internal/features/formbuilder/router"
	simplecmsrouter "github.com/codebeaver50/my-portal/server/internal/features/simplecms/router"
	simpleledgerrouter "github.com/codebeaver50/my-portal/server/internal/features/simpleledger/router"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	conn, err := db.Open(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	readOnly := isReadOnly(os.Getenv("READ_ONLY"))
	if readOnly {
		log.Print("READ_ONLY is enabled: write requests will be rejected")
	}

	engine := gin.Default()
	engine.Use(corsMiddleware(allowedOrigin))
	// /api/books/sync はcron等からしか叩かれない（CronAuthMiddlewareで別途保護
	// されている）バッチ更新のため、READ_ONLYの対象外にする。
	// simple-cmsのカテゴリ・記事の書き込み系は、アクセス管理（RBAC）の simple-cms の
	// write 権限で保護されており、サイト運営者が作成したアカウント（管理者等）だけが
	// 書き込めるため、本番でも操作できるよう対象外にする。ゲストロールには simple-cms の
	// write を付けないこと（付けると READ_ONLY でも未ログインの訪問者が書き込めてしまう）。
	// /api/auth/login・/api/auth/logout はセッションの行を書き込むが、デモデータを
	// 変更しないうえ、止めると本番でログインできなくなるため対象外にする。
	// /api/contact は認可のない公開エンドポイントだが、DB/デモデータを一切変更
	// せずメール転送するだけなので、READ_ONLYが保護したい対象（シードデータ）に
	// 当たらず対象外にする。
	engine.Use(readOnlyMiddleware(readOnly,
		"/api/books/sync",
		"/api/auth/login",
		"/api/auth/logout",
		"/api/simple-cms/categories",
		"/api/simple-cms/categories/:id",
		"/api/simple-cms/posts",
		"/api/simple-cms/posts/:id",
		"/api/contact",
	))

	engine.GET("/health", handleHealth)
	apiGroup := engine.Group("/api")
	authSvc := accesscontrolrouter.SetupAuthRoutes(apiGroup, conn)
	accesscontrolrouter.SetupAccessControlRoutes(apiGroup, conn, authSvc)
	contact.RegisterRoutes(apiGroup)
	bookdatabaserouter.SetupBookRoutes(apiGroup, conn, os.Getenv("CRON_SECRET"))
	simpleledgerrouter.SetupSimpleLedgerRoutes(apiGroup, conn, authSvc)
	simplecmsrouter.SetupSimpleCmsRoutes(apiGroup, conn, authSvc)
	formbuilderrouter.SetupFormBuilderRoutes(apiGroup, conn, authSvc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	if err := engine.Run(":" + port); err != nil {
		log.Printf("server exited: %v", err)
	}
}

func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// isReadOnly は環境変数 READ_ONLY の値を解釈する。"true" か "1"（大文字小文字を
// 区別しない）以外の値は、すべて無効として扱う。
func isReadOnly(value string) bool {
	switch strings.ToLower(value) {
	case "true", "1":
		return true
	default:
		return false
	}
}

// readOnlyMiddleware は、有効な場合にAPI全体の書き込み系リクエストを拒否する。
// 公開デモで、訪問者にシードデータを書き換えられないようにするため。
// GET/HEAD/OPTIONS（CORSのプリフライトは corsMiddleware で処理済み）は常に通す。
// exemptPaths に載せたルートも通す。ユーザーの操作からは呼ばれない、別の認可で
// 保護された書き込み（cron専用の同期処理等）のためのもの。
func readOnlyMiddleware(enabled bool, exemptPaths ...string) gin.HandlerFunc {
	exempt := make(map[string]bool, len(exemptPaths))
	for _, path := range exemptPaths {
		exempt[path] = true
	}

	return func(c *gin.Context) {
		if !enabled || exempt[c.FullPath()] {
			c.Next()
			return
		}

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "読み取り専用デモのため、この操作は無効です",
			})
		}
	}
}
