// Package middleware はセッションCookieからリクエストの主体（Principal）を解決し、
// リソースごとの権限をチェックするGinミドルウェアを提供する。各featureのルート登録から使う。
package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
)

// SessionCookieName はセッショントークンを入れるCookieの名前。
const SessionCookieName = "portal_session"

const principalKey = "accesscontrol.principal"

// Authenticate はセッションCookieから Principal を解決してコンテキストに入れる。
// Cookieがない・無効な場合もゲストの Principal を入れて次へ進む（拒否するかどうかは Require が決める）。
func Authenticate(auth service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(SessionCookieName)
		principal, err := auth.Authenticate(c.Request.Context(), token)
		if err != nil {
			log.Printf("accesscontrol: authenticate failed: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Set(principalKey, principal)
		c.Next()
	}
}

// Require は resource に required 以上の権限を要求する。Authenticate より後に登録すること。
// 未ログインで権限が足りない場合は401（ログインすれば操作できる可能性がある）、
// ログイン済みで足りない場合は403を返す。
func Require(resource permission.Resource, required permission.Level) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal := PrincipalFrom(c)
		if principal == nil {
			log.Printf("accesscontrol: Require(%s) used without Authenticate", resource)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if principal.Permissions.Allows(resource, required) {
			c.Next()
			return
		}

		if principal.User == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "この操作にはログインが必要です"})
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "この操作を行う権限がありません"})
	}
}

// PrincipalFrom は Authenticate がコンテキストに入れた Principal を返す。未設定の場合は nil。
func PrincipalFrom(c *gin.Context) *service.Principal {
	value, ok := c.Get(principalKey)
	if !ok {
		return nil
	}
	principal, _ := value.(*service.Principal)
	return principal
}
