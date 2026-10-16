package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/dto"
	"github.com/codebeaver50/my-portal/server/internal/accesscontrol/repository"
	"github.com/codebeaver50/my-portal/server/internal/models"
)

// newRoleFixture は admin / guest / editor のロールを作り、RoleService を返す。
func newRoleFixture(t *testing.T) (*gorm.DB, RoleService) {
	t.Helper()
	db := newTestDB(t)
	mustCreate(t, db, &[]models.Role{
		{ID: 1, Code: AdminRoleCode, Name: "管理者", IsSystem: true},
		{ID: 2, Code: GuestRoleCode, Name: "ゲスト", IsSystem: true},
		{ID: 3, Code: "editor", Name: "編集者"},
	})
	mustCreate(t, db, &[]models.RolePermission{
		{RoleID: 1, Resource: "access-control", Level: "write"},
		{RoleID: 2, Resource: "book-database", Level: "read"},
		{RoleID: 3, Resource: "simple-ledger", Level: "write"},
		{RoleID: 3, Resource: "form-builder", Level: "read"},
	})
	return db, NewRoleService(db, repository.NewRoleRepository(db))
}

func permissionRows(t *testing.T, db *gorm.DB, roleID uint) map[string]string {
	t.Helper()
	var rows []models.RolePermission
	if err := db.Where("role_id = ?", roleID).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.Resource] = row.Level
	}
	return result
}

func TestRoleService_List(t *testing.T) {
	_, svc := newRoleFixture(t)

	result, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Resources) != 6 || result.Resources[0].Resource != "book-database" {
		t.Errorf("resources = %+v", result.Resources)
	}
	for _, resource := range result.Resources {
		if resource.Resource == "shop-css" && resource.MaxLevel != "read" {
			t.Errorf("shop-css maxLevel = %s, want read", resource.MaxLevel)
		}
	}

	if len(result.Roles) != 3 {
		t.Fatalf("roles = %+v", result.Roles)
	}
	admin, editor := result.Roles[0], result.Roles[2]
	if admin.Editable || !editor.Editable {
		t.Errorf("editable: admin = %v, editor = %v", admin.Editable, editor.Editable)
	}
	if editor.Permissions["simple-ledger"] != "write" || editor.Permissions["form-builder"] != "read" {
		t.Errorf("editor permissions = %v", editor.Permissions)
	}
	// 行のないリソースも none として全キーを返す
	if len(editor.Permissions) != 6 || editor.Permissions["shop-css"] != "none" {
		t.Errorf("editor permissions should include every resource: %v", editor.Permissions)
	}
}

func TestRoleService_UpdatePermissions(t *testing.T) {
	t.Run("replaces all permissions of the role", func(t *testing.T) {
		db, svc := newRoleFixture(t)

		result, err := svc.UpdatePermissions(context.Background(), 3, &dto.UpdateRolePermissionsRequest{
			Permissions: map[string]string{"book-database": "write", "simple-ledger": "none", "shop-css": "read"},
		})
		if err != nil {
			t.Fatal(err)
		}

		want := map[string]string{"book-database": "write", "shop-css": "read"}
		got := permissionRows(t, db, 3)
		if len(got) != len(want) || got["book-database"] != "write" || got["shop-css"] != "read" {
			t.Errorf("rows = %v, want %v (none and omitted resources should have no row)", got, want)
		}
		if result.Permissions["simple-ledger"] != "none" || result.Permissions["form-builder"] != "none" {
			t.Errorf("response permissions = %v", result.Permissions)
		}
	})

	t.Run("rejects invalid requests without changing permissions", func(t *testing.T) {
		cases := map[string]struct {
			roleID      uint
			permissions map[string]string
		}{
			"admin role":                {1, map[string]string{"access-control": "read"}},
			"unknown resource":          {3, map[string]string{"unknown": "read"}},
			"unknown level":             {3, map[string]string{"book-database": "admin"}},
			"above the max level":       {3, map[string]string{"shop-css": "write"}},
			"access control for guests": {2, map[string]string{"access-control": "read"}},
		}
		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				db, svc := newRoleFixture(t)
				before := permissionRows(t, db, tc.roleID)

				_, err := svc.UpdatePermissions(context.Background(), tc.roleID, &dto.UpdateRolePermissionsRequest{Permissions: tc.permissions})
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("err = %v, want ValidationError", err)
				}
				if after := permissionRows(t, db, tc.roleID); len(after) != len(before) {
					t.Errorf("permissions changed: %v -> %v", before, after)
				}
			})
		}
	})

	t.Run("allows guest role to change feature permissions", func(t *testing.T) {
		db, svc := newRoleFixture(t)

		_, err := svc.UpdatePermissions(context.Background(), 2, &dto.UpdateRolePermissionsRequest{
			Permissions: map[string]string{"simple-ledger": "write", "access-control": "none"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := permissionRows(t, db, 2); len(got) != 1 || got["simple-ledger"] != "write" {
			t.Errorf("rows = %v", got)
		}
	})

	t.Run("returns ErrRoleNotFound for unknown role", func(t *testing.T) {
		_, svc := newRoleFixture(t)

		_, err := svc.UpdatePermissions(context.Background(), 99, &dto.UpdateRolePermissionsRequest{Permissions: map[string]string{}})
		if !errors.Is(err, ErrRoleNotFound) {
			t.Errorf("err = %v, want ErrRoleNotFound", err)
		}
	})
}
