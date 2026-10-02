package models

import "time"

// User: ポータル共通のアクセス管理でログインするユーザー
type User struct {
	ID               uint       `gorm:"column:id;primaryKey"`
	LoginID          string     `gorm:"column:login_id;size:50;uniqueIndex;not null"`
	DisplayName      string     `gorm:"column:display_name;size:100;not null"`
	PasswordHash     string     `gorm:"column:password_hash;size:255;not null" json:"-"` // 誤ってレスポンスに含めないよう、JSONには出さない
	IsActive         bool       `gorm:"column:is_active;not null;default:true"`
	FailedLoginCount int        `gorm:"column:failed_login_count;not null;default:0"`
	LockedUntil      *time.Time `gorm:"column:locked_until"`
	LastLoginAt      *time.Time `gorm:"column:last_login_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

// User 構造体は portal_users テーブルにマッピングされる
func (User) TableName() string {
	return "portal_users"
}
