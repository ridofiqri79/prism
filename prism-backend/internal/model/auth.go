package model

import "github.com/jackc/pgx/v5/pgtype"

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	ExpiresIn   int      `json:"expires_in"`
	User        UserInfo `json:"user"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type MeResponse struct {
	UserInfo
	Permissions []PermissionInfo `json:"permissions"`
}

type PermissionInfo struct {
	Module    string `json:"module"`
	CanCreate bool   `json:"can_create"`
	CanRead   bool   `json:"can_read"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

type AuthUser struct {
	ID       pgtype.UUID
	Username string
	Role     string
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=ADMIN STAFF"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"required,oneof=ADMIN STAFF"`
	IsActive bool   `json:"is_active"`
}

type UpdatePermissionsRequest struct {
	Permissions []PermissionItem `json:"permissions" validate:"required"`
}

type PermissionItem struct {
	Module    string `json:"module" validate:"required"`
	CanCreate bool   `json:"can_create"`
	CanRead   bool   `json:"can_read"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
