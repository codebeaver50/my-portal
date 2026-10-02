package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/service"
)

// FormController は /api/form-builder/forms のエンドポイントを実装する。
type FormController interface {
	List() gin.HandlerFunc
	Get() gin.HandlerFunc
	Create() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type formController struct {
	service service.FormService
}

// NewFormController は指定のserviceを使う FormController を作成する。
func NewFormController(service service.FormService) FormController {
	return &formController{service: service}
}

// List は GET /api/form-builder/forms を処理する。
func (ctrl *formController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.ListAll(c.Request.Context())
		if err != nil {
			log.Printf("form-builder: list forms failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Get は GET /api/form-builder/forms/:id を処理する。
func (ctrl *formController) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c, "id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form id"})
			return
		}

		result, err := ctrl.service.Get(c.Request.Context(), id)
		if errors.Is(err, service.ErrFormNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("form-builder: get form failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Create は POST /api/form-builder/forms を処理する。
func (ctrl *formController) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.FormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Create(c.Request.Context(), &req)
		if writeValidationError(c, err) {
			return
		}
		if err != nil {
			log.Printf("form-builder: create form failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

// Update は PUT /api/form-builder/forms/:id を処理する。
func (ctrl *formController) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c, "id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form id"})
			return
		}

		var req dto.FormRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Update(c.Request.Context(), id, &req)
		if errors.Is(err, service.ErrFormNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if writeValidationError(c, err) {
			return
		}
		if err != nil {
			log.Printf("form-builder: update form failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Delete は DELETE /api/form-builder/forms/:id を処理する。
func (ctrl *formController) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c, "id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form id"})
			return
		}

		err := ctrl.service.Delete(c.Request.Context(), id)
		if errors.Is(err, service.ErrFormNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("form-builder: delete form failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
