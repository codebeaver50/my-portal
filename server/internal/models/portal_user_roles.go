package models

import "time"

// UserRole: ユーザーとロールの割り当て（多対多の中間テーブル）
type UserRole struct {
	UserID    uint      `gorm:"column:user_id;primaryKey"`
	RoleID    uint      `gorm:"column:role_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// UserRole 構造体は portal_user_roles テーブルにマッピングされる
func (UserRole) TableName() string {
	return "portal_user_roles"
}
