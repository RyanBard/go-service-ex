package org

type Org struct {
	ID         string `json:"id"`
	Name       string `json:"name" validate:"required"`
	Desc       string `json:"desc" validate:"required"`
	IsArchived bool   `json:"is_archived"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}
