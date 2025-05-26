package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	logutil "github.com/RyanBard/go-log-util/pkg"
	"github.com/RyanBard/go-service-ex/pkg/user"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type dao struct {
	log     *slog.Logger
	timeout time.Duration
	db      *sqlx.DB
}

func NewDAO(log *slog.Logger, timeout time.Duration, db *sqlx.DB) *dao {
	return &dao{
		log:     log.With(logutil.LogAttrSVC("UserDAO")),
		timeout: timeout,
		db:      db,
	}
}

func (d dao) GetByID(ctx context.Context, id string) (u user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetByID"),
		logAttrUserID(id),
	)
	log.Debug("called")
	err = d.db.GetContext(ctx, &u, getByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, ErrNotFound{ID: id}
		}
		return u, err
	}
	log.Debug("success")
	return u, err
}

func (d dao) GetAll(ctx context.Context) (users []user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAll"),
	)
	log.Debug("called")
	err = d.db.SelectContext(ctx, &users, getAllQuery)
	if err != nil {
		return users, err
	}
	log.Debug("success")
	return users, err
}

func (d dao) GetAllByOrgID(ctx context.Context, orgID string) (users []user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAllByOrgID"),
		logAttrOrgID(orgID),
	)
	log.Debug("called")
	users = []user.User{}
	err = d.db.SelectContext(ctx, &users, getAllByOrgIDQuery, orgID)
	if err != nil {
		return users, err
	}
	log.Debug("success")
	return users, err
}

func (d dao) Create(ctx context.Context, tx *sqlx.Tx, u user.User) (err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Create"),
		logAttrUser(u),
	)
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, createQuery, &u)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "users_email_uk" {
				return ErrEmailAlreadyInUse{Email: u.Email}
			}
		}
		return err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if numRows != 1 {
		return fmt.Errorf("unexpected number of rows affected: %d", numRows)
	}
	log.Debug("success")
	return err
}

func (d dao) Update(ctx context.Context, tx *sqlx.Tx, input user.User) (u user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Update"),
		logAttrUser(input),
	)
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, updateQuery, &input)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "users_email_uk" {
				return u, ErrEmailAlreadyInUse{Email: input.Email}
			}
		}
		return u, err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		return u, err
	}
	if numRows == 0 {
		return u, ErrOptimisticLock{ID: input.ID, Version: input.Version}
	}
	if numRows != 1 {
		return u, fmt.Errorf("unexpected number of rows affected: %d", numRows)
	}
	log.Debug("success")
	input.Version = input.Version + 1
	return input, err
}

func (d dao) Delete(ctx context.Context, tx *sqlx.Tx, u user.DeleteUser) (err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Delete"),
		logAttrUser(u),
	)
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, deleteQuery, &u)
	if err != nil {
		return err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if numRows == 0 {
		return ErrOptimisticLock{ID: u.ID, Version: u.Version}
	}
	if numRows != 1 {
		return fmt.Errorf("unexpected number of rows affected: %d", numRows)
	}
	log.Debug("success")
	return err
}
