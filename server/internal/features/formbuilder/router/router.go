// Package router は form builder feature の repository・service・
// controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/controller"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/repository"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/service"
)

// SetupFormBuilderRoutes は form builder feature のルートをapiGroup配下に
// 登録する（例: GET /api/form-builder/forms）。
func SetupFormBuilderRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
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
		forms.POST("", formCtrl.Create())
		forms.GET("/:id", formCtrl.Get())
		forms.PUT("/:id", formCtrl.Update())
		forms.DELETE("/:id", formCtrl.Delete())

		forms.GET("/:id/records", recordCtrl.List())
		forms.POST("/:id/records", recordCtrl.Create())
		forms.DELETE("/:id/records/:recordId", recordCtrl.Delete())
	}
}
