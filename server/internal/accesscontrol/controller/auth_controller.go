// Package controller はポータル共通のアクセス管理の /api/auth エンドポイントのHTTPハンドラを実装する。
package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/middleware"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
)

// AuthController は /api/auth のエンドポイントを実装する。
type AuthController interface {
	Login() gin.HandlerFunc
	Logout() gin.HandlerFunc
	Session() gin.HandlerFunc
}

type authController struct {
	service service.AuthService
}

// NewAuthController は指定のserviceを使う AuthController を作成する。
func NewAuthController(service service.AuthService) AuthController {
	return &authController{service: service}
}

// Login は POST /api/auth/login を処理する。成功するとセッションCookieを発行し、ログイン後のセッション情報を返す。
func (ctrl *authController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ログインIDとパスワードを入力してください"})
			return
		}

		meta := service.LoginMeta{UserAgent: c.Request.UserAgent(), IPAddress: c.ClientIP()}
		token, principal, err := ctrl.service.Login(c.Request.Context(), &req, meta)
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ログインIDまたはパスワードが違います"})
			return
		}
		if errors.Is(err, service.ErrAccountLocked) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "ログインの失敗が続いたため、一時的にログインできません。しばらくしてから再度お試しください"})
			return
		}
		if err != nil {
			log.Printf("accesscontrol: login failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// 別のアカウントでログインし直した場合、上書きされて使われなくなる前のセッションを消しておく
		if previous, _ := c.Cookie(middleware.SessionCookieName); previous != "" {
			if err := ctrl.service.Logout(c.Request.Context(), previous); err != nil {
				log.Printf("accesscontrol: failed to delete the previous session: %v", err)
			}
		}

		setSessionCookie(c, token)
		c.JSON(http.StatusOK, service.ToSessionResponse(principal))
	}
}

// Logout は POST /api/auth/logout を処理する。セッションをDBから削除し、Cookieも消す。
func (ctrl *authController) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(middleware.SessionCookieName)
		if err := ctrl.service.Logout(c.Request.Context(), token); err != nil {
			log.Printf("accesscontrol: logout failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		clearSessionCookie(c)
		c.Status(http.StatusNoContent)
	}
}

// Session は GET /api/auth/session を処理する。ログイン中のユーザーと、リソースごとの権限を返す
// （未ログインの場合は user が null で、権限はゲストロールのもの）。middleware.Authenticate より後に登録すること。
func (ctrl *authController) Session() gin.HandlerFunc {
	return func(c *gin.Context) {
		principal := middleware.PrincipalFrom(c)
		if principal == nil {
			log.Print("accesscontrol: session handler used without Authenticate")
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, service.ToSessionResponse(principal))
	}
}

// secureCookie は本番（GinのReleaseMode）ではCookieをHTTPSのみで送るようにする。
// ローカル開発はHTTPのため付けない。
func secureCookie() bool {
	return gin.Mode() == gin.ReleaseMode
}

// setSessionCookie はセッションCookieを発行する。JavaScriptから読めないようHttpOnlyにし、
// 他サイトからのPOSTにCookieが付かないようSameSite=Laxにする（CSRF対策）。
func setSessionCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.SessionCookieName, token, int(service.SessionTTL.Seconds()), "/", "", secureCookie(), true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.SessionCookieName, "", -1, "/", "", secureCookie(), true)
}
