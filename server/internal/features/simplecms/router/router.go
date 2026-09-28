// Package router は simple cms feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/admin"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/service"
)

// SetupSimpleCmsRoutes は simple cms feature のルートをapiGroup配下に
// 登録する（例: GET /api/simple-cms/posts）。作成・更新・削除は
// admin.AuthMiddleware（ポートフォリオ共通の管理者ログイン）で保護する。
func SetupSimpleCmsRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db)

	categorySvc := service.NewCategoryService(categoryRepo)
	postSvc := service.NewPostService(postRepo, categoryRepo)

	categoryCtrl := controller.NewCategoryController(categorySvc)
	postCtrl := controller.NewPostController(postSvc)

	adminAuth := admin.AuthMiddleware()

	cms := apiGroup.Group("/simple-cms")
	{
		categories := cms.Group("/categories")
		{
			categories.GET("", categoryCtrl.List())
			categories.POST("", adminAuth, categoryCtrl.Create())
			categories.PUT("/:id", adminAuth, categoryCtrl.Update())
			categories.DELETE("/:id", adminAuth, categoryCtrl.Delete())
		}

		posts := cms.Group("/posts")
		{
			posts.GET("", postCtrl.List())
			posts.GET("/:slug", postCtrl.GetBySlug())
			posts.POST("", adminAuth, postCtrl.Create())
			posts.PUT("/:id", adminAuth, postCtrl.Update())
			posts.DELETE("/:id", adminAuth, postCtrl.Delete())
		}
	}
}
