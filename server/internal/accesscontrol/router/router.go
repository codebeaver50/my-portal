// Package router はポータル共通のアクセス管理の repository・service・controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/controller"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
)

// SetupAuthRoutes はログインAPI（POST /api/auth/login, POST /api/auth/logout, GET /api/auth/session）を
// apiGroup配下に登録する。各featureで権限チェックに使う middleware.Authenticate 用に AuthService を返す。
func SetupAuthRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) service.AuthService {
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	authSvc := service.NewAuthService(userRepo, roleRepo, sessionRepo)
	authCtrl := controller.NewAuthController(authSvc)

	auth := apiGroup.Group("/auth")
	{
		auth.POST("/login", authCtrl.Login())
		auth.POST("/logout", authCtrl.Logout())
		auth.GET("/session", middleware.Authenticate(authSvc), authCtrl.Session())
	}

	return authSvc
}

// SetupAccessControlRoutes はアクセス管理画面のAPI（/api/access-control）を apiGroup 配下に登録する。
// 閲覧には access-control の read、変更には write の権限が必要。authSvc は SetupAuthRoutes の戻り値を渡す。
func SetupAccessControlRoutes(apiGroup *gin.RouterGroup, db *gorm.DB, authSvc service.AuthService) {
	roleRepo := repository.NewRoleRepository(db)
	roleSvc := service.NewRoleService(db, roleRepo)
	roleCtrl := controller.NewRoleController(roleSvc)

	group := apiGroup.Group("/access-control", middleware.Authenticate(authSvc))
	{
		group.GET("/roles", middleware.Require(permission.AccessControl, permission.Read), roleCtrl.List())
		group.PUT("/roles/:id/permissions", middleware.Require(permission.AccessControl, permission.Write), roleCtrl.UpdatePermissions())
	}
}
