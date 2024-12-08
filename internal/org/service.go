package org

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/RyanBard/go-service-ex/internal/ctxutil"
	"github.com/RyanBard/go-service-ex/internal/logutil"
	"github.com/RyanBard/go-service-ex/pkg/org"
	"github.com/jmoiron/sqlx"
)

type OrgDAO interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
	GetAll(ctx context.Context) ([]org.Org, error)
	SearchByName(ctx context.Context, name string) ([]org.Org, error)
	Create(ctx context.Context, tx *sqlx.Tx, o org.Org) error
	Update(ctx context.Context, tx *sqlx.Tx, o org.Org) (org.Org, error)
	Delete(ctx context.Context, tx *sqlx.Tx, o org.DeleteOrg) error
}

type TXManager interface {
	Do(ctx context.Context, tx *sqlx.Tx, f func(*sqlx.Tx) error) error
}

type Timer interface {
	Now() time.Time
}

type IDGenerator interface {
	GenID() string
}

type service struct {
	log   *slog.Logger
	dao   OrgDAO
	txMGR TXManager
	timer Timer
	idGen IDGenerator
}

func NewService(log *slog.Logger, dao OrgDAO, txMGR TXManager, timer Timer, idGen IDGenerator) *service {
	return &service{
		log:   log.With(logutil.LogAttrSVC("OrgSVC")),
		dao:   dao,
		txMGR: txMGR,
		timer: timer,
		idGen: idGen,
	}
}

func (s service) GetByID(ctx context.Context, id string) (org.Org, error) {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetByID"),
		logAttrOrgID(id),
	)
	log.Debug("called")
	return s.dao.GetByID(ctx, id)
}

func (s service) GetAll(ctx context.Context, name string) ([]org.Org, error) {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAll"),
		logAttrOrgName(name),
	)
	log.Debug("called")
	if name == "" {
		return s.dao.GetAll(ctx)
	} else {
		return s.dao.SearchByName(ctx, strings.ToLower(name))
	}
}

func (s service) Save(ctx context.Context, o org.Org) (out org.Org, err error) {
	loggedInUserID, _ := ctx.Value(ctxutil.ContextKeyUserID{}).(string)
	if loggedInUserID == "" {
		return out, errors.New("user not logged in")
	}
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Save"),
		logAttrOrg(o),
	)
	log.Debug("called")
	if o.ID != "" {
		orgInDB, err := s.GetByID(ctx, o.ID)
		if err != nil {
			return out, err
		}
		if orgInDB.IsSystem {
			err = CannotModifySysOrgErr{ID: o.ID}
			return out, err
		}
	}
	err = s.txMGR.Do(ctx, nil, func(tx *sqlx.Tx) error {
		if o.ID == "" {
			o.ID = s.idGen.GenID()
			o.Version = 1
			o.CreatedAt = s.timer.Now()
			o.CreatedBy = loggedInUserID
			o.UpdatedAt = s.timer.Now()
			o.UpdatedBy = loggedInUserID
			o.IsSystem = false
			out = o
			return s.dao.Create(ctx, tx, o)
		} else {
			o.UpdatedAt = s.timer.Now()
			o.UpdatedBy = loggedInUserID
			out, err = s.dao.Update(ctx, tx, o)
			return err
		}
	})
	if err != nil {
		return org.Org{}, err
	}
	return out, nil
}

func (s service) Delete(ctx context.Context, o org.DeleteOrg) error {
	log := s.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Delete"),
		logAttrOrg(o),
	)
	log.Debug("called")
	orgInDB, err := s.GetByID(ctx, o.ID)
	if err != nil {
		return err
	}
	if orgInDB.IsSystem {
		err = CannotModifySysOrgErr{ID: o.ID}
		return err
	}
	err = s.txMGR.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return s.dao.Delete(ctx, tx, o)
	})
	if err != nil {
		return err
	}
	return nil
}
