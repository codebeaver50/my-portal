// Package router は form builder feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	acmiddleware "github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	acservice "github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/service"
)

// SetupFormBuilderRoutes は form builder feature のルートをapiGroup配下に
// 登録する（例: GET /api/form-builder/forms）。フォームの作成・更新・削除と、回答の投稿・削除には
// アクセス管理（RBAC）で form-builder の write 権限が必要。authSvc は accesscontrol の
// router.SetupAuthRoutes の戻り値を渡す。
func SetupFormBuilderRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, authSvc acservice.AuthService) {
	formRepo := repository.NewFormRepository(db)
	fieldRepo := repository.NewFormFieldRepository(db)
	recordRepo := repository.NewFormRecordRepository(db)

	formSvc := service.NewFormService(db, formRepo, fieldRepo, recordRepo)
	recordSvc := service.NewRecordService(formRepo, recordRepo)

	formCtrl := controller.NewFormController(formSvc)
	recordCtrl := controller.NewRecordController(recordSvc)

	forms := apiGroup.Group("/form-builder/forms")
	{
		forms.GET("", formCtrl.List())
		forms.GET("/:id", formCtrl.Get())
		forms.GET("/:id/records", recordCtrl.List())

		writes := forms.Group("", acmiddleware.Authenticate(authSvc), acmiddleware.Require(permission.FormBuilder, permission.Write))
		writes.POST("", formCtrl.Create())
		writes.PUT("/:id", formCtrl.Update())
		writes.DELETE("/:id", formCtrl.Delete())
		writes.POST("/:id/records", recordCtrl.Create())
		writes.DELETE("/:id/records/:recordId", recordCtrl.Delete())
	}
}
