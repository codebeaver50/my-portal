// Package service はポータル共通のアクセス管理（ログイン・セッション・権限の計算）を実装する。
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

const (
	// SessionTTL はログインしてからセッションが失効するまでの期間。
	SessionTTL = 7 * 24 * time.Hour
	// MaxFailedLogins 回連続でパスワードを間違えると、LockDuration のあいだログインできなくする。
	MaxFailedLogins = 5
	LockDuration    = 15 * time.Minute

	// GuestRoleCode は未ログインの訪問者に適用するロールのcode。
	GuestRoleCode = "guest"
)

// ErrInvalidCredentials はログインIDまたはパスワードが違う（もしくはユーザーが無効化されている）場合に返される。
// どれに当たるかを区別して返すと、ログインIDの存在を推測されてしまうため、まとめて1つのエラーにする。
var ErrInvalidCredentials = errors.New("invalid login id or password")

// ErrAccountLocked はログインの連続失敗によりアカウントがロック中の場合に返される。
var ErrAccountLocked = errors.New("account is temporarily locked")

// dummyPasswordHash は存在しないログインIDでもbcryptの比較を1回行い、応答時間からログインIDの存在を推測されないようにするためのハッシュ。
var dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), PasswordHashCost)

// PasswordHashCost はパスワードをハッシュ化するときのbcryptのコスト。
const PasswordHashCost = 12

// LoginMeta はセッションに記録するクライアントの情報。
type LoginMeta struct {
	UserAgent string
	IPAddress string
}

// Principal はリクエストを行っている主体。未ログインの場合は User が nil で、Permissions はゲストロールの権限。
type Principal struct {
	User        *models.User
	Roles       []models.Role
	Permissions permission.Set
}

// AuthService はログイン・ログアウトと、セッショントークンから Principal を解決する処理を実装する。
type AuthService interface {
	// Login はログインIDとパスワードを検証し、新しいセッションのトークン（Cookieに入れる値）を返す。
	Login(ctx context.Context, req *dto.LoginRequest, meta LoginMeta) (token string, principal *Principal, err error)
	// Logout はトークンに対応するセッションを削除する。
	Logout(ctx context.Context, token string) error
	// Authenticate はトークンから Principal を返す。トークンが空・無効・期限切れの場合はゲストの Principal を返す。
	Authenticate(ctx context.Context, token string) (*Principal, error)
}

type authService struct {
	users    *repository.UserRepository
	roles    *repository.RoleRepository
	sessions *repository.SessionRepository
	now      func() time.Time
}

// NewAuthService は指定のrepositoryを使う AuthService を作成する。
func NewAuthService(users *repository.UserRepository, roles *repository.RoleRepository, sessions *repository.SessionRepository) AuthService {
	return newAuthService(users, roles, sessions, time.Now)
}

// newAuthService はテストで現在時刻を差し替えられるようにしたコンストラクタ。
func newAuthService(users *repository.UserRepository, roles *repository.RoleRepository, sessions *repository.SessionRepository, now func() time.Time) *authService {
	return &authService{users: users, roles: roles, sessions: sessions, now: now}
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest, meta LoginMeta) (string, *Principal, error) {
	now := s.now()

	user, err := s.users.GetByLoginID(ctx, req.LoginID)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(req.Password))
		return "", nil, ErrInvalidCredentials
	}
	if user.LockedUntil != nil && now.Before(*user.LockedUntil) {
		return "", nil, ErrAccountLocked
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		if err := s.recordFailedLogin(ctx, user, now); err != nil {
			return "", nil, err
		}
		return "", nil, ErrInvalidCredentials
	}
	// パスワードが正しくても、無効化されたユーザーは存在しない場合と同じエラーにする
	if !user.IsActive {
		return "", nil, ErrInvalidCredentials
	}

	if err := s.users.UpdateLoginState(ctx, user.ID, 0, nil, &now); err != nil {
		return "", nil, err
	}
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	user.LastLoginAt = &now

	// ログインのたびに、そのユーザーの期限切れのセッションを掃除しておく
	if err := s.sessions.DeleteExpiredByUserID(ctx, user.ID, now); err != nil {
		return "", nil, err
	}

	token, err := newToken()
	if err != nil {
		return "", nil, err
	}
	session := models.Session{
		TokenHash: hashToken(token),
		UserID:    user.ID,
		UserAgent: truncate(meta.UserAgent, 255),
		IPAddress: truncate(meta.IPAddress, 45),
		ExpiresAt: now.Add(SessionTTL),
	}
	if err := s.sessions.Create(ctx, &session); err != nil {
		return "", nil, err
	}

	principal, err := s.principalFor(ctx, user)
	if err != nil {
		return "", nil, err
	}
	return token, principal, nil
}

// recordFailedLogin は連続失敗回数を1つ増やし、上限に達したらアカウントをロックする。
// ロック期限が過ぎたあとの失敗は、回数を1から数え直す。
func (s *authService) recordFailedLogin(ctx context.Context, user *models.User, now time.Time) error {
	count := user.FailedLoginCount + 1
	if user.LockedUntil != nil {
		count = 1
	}

	var lockedUntil *time.Time
	if count >= MaxFailedLogins {
		until := now.Add(LockDuration)
		lockedUntil = &until
	}
	return s.users.UpdateLoginState(ctx, user.ID, count, lockedUntil, user.LastLoginAt)
}

func (s *authService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.DeleteByTokenHash(ctx, hashToken(token))
}

func (s *authService) Authenticate(ctx context.Context, token string) (*Principal, error) {
	if token == "" {
		return s.principalFor(ctx, nil)
	}

	session, err := s.sessions.GetByTokenHash(ctx, hashToken(token))
	if err != nil {
		return nil, err
	}
	if session == nil || !s.now().Before(session.ExpiresAt) {
		return s.principalFor(ctx, nil)
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	// ログイン後に無効化されたユーザーは、セッションが残っていてもゲストとして扱う
	if user == nil || !user.IsActive {
		return s.principalFor(ctx, nil)
	}
	return s.principalFor(ctx, user)
}

// principalFor はユーザー（nilならゲスト）の Principal を組み立てる。
// 権限は、割り当てられたロールとゲストロールの権限の和集合（リソースごとに最も強いもの）とする。
// ゲストロールも含めるのは、ログインしたことで未ログインのときより権限が弱くなることがないようにするため。
func (s *authService) principalFor(ctx context.Context, user *models.User) (*Principal, error) {
	var roles []models.Role
	if user != nil {
		var err error
		roles, err = s.roles.GetByUserID(ctx, user.ID)
		if err != nil {
			return nil, err
		}
	}

	roleIDs := make([]uint, 0, len(roles)+1)
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	guest, err := s.roles.GetByCode(ctx, GuestRoleCode)
	if err != nil {
		return nil, err
	}
	if guest != nil {
		roleIDs = append(roleIDs, guest.ID)
	}

	rows, err := s.roles.GetPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	permissions := permission.Set{}
	for _, row := range rows {
		permissions.Grant(permission.Resource(row.Resource), permission.ParseLevel(row.Level))
	}

	return &Principal{User: user, Roles: roles, Permissions: permissions}, nil
}

// ToSessionResponse は Principal をAPIのレスポンスに変換する。permissions には全リソースのキーを含める。
func ToSessionResponse(principal *Principal) *dto.SessionResponse {
	permissions := make(map[string]string, len(permission.Resources))
	for _, def := range permission.Resources {
		permissions[string(def.Resource)] = principal.Permissions.Level(def.Resource).String()
	}

	response := &dto.SessionResponse{Permissions: permissions}
	if principal.User != nil {
		roles := make([]dto.RoleSummary, len(principal.Roles))
		for i, role := range principal.Roles {
			roles[i] = dto.RoleSummary{Code: role.Code, Name: role.Name}
		}
		response.User = &dto.SessionUser{
			ID:          principal.User.ID,
			LoginID:     principal.User.LoginID,
			DisplayName: principal.User.DisplayName,
			Roles:       roles,
		}
	}
	return response
}

// newToken はセッショントークン（256ビットの乱数をURLセーフなbase64にしたもの）を生成する。
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken はDBに保存・照合するためのトークンのSHA-256（16進）を返す。
// トークン自体が十分な長さの乱数なので、パスワードと違いソルトやストレッチングは不要。
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func truncate(value string, maxLen int) string {
	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}
	return string(runes[:maxLen])
}
