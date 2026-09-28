// Package controller は simple cms feature の /api/simple-cms エンドポイント
// のHTTPハンドラを実装する。
package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simplecms/service"
)

// CategoryController は /api/simple-cms/categories のエンドポイントを実装する。
type CategoryController interface {
	List() gin.HandlerFunc
	Create() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type categoryController struct {
	service service.CategoryService
}

// NewCategoryController は指定のserviceを使う CategoryController を作成する。
func NewCategoryController(service service.CategoryService) CategoryController {
	return &categoryController{service: service}
}

// List は GET /api/simple-cms/categories を処理する。
func (ctrl *categoryController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.ListAll(c.Request.Context())
		if err != nil {
			log.Printf("simple-cms: list categories failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Create は POST /api/simple-cms/categories を処理する。
func (ctrl *categoryController) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Create(c.Request.Context(), &req)
		if errors.Is(err, service.ErrCategorySlugTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

// Update は PUT /api/simple-cms/categories/:id を処理する。
func (ctrl *categoryController) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
			return
		}

		var req dto.CategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Update(c.Request.Context(), uint(id), &req)
		if errors.Is(err, service.ErrCategoryNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrCategorySlugTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Delete は DELETE /api/simple-cms/categories/:id を処理する。
func (ctrl *categoryController) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
			return
		}

		err = ctrl.service.Delete(c.Request.Context(), uint(id))
		if errors.Is(err, service.ErrCategoryNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("simple-cms: delete category failed: %v", err)
			c.JSON(http.StatusConflict, gin.H{"error": "カテゴリに紐づく記事が存在するため削除できません"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
