package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

const newUserPassword = "long-enough-password"

// newUserFixture は admin / guest のロールを作り、UserService を返す。
func newUserFixture(t *testing.T) (*gorm.DB, UserService) {
	t.Helper()
	db := newTestDB(t)
	mustCreate(t, db, &[]models.Role{
		{ID: 1, Code: AdminRoleCode, Name: "管理者", IsSystem: true},
		{ID: 2, Code: GuestRoleCode, Name: "ゲスト", IsSystem: true},
	})
	return db, NewUserService(db, repository.NewUserRepository(db), repository.NewRoleRepository(db))
}

func validCreateUserRequest() *dto.CreateUserRequest {
	return &dto.CreateUserRequest{LoginID: "owner", DisplayName: "サイト管理者", Password: newUserPassword, RoleCode: AdminRoleCode}
}

func TestUserService_Create(t *testing.T) {
	t.Run("creates a user with a hashed password and assigns the role", func(t *testing.T) {
		db, svc := newUserFixture(t)

		created, err := svc.Create(context.Background(), validCreateUserRequest())
		if err != nil {
			t.Fatal(err)
		}
		if created.LoginID != "owner" || created.RoleCode != AdminRoleCode {
			t.Errorf("created = %+v", created)
		}

		var user models.User
		if err := db.First(&user, created.ID).Error; err != nil {
			t.Fatal(err)
		}
		if user.PasswordHash == newUserPassword {
			t.Fatal("password is stored in plain text")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newUserPassword)); err != nil {
			t.Errorf("password hash does not match: %v", err)
		}
		if !user.IsActive {
			t.Error("user should be active")
		}
		var userRoles []models.UserRole
		if err := db.Where("user_id = ?", user.ID).Find(&userRoles).Error; err != nil {
			t.Fatal(err)
		}
		if len(userRoles) != 1 || userRoles[0].RoleID != 1 {
			t.Errorf("user roles = %+v", userRoles)
		}
	})

	t.Run("rejects a duplicate login id", func(t *testing.T) {
		_, svc := newUserFixture(t)
		if _, err := svc.Create(context.Background(), validCreateUserRequest()); err != nil {
			t.Fatal(err)
		}

		_, err := svc.Create(context.Background(), validCreateUserRequest())
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("err = %v, want ValidationError", err)
		}
	})

	invalid := []struct {
		name   string
		modify func(req *dto.CreateUserRequest)
	}{
		{"login id with invalid characters", func(req *dto.CreateUserRequest) { req.LoginID = "own er" }},
		{"too short login id", func(req *dto.CreateUserRequest) { req.LoginID = "ow" }},
		{"empty display name", func(req *dto.CreateUserRequest) { req.DisplayName = "  " }},
		{"too short password", func(req *dto.CreateUserRequest) { req.Password = "short-pass" }},
		{"password over 72 bytes", func(req *dto.CreateUserRequest) { req.Password = strings.Repeat("a", 73) }},
		{"unknown role", func(req *dto.CreateUserRequest) { req.RoleCode = "unknown" }},
		{"guest role", func(req *dto.CreateUserRequest) { req.RoleCode = GuestRoleCode }},
	}
	for _, tc := range invalid {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			db, svc := newUserFixture(t)
			req := validCreateUserRequest()
			tc.modify(req)

			_, err := svc.Create(context.Background(), req)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("err = %v, want ValidationError", err)
			}
			var count int64
			if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Errorf("users = %d, want 0", count)
			}
		})
	}
}
