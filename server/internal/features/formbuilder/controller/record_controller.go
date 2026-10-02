package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/formbuilder/service"
)

// RecordController は /api/form-builder/forms/:id/records のエンドポイントを実装する。
type RecordController interface {
	List() gin.HandlerFunc
	Create() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type recordController struct {
	service service.RecordService
}

// NewRecordController は指定のserviceを使う RecordController を作成する。
func NewRecordController(service service.RecordService) RecordController {
	return &recordController{service: service}
}

// List は GET /api/form-builder/forms/:id/records を処理する。
func (ctrl *recordController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		formID, ok := parseID(c, "id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form id"})
			return
		}
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
			return
		}
		pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
		if err != nil || pageSize < 1 || pageSize > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pageSize"})
			return
		}

		result, err := ctrl.service.ListPaginated(c.Request.Context(), formID, page, pageSize)
		if errors.Is(err, service.ErrFormNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("form-builder: list records failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Create は POST /api/form-builder/forms/:id/records を処理する。
func (ctrl *recordController) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		formID, ok := parseID(c, "id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form id"})
			return
		}

		var req dto.RecordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Create(c.Request.Context(), formID, &req)
		if errors.Is(err, service.ErrFormNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if writeValidationError(c, err) {
			return
		}
		if err != nil {
			log.Printf("form-builder: create record failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

// Delete は DELETE /api/form-builder/forms/:id/records/:recordId を処理する。
func (ctrl *recordController) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		formID, ok := parseID(c, "id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form id"})
			return
		}
		recordID, ok := parseID(c, "recordId")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
			return
		}

		err := ctrl.service.Delete(c.Request.Context(), formID, recordID)
		if errors.Is(err, service.ErrRecordNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("form-builder: delete record failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
