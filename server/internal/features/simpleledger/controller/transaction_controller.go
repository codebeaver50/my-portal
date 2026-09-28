package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/dto"
	"github.com/codebeaver50/my-portal/server/internal/features/simpleledger/service"
)

// TransactionController は /api/simple-ledger/transactions のエンドポイントを実装する。
type TransactionController interface {
	List() gin.HandlerFunc
	ListPaginated() gin.HandlerFunc
	Create() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type transactionController struct {
	service service.TransactionService
}

// NewTransactionController は指定のserviceを使う TransactionController を作成する。
func NewTransactionController(service service.TransactionService) TransactionController {
	return &transactionController{service: service}
}

// List は GET /api/simple-ledger/transactions を処理する。
func (ctrl *transactionController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := ctrl.service.ListAll(c.Request.Context())
		if err != nil {
			log.Printf("simple-ledger: list transactions failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// ListPaginated は GET /api/simple-ledger/transactions/paginated を処理する。
func (ctrl *transactionController) ListPaginated() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		keyword := c.Query("keyword")

		result, err := ctrl.service.ListPaginated(c.Request.Context(), page, pageSize, keyword)
		if err != nil {
			log.Printf("simple-ledger: list paginated transactions failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Create は POST /api/simple-ledger/transactions を処理する。
func (ctrl *transactionController) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.TransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Create(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

// Update は PUT /api/simple-ledger/transactions/:id を処理する。
func (ctrl *transactionController) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
			return
		}

		var req dto.TransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ctrl.service.Update(c.Request.Context(), uint(id), &req)
		if errors.Is(err, service.ErrTransactionNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// Delete は DELETE /api/simple-ledger/transactions/:id を処理する。
func (ctrl *transactionController) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
			return
		}

		err = ctrl.service.Delete(c.Request.Context(), uint(id))
		if errors.Is(err, service.ErrTransactionNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("simple-ledger: delete transaction failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
