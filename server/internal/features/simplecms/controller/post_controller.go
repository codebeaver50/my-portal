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

// PostController は /api/simple-cms/posts のエンドポイントを実装する。
type PostController interface {
	List() gin.HandlerFunc
	GetBySlug() gin.HandlerFunc
	Create() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type postController struct {
	service service.PostService
}

// NewPostController は指定のserviceを使う PostController を作成する。
func NewPostController(service service.PostService) PostController {
	return &postController{service: service}
}

// List は GET /api/simple-cms/posts を処理する。
func (ctrl *postController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.ListAll(c.Request.Context())
		if err != nil {
			log.Printf("simple-cms: list posts failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// GetBySlug は GET /api/simple-cms/posts/:slug を処理する。
func (ctrl *postController) GetBySlug() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.GetBySlug(c.Request.Context(), c.Param("slug"))
		if errors.Is(err, service.ErrPostNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("simple-cms: get post failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Create は POST /api/simple-cms/posts を処理する。
func (ctrl *postController) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.PostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Create(c.Request.Context(), &req)
		if errors.Is(err, service.ErrPostSlugTaken) {
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

// Update は PUT /api/simple-cms/posts/:id を処理する。
func (ctrl *postController) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
			return
		}

		var req dto.PostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Update(c.Request.Context(), uint(id), &req)
		if errors.Is(err, service.ErrPostNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrPostSlugTaken) {
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

// Delete は DELETE /api/simple-cms/posts/:id を処理する。
func (ctrl *postController) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
			return
		}

		err = ctrl.service.Delete(c.Request.Context(), uint(id))
		if errors.Is(err, service.ErrPostNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("simple-cms: delete post failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
