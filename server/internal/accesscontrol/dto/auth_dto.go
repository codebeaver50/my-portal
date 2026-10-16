// Package dto はポータル共通のアクセス管理のHTTP APIにおけるリクエスト/レスポンス型を定義する。
// service層がこれを組み立ててcontrollerに返す。
package dto

// LoginRequest は POST /api/auth/login のリクエスト。
// bcrypt は72バイトを超える部分を無視するため、パスワードは72文字までに制限する。
type LoginRequest struct {
	LoginID  string `json:"loginId" binding:"required,max=50"`
	Password string `json:"password" binding:"required,max=72"`
}

// RoleSummary はユーザーに割り当てられたロール1件分。
type RoleSummary struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// SessionUser はログイン中のユーザー。
type SessionUser struct {
	ID          uint          `json:"id"`
	LoginID     string        `json:"loginId"`
	DisplayName string        `json:"displayName"`
	Roles       []RoleSummary `json:"roles"`
}

// SessionResponse は GET /api/auth/session・POST /api/auth/login のレスポンス。
// 未ログインのときは user が null になり、permissions はゲストロールの権限になる。
// permissions はリソースごとの権限レベル（"none" / "read" / "write"）で、全リソースのキーを含む。
type SessionResponse struct {
	User        *SessionUser      `json:"user"`
	Permissions map[string]string `json:"permissions"`
}
