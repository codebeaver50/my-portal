package contact

import "github.com/gin-gonic/gin"

// RegisterRoutes はポートフォリオ共通の問い合わせフォームAPI（POST /api/contact）
// を登録する。
func RegisterRoutes(apiGroup *gin.RouterGroup) {
	ctrl := NewController(NewSMTPMailer())
	apiGroup.POST("/contact", ctrl.Submit())
}
