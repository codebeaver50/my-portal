// Package router は book database feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/service"
)

// SetupBookRoutes は book database feature のルートをapiGroup配下に登録
// する（例: GET /api/books）。
func SetupBookRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	bookRepo := repository.NewBookRepository(db)
	publisherRepo := repository.NewPublisherRepository(db)
	authorRepo := repository.NewAuthorRepository(db)

	bookSvc := service.NewBookService(bookRepo)
	syncSvc := service.NewSyncService(db, bookRepo, publisherRepo, authorRepo, service.NewNDLClient(), service.NewOpenBDClient())

	bookCtrl := controller.NewBookController(bookSvc)
	syncCtrl := controller.NewSyncController(syncSvc)

	books := apiGroup.Group("/books")
	{
		books.GET("", bookCtrl.List())
		books.GET("/:id/cover", bookCtrl.Cover())

		sync := books.Group("/sync")
		sync.Use(controller.CronAuthMiddleware())
		{
			sync.GET("", syncCtrl.Sync())
			sync.POST("", syncCtrl.Sync())
		}
	}
}
