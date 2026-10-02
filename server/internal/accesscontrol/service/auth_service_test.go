package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// newTestDB はアクセス管理のテーブルをマイグレーション済みの、テスト専用のインメモリSQLite DBを返す。
// 単一コネクションに固定し、複数コネクションがそれぞれ別の空DBを見てしまうのを防ぐ。本番のMySQLスキーマ管理
// （golang-migrate）とは無関係の、テスト内でのみ使い捨てるDBであるため AutoMigrate を使う。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Session{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

const testPassword = "correct-password"

// fixture はテスト用のロール（guest / editor / viewer）とユーザーを作り、現在時刻を進められる authService を返す。
type fixture struct {
	db  *gorm.DB
	svc *authService
	now time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := newTestDB(t)
	f := &fixture{db: db, now: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	f.svc = newAuthService(
		repository.NewUserRepository(db),
		repository.NewRoleRepository(db),
		repository.NewSessionRepository(db),
		func() time.Time { return f.now },
	)

	roles := []models.Role{
		{ID: 1, Code: GuestRoleCode, Name: "ゲスト"},
		{ID: 2, Code: "editor", Name: "編集者"},
		{ID: 3, Code: "cms-viewer", Name: "CMS閲覧者"},
	}
	mustCreate(t, db, &roles)
	mustCreate(t, db, &[]models.RolePermission{
		{RoleID: 1, Resource: "book-database", Level: "read"},
		{RoleID: 2, Resource: "simple-ledger", Level: "write"},
		{RoleID: 3, Resource: "simple-cms", Level: "read"},
		{RoleID: 3, Resource: "simple-ledger", Level: "read"},
	})
	return f
}

func mustCreate(t *testing.T, db *gorm.DB, value any) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatalf("failed to create fixture: %v", err)
	}
}

// createUser はパスワードが testPassword のユーザーを作り、指定のロールを割り当てる。
func (f *fixture) createUser(t *testing.T, loginID string, roleIDs ...uint) *models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{LoginID: loginID, DisplayName: loginID, PasswordHash: string(hash), IsActive: true}
	mustCreate(t, f.db, &user)
	for _, roleID := range roleIDs {
		mustCreate(t, f.db, &models.UserRole{UserID: user.ID, RoleID: roleID})
	}
	return &user
}

func (f *fixture) login(loginID, password string) (string, *Principal, error) {
	return f.svc.Login(context.Background(), &dto.LoginRequest{LoginID: loginID, Password: password}, LoginMeta{UserAgent: "test", IPAddress: "127.0.0.1"})
}

func (f *fixture) reloadUser(t *testing.T, id uint) models.User {
	t.Helper()
	var user models.User
	if err := f.db.First(&user, id).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func TestAuthService_Login(t *testing.T) {
	t.Run("issues a session whose token resolves to the user", func(t *testing.T) {
		f := newFixture(t)
		user := f.createUser(t, "alice", 2)

		token, principal, err := f.login("alice", testPassword)
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if token == "" || principal.User == nil || principal.User.ID != user.ID {
			t.Fatalf("token = %q, principal = %+v", token, principal)
		}

		var session models.Session
		if err := f.db.First(&session).Error; err != nil {
			t.Fatal(err)
		}
		if session.TokenHash == token || session.TokenHash != hashToken(token) {
			t.Error("the session must store the token's hash, not the token itself")
		}
		if !session.ExpiresAt.Equal(f.now.Add(SessionTTL)) {
			t.Errorf("expiresAt = %v", session.ExpiresAt)
		}

		resolved, err := f.svc.Authenticate(context.Background(), token)
		if err != nil || resolved.User == nil || resolved.User.ID != user.ID {
			t.Errorf("authenticate: %+v, %v", resolved, err)
		}
		if got := f.reloadUser(t, user.ID); got.LastLoginAt == nil {
			t.Error("lastLoginAt was not recorded")
		}
	})

	t.Run("rejects an unknown login id and a wrong password with the same error", func(t *testing.T) {
		f := newFixture(t)
		f.createUser(t, "alice")

		if _, _, err := f.login("nobody", testPassword); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("unknown login id: err = %v", err)
		}
		if _, _, err := f.login("alice", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("wrong password: err = %v", err)
		}
	})

	t.Run("rejects a deactivated user even with the correct password", func(t *testing.T) {
		f := newFixture(t)
		user := f.createUser(t, "alice")
		f.db.Model(user).Update("is_active", false)

		if _, _, err := f.login("alice", testPassword); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("locks the account after consecutive failures until the lock expires", func(t *testing.T) {
		f := newFixture(t)
		user := f.createUser(t, "alice")

		for i := 0; i < MaxFailedLogins; i++ {
			if _, _, err := f.login("alice", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("attempt %d: err = %v", i+1, err)
			}
		}
		if _, _, err := f.login("alice", testPassword); !errors.Is(err, ErrAccountLocked) {
			t.Fatalf("while locked: err = %v", err)
		}

		f.now = f.now.Add(LockDuration)
		if _, _, err := f.login("alice", testPassword); err != nil {
			t.Fatalf("after the lock expired: err = %v", err)
		}
		if got := f.reloadUser(t, user.ID); got.FailedLoginCount != 0 || got.LockedUntil != nil {
			t.Errorf("login state was not reset: %+v", got)
		}
	})

	t.Run("a successful login resets the failure count", func(t *testing.T) {
		f := newFixture(t)
		user := f.createUser(t, "alice")

		for i := 0; i < MaxFailedLogins-1; i++ {
			f.login("alice", "wrong")
		}
		if _, _, err := f.login("alice", testPassword); err != nil {
			t.Fatalf("err = %v", err)
		}
		if got := f.reloadUser(t, user.ID); got.FailedLoginCount != 0 {
			t.Errorf("failedLoginCount = %d", got.FailedLoginCount)
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	f := newFixture(t)
	f.createUser(t, "alice", 2)
	token, _, err := f.login("alice", testPassword)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.svc.Logout(context.Background(), token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	principal, err := f.svc.Authenticate(context.Background(), token)
	if err != nil || principal.User != nil {
		t.Errorf("the token should no longer resolve to the user: %+v, %v", principal, err)
	}
}

func TestAuthService_Authenticate(t *testing.T) {
	t.Run("resolves an empty or unknown token to the guest", func(t *testing.T) {
		f := newFixture(t)

		for _, token := range []string{"", "unknown-token"} {
			principal, err := f.svc.Authenticate(context.Background(), token)
			if err != nil || principal.User != nil {
				t.Fatalf("token %q: %+v, %v", token, principal, err)
			}
			if principal.Permissions.Level(permission.BookDatabase) != permission.Read {
				t.Errorf("token %q: the guest role's permissions were not applied: %v", token, principal.Permissions)
			}
		}
	})

	t.Run("resolves an expired session to the guest", func(t *testing.T) {
		f := newFixture(t)
		f.createUser(t, "alice", 2)
		token, _, _ := f.login("alice", testPassword)

		f.now = f.now.Add(SessionTTL)
		principal, err := f.svc.Authenticate(context.Background(), token)
		if err != nil || principal.User != nil {
			t.Errorf("%+v, %v", principal, err)
		}
	})

	t.Run("resolves a user deactivated after login to the guest", func(t *testing.T) {
		f := newFixture(t)
		user := f.createUser(t, "alice", 2)
		token, _, _ := f.login("alice", testPassword)
		f.db.Model(user).Update("is_active", false)

		principal, err := f.svc.Authenticate(context.Background(), token)
		if err != nil || principal.User != nil {
			t.Errorf("%+v, %v", principal, err)
		}
	})

	t.Run("grants the strongest level across the user's roles and the guest role", func(t *testing.T) {
		f := newFixture(t)
		f.createUser(t, "alice", 2, 3)
		token, _, _ := f.login("alice", testPassword)

		principal, err := f.svc.Authenticate(context.Background(), token)
		if err != nil {
			t.Fatal(err)
		}
		want := permission.Set{
			permission.BookDatabase: permission.Read,  // ゲストロール
			permission.SimpleLedger: permission.Write, // editor(write) と cms-viewer(read) のうち強い方
			permission.SimpleCMS:    permission.Read,  // cms-viewer
		}
		for resource, level := range want {
			if got := principal.Permissions.Level(resource); got != level {
				t.Errorf("%s = %v, want %v", resource, got, level)
			}
		}
		if got := principal.Permissions.Level(permission.AccessControl); got != permission.None {
			t.Errorf("access-control = %v, want none", got)
		}
	})
}

func TestToSessionResponse(t *testing.T) {
	principal := &Principal{
		User:        &models.User{ID: 1, LoginID: "alice", DisplayName: "Alice"},
		Roles:       []models.Role{{Code: "editor", Name: "編集者"}},
		Permissions: permission.Set{permission.SimpleLedger: permission.Write},
	}
	response := ToSessionResponse(principal)

	if len(response.Permissions) != len(permission.Resources) {
		t.Errorf("permissions should include every resource: %v", response.Permissions)
	}
	if response.Permissions["simple-ledger"] != "write" || response.Permissions["book-database"] != "none" {
		t.Errorf("permissions = %v", response.Permissions)
	}
	if response.User == nil || response.User.LoginID != "alice" || len(response.User.Roles) != 1 {
		t.Errorf("user = %+v", response.User)
	}
	if guest := ToSessionResponse(&Principal{Permissions: permission.Set{}}); guest.User != nil {
		t.Errorf("guest user = %+v, want nil", guest.User)
	}
}
