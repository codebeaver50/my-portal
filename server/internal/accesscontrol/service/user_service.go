package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

const (
	// MinPasswordLength はユーザー作成時に求めるパスワードの最小文字数。
	MinPasswordLength = 12
	// MaxPasswordBytes はパスワードの最大バイト数。bcrypt は72バイトを超える部分を無視するため、それ以上は受け付けない。
	MaxPasswordBytes = 72
)

// loginIDPattern はログインIDに使える文字（英数字と . _ -）と長さ（3〜50文字）。
var loginIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{3,50}$`)

// UserService はユーザーの作成を実装する。
type UserService interface {
	// Create はユーザーを作成し、RoleCode のロールを割り当てる。パスワードはbcryptでハッシュ化して保存する。
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
}

type userService struct {
	db    *gorm.DB
	users *repository.UserRepository
	roles *repository.RoleRepository
}

// NewUserService は指定のrepositoryを使う UserService を作成する。
func NewUserService(db *gorm.DB, users *repository.UserRepository, roles *repository.RoleRepository) UserService {
	return &userService{db: db, users: users, roles: roles}
}

func (s *userService) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	loginID := strings.TrimSpace(req.LoginID)
	displayName := strings.TrimSpace(req.DisplayName)
	if err := validateNewUser(loginID, displayName, req.Password); err != nil {
		return nil, err
	}

	role, err := s.roles.GetByCode(ctx, req.RoleCode)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, &ValidationError{Message: fmt.Sprintf("ロールが存在しません: %s", req.RoleCode)}
	}
	// ゲストロールは全員に常に適用されるため、ユーザーに割り当てても意味がない
	if role.Code == GuestRoleCode {
		return nil, &ValidationError{Message: "ゲストロールはユーザーに割り当てられません"}
	}

	existing, err := s.users.GetByLoginID(ctx, loginID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, &ValidationError{Message: fmt.Sprintf("ログインIDはすでに使われています: %s", loginID)}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), PasswordHashCost)
	if err != nil {
		return nil, err
	}
	user := models.User{LoginID: loginID, DisplayName: displayName, PasswordHash: string(hash), IsActive: true}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.users.WithTx(tx).Create(ctx, &user); err != nil {
			return err
		}
		return s.roles.WithTx(tx).CreateUserRole(ctx, &models.UserRole{UserID: user.ID, RoleID: role.ID})
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{ID: user.ID, LoginID: user.LoginID, DisplayName: user.DisplayName, RoleCode: role.Code}, nil
}

func validateNewUser(loginID, displayName, password string) error {
	if !loginIDPattern.MatchString(loginID) {
		return &ValidationError{Message: "ログインIDは英数字と . _ - の3〜50文字にしてください"}
	}
	if displayName == "" || utf8.RuneCountInString(displayName) > 100 {
		return &ValidationError{Message: "表示名は1〜100文字にしてください"}
	}
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return &ValidationError{Message: fmt.Sprintf("パスワードは%d文字以上にしてください", MinPasswordLength)}
	}
	if len(password) > MaxPasswordBytes {
		return &ValidationError{Message: fmt.Sprintf("パスワードは%dバイト以内にしてください", MaxPasswordBytes)}
	}
	return nil
}
