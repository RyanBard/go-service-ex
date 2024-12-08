package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/RyanBard/go-service-ex/internal/logutil"
	"github.com/RyanBard/go-service-ex/pkg/org"
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
		log:     log.With(logutil.LogAttrSVC("OrgDAO")),
		timeout: timeout,
		db:      db,
	}
}

func (d dao) GetByID(ctx context.Context, id string) (o org.Org, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetByID"),
		logAttrOrgID(id),
	)
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
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAll"),
	)
	log.Debug("called")
	err = d.db.SelectContext(ctx, &orgs, getAllQuery)
	if err != nil {
		return orgs, err
	}
	log.Debug("success")
	return orgs, err
}

func (d dao) SearchByName(ctx context.Context, name string) (orgs []org.Org, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("SearchByName"),
		logAttrOrgName(name),
	)
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
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Create"),
		logAttrOrg(o),
	)
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, createQuery, &o)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "orgs_name_uk" {
				return NameAlreadyInUseErr{Name: o.Name}
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

func (d dao) Update(ctx context.Context, tx *sqlx.Tx, input org.Org) (o org.Org, err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Update"),
		logAttrOrg(input),
	)
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, updateQuery, &input)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "orgs_name_uk" {
				return o, NameAlreadyInUseErr{Name: input.Name}
			}
		}
		return o, err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		return o, err
	}
	if numRows == 0 {
		return o, OptimisticLockErr{ID: input.ID, Version: input.Version}
	}
	if numRows != 1 {
		return o, fmt.Errorf("unexpected number of rows affected: %d", numRows)
	}
	log.Debug("success")
	input.Version = input.Version + 1
	return input, err
}

func (d dao) Delete(ctx context.Context, tx *sqlx.Tx, o org.DeleteOrg) (err error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	log := d.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Delete"),
		logAttrOrg(o),
	)
	log.Debug("called")
	r, err := tx.NamedExecContext(ctx, deleteQuery, &o)
	if err != nil {
		return err
	}
	log.Debug("query ran")
	numRows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if numRows == 0 {
		return OptimisticLockErr{ID: o.ID, Version: o.Version}
	}
	if numRows != 1 {
		return fmt.Errorf("unexpected number of rows affected: %d", numRows)
	}
	log.Debug("success")
	return err
}
