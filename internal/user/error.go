package user

import (
	"fmt"
)

type ErrEmailAlreadyInUse struct {
	Email string
}

func (err ErrEmailAlreadyInUse) Error() string {
	return fmt.Sprintf("Cannot save user, email '%s' is already in use by another user", err.Email)
}

type ErrCannotAssociateSysOrg struct {
	UserID string
	OrgID  string
}

func (err ErrCannotAssociateSysOrg) Error() string {
	return fmt.Sprintf("Cannot associate user with system org: userID=%s orgID=%s", err.UserID, err.OrgID)
}

type ErrCannotModifySysUser struct {
	ID string
}

func (err ErrCannotModifySysUser) Error() string {
	return fmt.Sprintf("Cannot modify system user: id=%s", err.ID)
}

type ErrNotFound struct {
	ID string
}

func (err ErrNotFound) Error() string {
	return fmt.Sprintf("User not found: id=%s ", err.ID)
}

type ErrOptimisticLock struct {
	ID      string
	Version int64
}

func (err ErrOptimisticLock) Error() string {
	return fmt.Sprintf("User was modified since last retrieved: id=%s version=%d", err.ID, err.Version)
}
