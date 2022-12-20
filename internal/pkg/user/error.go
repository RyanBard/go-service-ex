package user

import (
	"fmt"
)

type CannotAssociateSysOrgErr struct {
	UserID string
	OrgID  string
}

func (err CannotAssociateSysOrgErr) Error() string {
	return fmt.Sprintf("Cannot associate user with system org: userID=%s orgID=%s", err.UserID, err.OrgID)
}

type CannotModifySysUserErr struct {
	ID string
}

func (err CannotModifySysUserErr) Error() string {
	return fmt.Sprintf("Cannot modify system user: id=%s", err.ID)
}

type NotFoundErr struct {
	ID string
}

func (err NotFoundErr) Error() string {
	return fmt.Sprintf("User not found: id=%s ", err.ID)
}

type OptimisticLockErr struct {
	ID      string
	Version int64
}

func (err OptimisticLockErr) Error() string {
	return fmt.Sprintf("User was modified since last retrieved: id=%s version=%d", err.ID, err.Version)
}
