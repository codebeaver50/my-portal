// Package router は simple cms feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	acmiddleware "github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	acservice "github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/service"
)

// SetupSimpleCmsRoutes は simple cms feature のルートをapiGroup配下に
// 登録する（例: GET /api/simple-cms/posts）。カテゴリ・記事の作成・更新・削除には
// アクセス管理（RBAC）で simple-cms の write 権限が必要。authSvc は accesscontrol の
// router.SetupAuthRoutes の戻り値を渡す。
func SetupSimpleCmsRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, authSvc acservice.AuthService) {
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db)

	categorySvc := service.NewCategoryService(categoryRepo, postRepo)
	postSvc := service.NewPostService(postRepo, categoryRepo)

	categoryCtrl := controller.NewCategoryController(categorySvc)
	postCtrl := controller.NewPostController(postSvc)

	canWrite := []gin.HandlerFunc{acmiddleware.Authenticate(authSvc), acmiddleware.Require(permission.SimpleCMS, permission.Write)}

	cms := apiGroup.Group("/simple-cms")
	{
		categories := cms.Group("/categories")
		{
			categories.GET("", categoryCtrl.List())

			writes := categories.Group("", canWrite...)
			writes.POST("", categoryCtrl.Create())
			writes.PUT("/:id", categoryCtrl.Update())
			writes.DELETE("/:id", categoryCtrl.Delete())
		}

		posts := cms.Group("/posts")
		{
			posts.GET("", postCtrl.List())
			posts.GET("/:slug", postCtrl.GetBySlug())

			writes := posts.Group("", canWrite...)
			writes.POST("", postCtrl.Create())
			writes.PUT("/:id", postCtrl.Update())
			writes.DELETE("/:id", postCtrl.Delete())
		}
	}
}
