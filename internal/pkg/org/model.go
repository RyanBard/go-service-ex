package org

import "time"

type Org struct {
	ID         string    `json:"id" db:"org_id"`
	Name       string    `json:"name" validate:"required" db:"name"`
	Desc       string    `json:"desc" validate:"required" db:"description"`
	IsArchived bool      `json:"is_archived" db:"is_archived"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Version    int64     `json:"version" db:"version"`
}
