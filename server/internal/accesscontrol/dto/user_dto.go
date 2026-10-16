package dto

// CreateUserRequest はユーザーを作成するときの入力。RoleCode のロールを1つ割り当てる。
type CreateUserRequest struct {
	LoginID     string
	DisplayName string
	Password    string
	RoleCode    string
}

// UserResponse は作成したユーザーの情報。パスワードのハッシュは含めない。
type UserResponse struct {
	ID          uint   `json:"id"`
	LoginID     string `json:"loginId"`
	DisplayName string `json:"displayName"`
	RoleCode    string `json:"roleCode"`
}
