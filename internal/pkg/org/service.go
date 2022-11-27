package org

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

type OrgDAO interface {
	GetByID(ctx context.Context, id string) (Org, error)
	GetAll(ctx context.Context) ([]Org, error)
	SearchByName(ctx context.Context, name string) ([]Org, error)
	Create(ctx context.Context, o Org) (Org, error)
	Update(ctx context.Context, o Org) (Org, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	log logrus.FieldLogger
	dao OrgDAO
}

func NewOrgService(log logrus.FieldLogger, dao OrgDAO) *service {
	return &service{
		log: log.WithField("SVC", "OrgSVC"),
		dao: dao,
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
	o.UpdatedAt = time.Now().UnixMilli()
	if o.ID == "" {
		o.ID = uuid.New().String()
		o.CreatedAt = time.Now().UnixMilli()
		return s.dao.Create(ctx, o)
	} else {
		// TODO - look up previous org for validation and CreatedAt
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
