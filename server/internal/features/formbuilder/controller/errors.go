package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/service"
)

// writeValidationError は err が service.ValidationError の場合に400レスポンスを書き込み、
// true を返す。項目ごとのエラーがあれば fieldErrors として返す。
func writeValidationError(c *gin.Context, err error) bool {
	var validationErr *service.ValidationError
	if !errors.As(err, &validationErr) {
		return false
	}

	body := gin.H{"error": validationErr.Message}
	if len(validationErr.FieldErrors) > 0 {
		body["fieldErrors"] = validationErr.FieldErrors
	}
	c.JSON(http.StatusBadRequest, body)
	return true
}

func parseID(c *gin.Context, param string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(param), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}
