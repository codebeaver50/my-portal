// Package router はポータル共通のアクセス管理の repository・service・controller を組み立て、ルートを登録する。
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/controller"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
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
