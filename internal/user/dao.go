package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RyanBard/go-service-ex/pkg/user"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type dao struct {
	log     logrus.FieldLogger
	timeout time.Duration
	db      *sqlx.DB
}

func NewDAO(log logrus.FieldLogger, timeout time.Duration, db *sqlx.DB) *dao {
	return &dao{
		log:     log.WithField("svc", "UserDAO"),
		timeout: timeout,
		db:      db,
	}
}

func (d dao) GetByID(ctx context.Context, id string) (u user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetByID",
		"id":             id,
	})
	log.Debug("called")
	err = d.db.GetContext(ctx, &u, getByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, NotFoundErr{ID: id}
		}
		return u, err
	}
	log.Debug("success")
	return u, err
}

func (d dao) GetAll(ctx context.Context) (users []user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetAll",
	})
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
	log := d.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetAllByOrgID",
		"orgID":          orgID,
	})
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
	log := d.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "Create",
		"user":           u,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, createQuery, &u)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "users_email_uk" {
				return EmailAlreadyInUseErr{Email: u.Email}
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
		return errors.New(fmt.Sprintf("unexpected number of rows affected: %d", numRows))
	}
	log.Debug("success")
	return err
}

func (d dao) Update(ctx context.Context, tx *sqlx.Tx, input user.User) (u user.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "Update",
		"user":           input,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, updateQuery, &input)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "users_email_uk" {
				return u, EmailAlreadyInUseErr{Email: input.Email}
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
		return u, OptimisticLockErr{ID: input.ID, Version: input.Version}
	}
	if numRows != 1 {
		return u, errors.New(fmt.Sprintf("unexpected number of rows affected: %d", numRows))
	}
	log.Debug("success")
	input.Version = input.Version + 1
	return input, err
}

func (d dao) Delete(ctx context.Context, tx *sqlx.Tx, u user.DeleteUser) (err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "Delete",
		"u":              u,
	})
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
		return OptimisticLockErr{ID: u.ID, Version: u.Version}
	}
	if numRows != 1 {
		return errors.New(fmt.Sprintf("unexpected number of rows affected: %d", numRows))
	}
	log.Debug("success")
	return err
}
