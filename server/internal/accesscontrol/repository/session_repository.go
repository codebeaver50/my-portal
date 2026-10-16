package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// SessionRepository は portal_sessions テーブルにアクセスする。
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository は指定のコネクションに紐づく SessionRepository を作成する。
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create は新しいセッションの行を挿入する。
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByTokenHash はトークンのハッシュでセッションを返す。存在しない場合は nil を返す（期限切れかどうかは見ない）。
func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteByTokenHash はトークンのハッシュでセッションを削除する。存在しなくてもエラーにしない。
func (r *SessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&models.Session{}).Error
}

// DeleteExpiredByUserID は指定のユーザーの、before 時点で期限切れのセッションを削除する。
func (r *SessionRepository) DeleteExpiredByUserID(ctx context.Context, userID uint, before time.Time) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND expires_at <= ?", userID, before).Delete(&models.Session{}).Error
}
