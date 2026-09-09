package user

import "time"

type CreateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type CreateByAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     Role   `json:"role"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      Role   `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewUserResponse(user User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
