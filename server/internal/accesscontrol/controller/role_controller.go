package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
)

// RoleController は /api/access-control/roles のエンドポイントを実装する。
// 権限のチェック（middleware.Require）はルート登録側で行う。
type RoleController interface {
	List() gin.HandlerFunc
	UpdatePermissions() gin.HandlerFunc
}

type roleController struct {
	service service.RoleService
}

// NewRoleController は指定のserviceを使う RoleController を作成する。
func NewRoleController(service service.RoleService) RoleController {
	return &roleController{service: service}
}

// List は GET /api/access-control/roles を処理する。
func (ctrl *roleController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.List(c.Request.Context())
		if err != nil {
			log.Printf("accesscontrol: list roles failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// UpdatePermissions は PUT /api/access-control/roles/:id/permissions を処理する。
func (ctrl *roleController) UpdatePermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
			return
		}

		var req dto.UpdateRolePermissionsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.UpdatePermissions(c.Request.Context(), uint(id), &req)
		if errors.Is(err, service.ErrRoleNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		var validationErr *service.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Message})
			return
		}
		if err != nil {
			log.Printf("accesscontrol: update role permissions failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
