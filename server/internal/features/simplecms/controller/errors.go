package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/service"
)

// writeValidationError は err が service.ValidationError の場合に400レスポンスを書き込み、
// true を返す。
func writeValidationError(c *gin.Context, err error) bool {
	var validationErr *service.ValidationError
	if !errors.As(err, &validationErr) {
		return false
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})
	return true
}
