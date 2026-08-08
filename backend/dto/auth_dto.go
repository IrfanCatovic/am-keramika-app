package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	FullName string `json:"fullName,omitempty"`
	IsActive bool   `json:"isActive,omitempty"`
}

type LoginResponse struct {
	Token string           `json:"token"`
	User  AuthUserResponse `json:"user"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
	FullName string `json:"fullName" binding:"required,min=2"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"required"`
	FullName string `json:"fullName" binding:"required,min=2"`
}

type UpdateUserPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"isActive"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	FullName string `json:"fullName"`
	IsActive bool   `json:"isActive"`
}

type UserSummaryResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}
