// Package contact はポートフォリオ共通の問い合わせフォームAPIを提供する。
// 特定のfeatureに属さないため internal/features/ の配下ではなく、
// internal/admin/ と同様のfeature非依存の場所に置く。
package contact

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller interface {
	Submit() gin.HandlerFunc
}

type controller struct {
	mailer Mailer
}

func NewController(mailer Mailer) Controller {
	return &controller{mailer: mailer}
}

type submitRequest struct {
	Name    string `json:"name" binding:"required,max=100"`
	Email   string `json:"email" binding:"required,email"`
	Message string `json:"message" binding:"required,max=2000"`
}

func (ctrl *controller) Submit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req submitRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容を確認してください"})
			return
		}

		if err := ctrl.mailer.Send(req.Name, req.Email, req.Message); err != nil {
			log.Printf("contact: failed to send message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "送信に失敗しました。しばらくしてから再度お試しください"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
