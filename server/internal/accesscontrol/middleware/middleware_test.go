package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/service"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// fakeAuthService は service.AuthService をDBアクセスなしで差し替える。
// トークンごとに返す Principal を決めておき、それ以外のトークンはゲストにする。
type fakeAuthService struct {
	principals map[string]*service.Principal
	guest      *service.Principal
	err        error
	gotToken   string
}

func (f *fakeAuthService) Login(ctx context.Context, req *dto.LoginRequest, meta service.LoginMeta) (string, *service.Principal, error) {
	return "", nil, errors.New("not implemented")
}

func (f *fakeAuthService) Logout(ctx context.Context, token string) error {
	return nil
}

func (f *fakeAuthService) Authenticate(ctx context.Context, token string) (*service.Principal, error) {
	f.gotToken = token
	if f.err != nil {
		return nil, f.err
	}
	if principal, ok := f.principals[token]; ok {
		return principal, nil
	}
	return f.guest, nil
}

func newEngine(auth service.AuthService) *gin.Engine {
	engine := gin.New()
	engine.Use(Authenticate(auth))
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	engine.GET("/ledger", Require(permission.SimpleLedger, permission.Read), ok)
	engine.POST("/ledger", Require(permission.SimpleLedger, permission.Write), ok)
	return engine
}

func request(engine *gin.Engine, method, token string) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/ledger", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	}
	engine.ServeHTTP(w, req)
	return w.Code
}

func TestRequire(t *testing.T) {
	auth := &fakeAuthService{
		guest: &service.Principal{Permissions: permission.Set{permission.SimpleLedger: permission.Read}},
		principals: map[string]*service.Principal{
			"viewer": {User: &models.User{ID: 1}, Permissions: permission.Set{permission.SimpleLedger: permission.Read}},
			"editor": {User: &models.User{ID: 2}, Permissions: permission.Set{permission.SimpleLedger: permission.Write}},
		},
	}
	engine := newEngine(auth)

	cases := []struct {
		name   string
		method string
		token  string
		want   int
	}{
		{"guest can read", http.MethodGet, "", http.StatusOK},
		{"guest is asked to log in to write", http.MethodPost, "", http.StatusUnauthorized},
		{"viewer is forbidden to write", http.MethodPost, "viewer", http.StatusForbidden},
		{"editor can write", http.MethodPost, "editor", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := request(engine, tc.method, tc.token); got != tc.want {
				t.Errorf("code = %d, want %d", got, tc.want)
			}
		})
	}

	t.Run("passes the cookie's token to the service", func(t *testing.T) {
		request(engine, http.MethodGet, "editor")
		if auth.gotToken != "editor" {
			t.Errorf("token = %q", auth.gotToken)
		}
	})
}

func TestAuthenticate_ServiceError(t *testing.T) {
	engine := newEngine(&fakeAuthService{err: errors.New("db down")})
	if got := request(engine, http.MethodGet, "editor"); got != http.StatusInternalServerError {
		t.Errorf("code = %d, want 500", got)
	}
}
