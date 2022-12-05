package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type NotFoundErr struct {
	ID string
}

func (err NotFoundErr) Error() string {
	return fmt.Sprintf("Org not found: id=%s ", err.ID)
}

type OptimisticLockErr struct {
	ID      string
	Version int64
}

func (err OptimisticLockErr) Error() string {
	return fmt.Sprintf("Org was modified since last retrieved: id=%s version=%d", err.ID, err.Version)
}

type dao struct {
	log logrus.FieldLogger
	db  *sqlx.DB
}

func NewOrgDAO(log logrus.FieldLogger, db *sqlx.DB) *dao {
	return &dao{
		log: log.WithField("SVC", "OrgDAO"),
		db:  db,
	}
}

func (d dao) GetByID(ctx context.Context, id string) (o Org, err error) {
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

func (d dao) GetAll(ctx context.Context) (orgs []Org, err error) {
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

func (d dao) SearchByName(ctx context.Context, name string) (orgs []Org, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "SearchByName",
		"name":  name,
	})
	log.Debug("called")
	orgs = []Org{}
	err = d.db.SelectContext(ctx, &orgs, searchByNameQuery, "%"+name+"%")
	if err != nil {
		return orgs, err
	}
	log.Debug("success")
	return orgs, err
}

func (d dao) Create(ctx context.Context, o Org) (err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Create",
		"org":   o,
	})
	log.Debug("called")
	tx, ok := ctx.Value("pg-tx").(*sqlx.Tx)
	if ok {
		log.Debug("tx found, joining it")
	} else {
		log.Debug("tx not found, creating a new one")
		tx, err = d.db.Beginx()
		if err != nil {
			log.WithError(err).Error("tx not found, failed to create a new one")
			return err
		}
		defer func() {
			if err != nil {
				log.Warn("rolling back automatically created tx")
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					log.WithError(rollbackErr).Error("failed to rollback")
				}
				return
			}
			log.Debug("committing automatically created tx")
			err = tx.Commit()
		}()
	}
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

func (d dao) Update(ctx context.Context, input Org) (o Org, err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Update",
		"org":   input,
	})
	log.Debug("called")
	tx, ok := ctx.Value("pg-tx").(*sqlx.Tx)
	if ok {
		log.Debug("tx found, joining it")
	} else {
		log.Debug("tx not found, creating a new one")
		tx, err = d.db.Beginx()
		if err != nil {
			log.WithError(err).Error("tx not found, failed to create a new one")
			return o, err
		}
		defer func() {
			if err != nil {
				log.Warn("rolling back automatically created tx")
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					log.WithError(rollbackErr).Error("failed to rollback")
				}
				return
			}
			log.Debug("committing automatically created tx")
			err = tx.Commit()
		}()
	}
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

func (d dao) Delete(ctx context.Context, o Org) (err error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"o":     o,
	})
	log.Debug("called")
	tx, ok := ctx.Value("pg-tx").(*sqlx.Tx)
	if ok {
		log.Debug("tx found, joining it")
	} else {
		log.Debug("tx not found, creating a new one")
		tx, err = d.db.Beginx()
		if err != nil {
			log.WithError(err).Error("tx not found, failed to create a new one")
			return err
		}
		defer func() {
			if err != nil {
				log.Warn("rolling back automatically created tx")
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					log.WithError(rollbackErr).Error("failed to rollback")
				}
				return
			}
			log.Debug("committing automatically created tx")
			err = tx.Commit()
		}()
	}
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
