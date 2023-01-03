package user

import "time"

type User struct {
	ID        string    `json:"id,omitempty" db:"id"`
	OrgID     string    `json:"org_id,omitempty" db:"org_id"`
	Name      string    `json:"name,omitempty" validate:"required" db:"name"`
	Email     string    `json:"email,omitempty" validate:"required" db:"email"`
	IsSystem  bool      `json:"is_system" db:"is_system"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy string    `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty" db:"updated_by"`
	Version   int64     `json:"version" db:"version"`
}

type DeleteUser struct {
	ID      string `json:"id,omitempty" validate:"required" db:"id"`
	Version int64  `json:"version" validate:"required" db:"version"`
}
