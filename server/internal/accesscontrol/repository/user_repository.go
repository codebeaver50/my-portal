// Package repository はポータル共通のアクセス管理のDBアクセス（読み書き）を実装する。
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// UserRepository は portal_users テーブルにアクセスする。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository は指定のコネクションに紐づく UserRepository を作成する。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく UserRepository を返す。
func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return &UserRepository{db: tx}
}

// Create はユーザーを挿入する。挿入後、user.ID に採番されたIDが入る。
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID はIDでユーザーを返す。存在しない場合は nil を返す。
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByLoginID はログインIDでユーザーを返す。存在しない場合は nil を返す。
func (r *UserRepository) GetByLoginID(ctx context.Context, loginID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("login_id = ?", loginID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateLoginState はログイン試行の結果（連続失敗回数・ロック期限・最終ログイン日時）を書き込む。
func (r *UserRepository) UpdateLoginState(ctx context.Context, id uint, failedLoginCount int, lockedUntil, lastLoginAt *time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(map[string]any{
		"failed_login_count": failedLoginCount,
		"locked_until":       lockedUntil,
		"last_login_at":      lastLoginAt,
	}).Error
}
