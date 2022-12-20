package org

import (
	"fmt"
)

type CannotModifySysOrgErr struct {
	ID string
}

func (err CannotModifySysOrgErr) Error() string {
	return fmt.Sprintf("Cannot modify system org: id=%s", err.ID)
}

type NotFoundErr struct {
	ID string
}

func (err NotFoundErr) Error() string {
	return fmt.Sprintf("Org not found: id=%s", err.ID)
}

type OptimisticLockErr struct {
	ID      string
	Version int64
}

func (err OptimisticLockErr) Error() string {
	return fmt.Sprintf("Org was modified since last retrieved: id=%s version=%d", err.ID, err.Version)
}
