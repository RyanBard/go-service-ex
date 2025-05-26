package org

import (
	"fmt"
)

type ErrNameAlreadyInUse struct {
	Name string
}

func (err ErrNameAlreadyInUse) Error() string {
	return fmt.Sprintf("Cannot save org, name '%s' is already in use by another org", err.Name)
}

type ErrCannotModifySysOrg struct {
	ID string
}

func (err ErrCannotModifySysOrg) Error() string {
	return fmt.Sprintf("Cannot modify system org: id=%s", err.ID)
}

type ErrNotFound struct {
	ID string
}

func (err ErrNotFound) Error() string {
	return fmt.Sprintf("Org not found: id=%s", err.ID)
}

type ErrOptimisticLock struct {
	ID      string
	Version int64
}

func (err ErrOptimisticLock) Error() string {
	return fmt.Sprintf("Org was modified since last retrieved: id=%s version=%d", err.ID, err.Version)
}
