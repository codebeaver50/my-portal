package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestIsReadOnly(t *testing.T) {
	cases := map[string]bool{
		"true": true, "TRUE": true, "1": true,
		"": false, "false": false, "0": false, "yes": false,
	}
	for value, want := range cases {
		if got := isReadOnly(value); got != want {
			t.Errorf("isReadOnly(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestReadOnlyMiddleware(t *testing.T) {
	// /x はユーザーの操作から呼ばれる通常の書き込みのルート、/api/books/sync は
	// main() で除外している cron 専用の同期処理のルートを表す。
	newEngine := func(enabled bool, exemptPaths ...string) *gin.Engine {
		r := gin.New()
		r.Use(readOnlyMiddleware(enabled, exemptPaths...))
		ok := func(c *gin.Context) { c.Status(http.StatusOK) }
		r.Any("/x", ok)
		r.Any("/api/books/sync", ok)
		return r
	}

	t.Run("disabled passes every method through", func(t *testing.T) {
		r := newEngine(false)
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(method, "/x", nil))
			if w.Code != http.StatusOK {
				t.Errorf("method %s: status = %d, want 200", method, w.Code)
			}
		}
	})

	t.Run("enabled blocks write methods with 403", func(t *testing.T) {
		r := newEngine(true)
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(method, "/x", nil))
			if w.Code != http.StatusForbidden {
				t.Errorf("method %s: status = %d, want 403", method, w.Code)
			}
		}
	})

	t.Run("enabled still allows GET/HEAD/OPTIONS", func(t *testing.T) {
		r := newEngine(true)
		for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(method, "/x", nil))
			if w.Code != http.StatusOK {
				t.Errorf("method %s: status = %d, want 200", method, w.Code)
			}
		}
	})

	t.Run("enabled still blocks writes on a non-exempt path even with an exemption list set", func(t *testing.T) {
		r := newEngine(true, "/api/books/sync")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/x", nil))
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
	})

	t.Run("enabled allows writes on an exempt path", func(t *testing.T) {
		r := newEngine(true, "/api/books/sync")
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(method, "/api/books/sync", nil))
			if w.Code != http.StatusOK {
				t.Errorf("method %s: status = %d, want 200", method, w.Code)
			}
		}
	})
}
