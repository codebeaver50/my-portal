package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/permission"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// AdminRoleCode は全権限を持つ管理者ロールのcode。
// 権限を外して誰もアクセス管理画面を操作できなくなることがないよう、権限は変更できない。
const AdminRoleCode = "admin"

// ErrRoleNotFound は指定のロールが存在しない場合に返される。
var ErrRoleNotFound = errors.New("role not found")

// ValidationError はリクエストの内容が不正な場合に返される。Message はそのままAPIのエラーメッセージになる。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// RoleService はロールの一覧と、ロールごとの権限の変更を実装する。
type RoleService interface {
	// List は全ロールと、それぞれのリソースごとの権限を返す。
	List(ctx context.Context) (*dto.ListRolesResponse, error)
	// UpdatePermissions はロールの権限をリクエストの内容でまるごと置き換える。
	UpdatePermissions(ctx context.Context, roleID uint, req *dto.UpdateRolePermissionsRequest) (*dto.RoleResponse, error)
}

type roleService struct {
	db    *gorm.DB
	roles *repository.RoleRepository
}

// NewRoleService は指定のrepositoryを使う RoleService を作成する。
func NewRoleService(db *gorm.DB, roles *repository.RoleRepository) RoleService {
	return &roleService{db: db, roles: roles}
}

func (s *roleService) List(ctx context.Context) (*dto.ListRolesResponse, error) {
	roles, err := s.roles.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]uint, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}
	rows, err := s.roles.GetPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	permissionsByRole := make(map[uint]permission.Set, len(roles))
	for _, row := range rows {
		if permissionsByRole[row.RoleID] == nil {
			permissionsByRole[row.RoleID] = permission.Set{}
		}
		permissionsByRole[row.RoleID].Grant(permission.Resource(row.Resource), permission.ParseLevel(row.Level))
	}

	resources := make([]dto.ResourceResponse, len(permission.Resources))
	for i, def := range permission.Resources {
		resources[i] = dto.ResourceResponse{Resource: string(def.Resource), MaxLevel: def.MaxLevel.String()}
	}
	responses := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = toRoleResponse(role, permissionsByRole[role.ID])
	}
	return &dto.ListRolesResponse{Resources: resources, Roles: responses}, nil
}

func (s *roleService) UpdatePermissions(ctx context.Context, roleID uint, req *dto.UpdateRolePermissionsRequest) (*dto.RoleResponse, error) {
	role, err := s.roles.GetByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	if !isEditable(*role) {
		return nil, &ValidationError{Message: "管理者ロールの権限は変更できません"}
	}

	permissions, err := validatePermissions(*role, req.Permissions)
	if err != nil {
		return nil, err
	}

	rows := make([]models.RolePermission, 0, len(permissions))
	for _, def := range permission.Resources {
		if level := permissions.Level(def.Resource); level != permission.None {
			rows = append(rows, models.RolePermission{RoleID: role.ID, Resource: string(def.Resource), Level: level.String()})
		}
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roles := s.roles.WithTx(tx)
		if err := roles.DeletePermissionsByRoleID(ctx, role.ID); err != nil {
			return err
		}
		return roles.CreatePermissions(ctx, rows)
	})
	if err != nil {
		return nil, err
	}

	response := toRoleResponse(*role, permissions)
	return &response, nil
}

// validatePermissions はリクエストの権限を検証して permission.Set に変換する。
// 未知のリソース・不正な権限レベル・リソースの上限を超える権限は、丸めずにエラーにする
// （画面の表示と保存される内容が食い違わないようにするため）。
func validatePermissions(role models.Role, requested map[string]string) (permission.Set, error) {
	permissions := permission.Set{}
	for resource, value := range requested {
		def, ok := permission.Lookup(permission.Resource(resource))
		if !ok {
			return nil, &ValidationError{Message: fmt.Sprintf("不明なリソースです: %s", resource)}
		}
		level, ok := permission.LookupLevel(value)
		if !ok {
			return nil, &ValidationError{Message: fmt.Sprintf("不明な権限レベルです: %s", value)}
		}
		if level > def.MaxLevel {
			return nil, &ValidationError{Message: fmt.Sprintf("%s に設定できる権限は %s までです", resource, def.MaxLevel)}
		}
		permissions.Grant(def.Resource, level)
	}

	// ゲストロールは未ログインの訪問者全員に適用されるため、アクセス管理画面の権限は付けられない
	if role.Code == GuestRoleCode && permissions.Level(permission.AccessControl) != permission.None {
		return nil, &ValidationError{Message: "ゲストロールにはアクセス管理の権限を付けられません"}
	}
	return permissions, nil
}

func isEditable(role models.Role) bool {
	return role.Code != AdminRoleCode
}

// toRoleResponse はロールと権限をAPIのレスポンスに変換する。permissions には全リソースのキーを含める。
func toRoleResponse(role models.Role, permissions permission.Set) dto.RoleResponse {
	levels := make(map[string]string, len(permission.Resources))
	for _, def := range permission.Resources {
		levels[string(def.Resource)] = permissions.Level(def.Resource).String()
	}
	return dto.RoleResponse{
		ID:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Editable:    isEditable(role),
		Permissions: levels,
	}
}
