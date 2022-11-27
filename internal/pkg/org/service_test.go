package org

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSVCGetByID(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	ctx := context.Background()
	id := "foo"
	actual, err := s.GetByID(ctx, id)
	assert.Nil(t, err)
	// TODO
	assert.Equal(t, id, actual.ID)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
}

func TestSVCGetAll(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	ctx := context.Background()
	actuals, err := s.GetAll(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	// TODO
	assert.Equal(t, "Foo Name", actual.Name)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
}

func TestSVCGetAll_NameSpecified(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	ctx := context.Background()
	name := "foo"
	actuals, err := s.GetAll(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	// TODO
	assert.Equal(t, name+"x Name", actual.Name)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
}

func TestSVCSave_NoID(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	ctx := context.Background()
	o := Org{
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
	}
	actual, err := s.Save(ctx, o)
	assert.Nil(t, err)
	// TODO
	assert.NotEqual(t, "", actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestSVCSave_ID(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	ctx := context.Background()
	o := Org{
		ID:         "foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		CreatedAt:  now,
	}
	actual, err := s.Save(ctx, o)
	assert.Nil(t, err)
	// TODO
	assert.Equal(t, o.ID, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestSVCDelete(t *testing.T) {
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	ctx := context.Background()
	id := "foo"
	err := s.Delete(ctx, id)
	assert.Nil(t, err)
}
