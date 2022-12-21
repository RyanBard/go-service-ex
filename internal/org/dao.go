package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/RyanBard/gin-ex/pkg/org"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type dao struct {
	log logrus.FieldLogger
	db  *sqlx.DB
}

func NewDAO(log logrus.FieldLogger, db *sqlx.DB) *dao {
	return &dao{
		log: log.WithField("SVC", "OrgDAO"),
		db:  db,
	}
}

func (d dao) GetByID(ctx context.Context, id string) (o org.Org, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
	})
	log.Debug("called")
	err = d.db.GetContext(ctx, &o, getByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return o, NotFoundErr{ID: id}
		}
		return o, err
	}
	log.Debug("success")
	return o, err
}

func (d dao) GetAll(ctx context.Context) (orgs []org.Org, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAll",
	})
	log.Debug("called")
	err = d.db.SelectContext(ctx, &orgs, getAllQuery)
	if err != nil {
		return orgs, err
	}
	log.Debug("success")
	return orgs, err
}

func (d dao) SearchByName(ctx context.Context, name string) (orgs []org.Org, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "SearchByName",
		"name":  name,
	})
	log.Debug("called")
	orgs = []org.Org{}
	err = d.db.SelectContext(ctx, &orgs, searchByNameQuery, "%"+name+"%")
	if err != nil {
		return orgs, err
	}
	log.Debug("success")
	return orgs, err
}

func (d dao) Create(ctx context.Context, tx *sqlx.Tx, o org.Org) (err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Create",
		"org":   o,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, createQuery, &o)
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

func (d dao) Update(ctx context.Context, tx *sqlx.Tx, input org.Org) (o org.Org, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Update",
		"org":   input,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, updateQuery, &input)
	if err != nil {
		log.WithError(err).Error("failed to execute query")
		return o, err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		log.WithError(err).Error("failed to get number of rows affected")
		return o, err
	}
	if numRows == 0 {
		return o, OptimisticLockErr{ID: input.ID, Version: input.Version}
	}
	if numRows != 1 {
		return o, errors.New(fmt.Sprintf("unexpected number of rows affected: %d", numRows))
	}
	log.Debug("success")
	input.Version = input.Version + 1
	return input, err
}

func (d dao) Delete(ctx context.Context, tx *sqlx.Tx, o org.DeleteOrg) (err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"o":     o,
	})
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, deleteQuery, &o)
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
		return OptimisticLockErr{ID: o.ID, Version: o.Version}
	}
	if numRows != 1 {
		return errors.New(fmt.Sprintf("unexpected number of rows affected: %d", numRows))
	}
	log.Debug("success")
	return err
}
