package org

import "time"

type Org struct {
	ID        string    `json:"id,omitempty" db:"id"`
	Name      string    `json:"name,omitempty" binding:"required" db:"name"`
	Desc      string    `json:"desc,omitempty" binding:"required" db:"description"`
	IsSystem  bool      `json:"is_system" db:"is_system"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy string    `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty" db:"updated_by"`
	Version   int64     `json:"version" db:"version"`
}

type DeleteOrg struct {
	ID      string `json:"id,omitempty" binding:"required" db:"id"`
	Version int64  `json:"version" binding:"required" db:"version"`
}
