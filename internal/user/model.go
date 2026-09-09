package user

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleEmployee Role = "employee"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateInput struct {
	Email        string
	Password     string
	PasswordHash string
	FullName     string
	Role         Role
}

type ListFilter struct {
	IsActive *bool
}
