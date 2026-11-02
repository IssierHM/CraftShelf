package domain

import (
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
)

type User struct {
	ID           int64     `json:"id"`
	PublicID     string    `json:"public_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Bio          string    `json:"bio"`
	AvatarURL    string    `json:"avatar_url"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
