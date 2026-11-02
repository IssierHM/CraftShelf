package dto

import (
	"time"

	"github.com/IssierHM/CraftShelf/internal/domain"
)

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email,max=254"`
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	PublicID  string `json:"public_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func ToUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		PublicID:  u.PublicID,
		Email:     u.Email,
		Username:  u.Username,
		Bio:       u.Bio,
		AvatarURL: u.AvatarURL,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
	}
}
