package org

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type OrgDAO interface {
	GetByID(ctx context.Context, id string) (Org, error)
	GetAll(ctx context.Context) ([]Org, error)
	SearchByName(ctx context.Context, name string) ([]Org, error)
	Create(ctx context.Context, tx *sqlx.Tx, o Org) error
	Update(ctx context.Context, tx *sqlx.Tx, o Org) (Org, error)
	Delete(ctx context.Context, tx *sqlx.Tx, o DeleteOrg) error
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
	log   logrus.FieldLogger
	dao   OrgDAO
	txMGR TXManager
	timer Timer
	idGen IDGenerator
}

func NewService(log logrus.FieldLogger, dao OrgDAO, txMGR TXManager, timer Timer, idGen IDGenerator) *service {
	return &service{
		log:   log.WithField("SVC", "OrgSVC"),
		dao:   dao,
		txMGR: txMGR,
		timer: timer,
		idGen: idGen,
	}
}

func (s service) GetByID(ctx context.Context, id string) (Org, error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
	})
	log.Debug("called")
	return s.dao.GetByID(ctx, id)
}

func (s service) GetAll(ctx context.Context, name string) ([]Org, error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAll",
		"name":  name,
	})
	log.Debug("called")
	if name == "" {
		return s.dao.GetAll(ctx)
	} else {
		return s.dao.SearchByName(ctx, name)
	}
}

func (s service) Save(ctx context.Context, o Org) (out Org, err error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Save",
		"org":   o,
	})
	log.Debug("called")
	if o.ID != "" {
		orgInDB, err := s.GetByID(ctx, o.ID)
		if err != nil {
			log.WithError(err).Error("couldn't find org to update")
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
			o.CreatedBy = "TODO"
			o.UpdatedAt = s.timer.Now()
			o.UpdatedBy = "TODO"
			out = o
			return s.dao.Create(ctx, tx, o)
		} else {
			o.UpdatedAt = s.timer.Now()
			o.UpdatedBy = "TODO"
			out, err = s.dao.Update(ctx, tx, o)
			return err
		}
	})
	if err != nil {
		return Org{}, err
	}
	return out, nil
}

func (s service) Delete(ctx context.Context, o DeleteOrg) error {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"o":     o,
	})
	log.Debug("called")
	orgInDB, err := s.GetByID(ctx, o.ID)
	if err != nil {
		log.WithError(err).Error("couldn't find org to delete")
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
