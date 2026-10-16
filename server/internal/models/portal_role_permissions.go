package models

import "time"

// RolePermission: ロールが持つリソース（feature等）への権限レベル（read / write）
type RolePermission struct {
	RoleID    uint      `gorm:"column:role_id;primaryKey"`
	Resource  string    `gorm:"column:resource;size:50;primaryKey"`
	Level     string    `gorm:"column:level;size:10;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// RolePermission 構造体は portal_role_permissions テーブルにマッピングされる
func (RolePermission) TableName() string {
	return "portal_role_permissions"
}
