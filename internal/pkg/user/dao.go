package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type dao struct {
	log logrus.FieldLogger
	db  *sqlx.DB
}

func NewDAO(log logrus.FieldLogger, db *sqlx.DB) *dao {
	return &dao{
		log: log.WithField("SVC", "UserDAO"),
		db:  db,
	}
}

func (d dao) GetByID(ctx context.Context, id string) (u User, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
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

func (d dao) GetAll(ctx context.Context) (users []User, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAll",
	})
	log.Debug("called")
	err = d.db.SelectContext(ctx, &users, getAllQuery)
	if err != nil {
		return users, err
	}
	log.Debug("success")
	return users, err
}

func (d dao) GetAllByOrgID(ctx context.Context, orgID string) (users []User, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAllByOrgID",
		"orgID": orgID,
	})
	log.Debug("called")
	users = []User{}
	err = d.db.SelectContext(ctx, &users, getAllByOrgIDQuery, orgID)
	if err != nil {
		return users, err
	}
	log.Debug("success")
	return users, err
}

func (d dao) Create(ctx context.Context, tx *sqlx.Tx, u User) (err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Create",
		"user":  u,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, createQuery, &u)
	if err != nil {
		log.WithError(err).Error("failed to execute query")
		return err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		log.WithError(err).Error("failed to get number of rows affected")
		return err
	}
	if numRows != 1 {
		return errors.New(fmt.Sprintf("unexpected number of rows affected: %d", numRows))
	}
	log.Debug("success")
	return err
}

func (d dao) Update(ctx context.Context, tx *sqlx.Tx, input User) (u User, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Update",
		"user":  input,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, updateQuery, &input)
	if err != nil {
		log.WithError(err).Error("failed to execute query")
		return u, err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		log.WithError(err).Error("failed to get number of rows affected")
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

func (d dao) Delete(ctx context.Context, tx *sqlx.Tx, u DeleteUser) (err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"u":     u,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, deleteQuery, &u)
	if err != nil {
		log.WithError(err).Error("failed to execute query")
		return err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		log.WithError(err).Error("failed to get number of rows affected")
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
