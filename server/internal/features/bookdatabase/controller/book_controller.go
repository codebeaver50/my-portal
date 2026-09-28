// Package controller は book database feature の /api/books エンドポイント
// のHTTPハンドラを実装する。
package controller

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/features/bookdatabase/service"
)

// BookController は GET /api/books とそのカバー画像プロキシを実装する。
type BookController interface {
	List() gin.HandlerFunc
	Cover() gin.HandlerFunc
}

type bookController struct {
	service service.BookService
}

// NewBookController は指定のserviceを使う BookController を作成する。
func NewBookController(service service.BookService) BookController {
	return &bookController{service: service}
}

// List は GET /api/books?skip=N を処理する。
func (ctrl *bookController) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		skip, _ := strconv.Atoi(c.Query("skip"))
		if skip < 0 {
			skip = 0
		}

		result, err := ctrl.service.List(c.Request.Context(), skip)
		if err != nil {
			log.Printf("books: list failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// Cover は GET /api/books/:id/cover を処理する。NDLのサムネイルサーバーは
// Refererが自ドメインでないリクエストを拒否するため、ブラウザから直接カバー
// 画像を読み込めない — そこでサーバー側でフェッチをプロキシする。
func (ctrl *bookController) Cover() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		coverImageURL, err := ctrl.service.GetCoverImageURL(c.Request.Context(), bookID)
		if err != nil {
			log.Printf("books: cover lookup failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		if coverImageURL == "" {
			c.Status(http.StatusNotFound)
			return
		}

		resp, err := ctrl.service.FetchCoverImage(coverImageURL)
		if err != nil {
			log.Printf("books: cover fetch failed: %v", err)
			c.Status(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.Status(http.StatusBadGateway)
			return
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		c.Header("Content-Type", contentType)
		c.Header("Cache-Control", "public, max-age=86400")
		c.Status(http.StatusOK)
		io.Copy(c.Writer, resp.Body)
	}
}
