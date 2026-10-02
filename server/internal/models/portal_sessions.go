package models

import "time"

// Session: ログインセッション。Cookieのトークンそのものではなく、そのSHA-256を保存する
type Session struct {
	TokenHash string    `gorm:"column:token_hash;size:64;primaryKey"`
	UserID    uint      `gorm:"column:user_id;not null;index"`
	UserAgent string    `gorm:"column:user_agent;size:255;not null;default:''"`
	IPAddress string    `gorm:"column:ip_address;size:45;not null;default:''"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// Session 構造体は portal_sessions テーブルにマッピングされる
func (Session) TableName() string {
	return "portal_sessions"
}
