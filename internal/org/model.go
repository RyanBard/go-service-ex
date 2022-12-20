package org

import "time"

type Org struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" validate:"required" db:"name"`
	Desc      string    `json:"desc" validate:"required" db:"description"`
	IsSystem  bool      `json:"is_system" db:"is_system"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy string    `json:"updated_by" db:"updated_by"`
	Version   int64     `json:"version" db:"version"`
}

type DeleteOrg struct {
	ID      string `json:"id" validate:"required" db:"id"`
	Version int64  `json:"version" validate:"required" db:"version"`
}
