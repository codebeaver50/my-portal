package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/models"
)

// RoleRepository は portal_roles / portal_user_roles / portal_role_permissions テーブルにアクセスする。
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository は指定のコネクションに紐づく RoleRepository を作成する。
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// WithTx は指定のトランザクションに紐づく RoleRepository を返す。
func (r *RoleRepository) WithTx(tx *gorm.DB) *RoleRepository {
	return &RoleRepository{db: tx}
}

// GetAll は全ロールをID順で返す。
func (r *RoleRepository) GetAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Order("id ASC").Find(&roles).Error
	return roles, err
}

// GetByID はIDでロールを返す。存在しない場合は nil を返す。
func (r *RoleRepository) GetByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByCode はcodeでロールを返す。存在しない場合は nil を返す。
func (r *RoleRepository) GetByCode(ctx context.Context, code string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByUserID はユーザーに割り当てられたロールをID順で返す。
func (r *RoleRepository) GetByUserID(ctx context.Context, userID uint) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN portal_user_roles ON portal_user_roles.role_id = portal_roles.id").
		Where("portal_user_roles.user_id = ?", userID).
		Order("portal_roles.id ASC").
		Find(&roles).Error
	return roles, err
}

// GetPermissionsByRoleIDs は指定のロールすべての権限を返す。
func (r *RoleRepository) GetPermissionsByRoleIDs(ctx context.Context, roleIDs []uint) ([]models.RolePermission, error) {
	var permissions []models.RolePermission
	if len(roleIDs) == 0 {
		return permissions, nil
	}
	err := r.db.WithContext(ctx).Where("role_id IN ?", roleIDs).Find(&permissions).Error
	return permissions, err
}

// DeletePermissionsByRoleID は指定のロールの権限をすべて削除する。
func (r *RoleRepository) DeletePermissionsByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
}

// CreatePermissions は複数の権限の行をまとめて挿入する。
func (r *RoleRepository) CreatePermissions(ctx context.Context, permissions []models.RolePermission) error {
	if len(permissions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&permissions).Error
}
