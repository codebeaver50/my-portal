package dto

// ResourceResponse は権限を設定できるリソース1件分。maxLevel はそのリソースに設定できる最大の権限（"read" / "write"）。
type ResourceResponse struct {
	Resource string `json:"resource"`
	MaxLevel string `json:"maxLevel"`
}

// RoleResponse はロール1件分と、そのロールのリソースごとの権限。
// permissions は全リソースのキーを含み、値は "none" / "read" / "write"。
// editable が false のロール（admin）は権限を変更できない。
type RoleResponse struct {
	ID          uint              `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	IsSystem    bool              `json:"isSystem"`
	Editable    bool              `json:"editable"`
	Permissions map[string]string `json:"permissions"`
}

// ListRolesResponse は GET /api/access-control/roles のレスポンス。resources は権限設定画面の列の表示順。
type ListRolesResponse struct {
	Resources []ResourceResponse `json:"resources"`
	Roles     []RoleResponse     `json:"roles"`
}

// UpdateRolePermissionsRequest は PUT /api/access-control/roles/:id/permissions のリクエスト。
// ロールの権限をまるごと置き換える。キーのないリソースは権限なし（"none"）になる。
type UpdateRolePermissionsRequest struct {
	Permissions map[string]string `json:"permissions" binding:"required"`
}
