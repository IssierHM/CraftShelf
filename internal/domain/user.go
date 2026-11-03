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
	ID           int64
	PublicID     string
	Username     string
	Email        string
	PasswordHash string
	Bio          string
	AvatarURL    string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
