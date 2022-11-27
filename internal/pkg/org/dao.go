package org

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

type dao struct {
	log logrus.FieldLogger
}

func NewOrgDAO(log logrus.FieldLogger) *dao {
	return &dao{
		log: log.WithField("SVC", "OrgDAO"),
	}
}

func (d dao) GetByID(ctx context.Context, id string) (Org, error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
	})
	log.Debug("called")
	o := Org{
		ID:         id,
		Name:       "Foo Name",
		Desc:       "Foo Desc",
		IsArchived: false,
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
	}
	return o, nil
}

func (d dao) GetAll(ctx context.Context) ([]Org, error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "GetAll",
	})
	log.Debug("called")
	o := Org{
		ID:         uuid.New().String(),
		Name:       "Foo Name",
		Desc:       "Foo Desc",
		IsArchived: false,
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
	}
	return []Org{o}, nil
}

func (d dao) SearchByName(ctx context.Context, name string) ([]Org, error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "SearchByName",
		"name":  name,
	})
	log.Debug("called")
	o := Org{
		ID:         uuid.New().String(),
		Name:       name + "x Name",
		Desc:       "Foo Desc",
		IsArchived: false,
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
	}
	return []Org{o}, nil
}

func (d dao) Create(ctx context.Context, o Org) (Org, error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Create",
		"org":   o,
	})
	log.Debug("called")
	return o, nil
}

func (d dao) Update(ctx context.Context, o Org) (Org, error) {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Update",
		"org":   o,
	})
	log.Debug("called")
	return o, nil
}

func (d dao) Delete(ctx context.Context, id string) error {
	log := d.log.WithFields(logrus.Fields{
		"reqID": ctx.Value("reqID"),
		"fn":    "Delete",
		"id":    id,
	})
	log.Debug("called")
	return nil
}
