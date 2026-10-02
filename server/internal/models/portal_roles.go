package models

import "time"

// Role: ユーザーに割り当てるロール。権限は RolePermission で持つ
type Role struct {
	ID          uint      `gorm:"column:id;primaryKey"`
	Code        string    `gorm:"column:code;size:50;uniqueIndex;not null"`
	Name        string    `gorm:"column:name;size:100;not null"`
	Description string    `gorm:"column:description;size:255;not null;default:''"`
	IsSystem    bool      `gorm:"column:is_system;not null;default:false"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// Role 構造体は portal_roles テーブルにマッピングされる
func (Role) TableName() string {
	return "portal_roles"
}
