package user

import "time"

type User struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Name      string    `json:"name" validate:"required" db:"name"`
	Email     string    `json:"email" validate:"required" db:"email"`
	IsSystem  bool      `json:"is_system" db:"is_system"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy string    `json:"updated_by" db:"updated_by"`
	Version   int64     `json:"version" db:"version"`
}

type DeleteUser struct {
	ID      string `json:"id" validate:"required" db:"id"`
	Version int64  `json:"version" validate:"required" db:"version"`
}
