package org

import (
	"context"
	"github.com/sirupsen/logrus"
)

type OrgDAO interface {
	GetByID(ctx context.Context, id string) (Org, error)
	GetAll(ctx context.Context) ([]Org, error)
	SearchByName(ctx context.Context, name string) ([]Org, error)
	Create(ctx context.Context, o Org) (Org, error)
	Update(ctx context.Context, o Org) (Org, error)
	Delete(ctx context.Context, id string) error
}

type Timer interface {
	Now() int64
}

type IDGenerator interface {
	GenID() string
}

type service struct {
	log   logrus.FieldLogger
	dao   OrgDAO
	timer Timer
	idGen IDGenerator
}

func NewOrgService(log logrus.FieldLogger, dao OrgDAO, timer Timer, idGen IDGenerator) *service {
	return &service{
		log:   log.WithField("SVC", "OrgSVC"),
		dao:   dao,
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

func (s service) Save(ctx context.Context, o Org) (Org, error) {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Save",
		"org":   o,
	})
	log.Debug("called")
	if o.ID == "" {
		o.ID = s.idGen.GenID()
		o.CreatedAt = s.timer.Now()
		o.UpdatedAt = s.timer.Now()
		return s.dao.Create(ctx, o)
	} else {
		// TODO - maybe just take this out and don't look at createdAt when doing an update
		prev, err := s.GetByID(ctx, o.ID)
		if err != nil {
			return Org{}, err
		}
		o.CreatedAt = prev.CreatedAt
		o.UpdatedAt = s.timer.Now()
		return s.dao.Update(ctx, o)
	}
}

func (s service) Delete(ctx context.Context, id string) error {
	log := s.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"id":    id,
	})
	log.Debug("called")
	return s.dao.Delete(ctx, id)
}
