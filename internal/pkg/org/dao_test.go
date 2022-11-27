package org

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDAOGetByID(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	ctx := context.Background()
	id := "foo"
	actual, err := d.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, "Foo Name", actual.Name)
	assert.Equal(t, "Foo Desc", actual.Desc)
	assert.Equal(t, false, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestDAOGetAll(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	ctx := context.Background()
	actuals, err := d.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.NotEqual(t, "", actual.ID)
	assert.Equal(t, "Foo Name", actual.Name)
	assert.Equal(t, "Foo Desc", actual.Desc)
	assert.Equal(t, false, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestDAOSearchByName(t *testing.T) {
	now := time.Now().UnixMilli()
	name := "foo"
	d := NewOrgDAO(logrus.StandardLogger())
	ctx := context.Background()
	actuals, err := d.SearchByName(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.NotEqual(t, "", actual.ID)
	assert.Equal(t, name+"x Name", actual.Name)
	assert.Equal(t, "Foo Desc", actual.Desc)
	assert.Equal(t, false, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestDAOCreate(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	ctx := context.Background()
	o := Org{
		ID:         "foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	actual, err := d.Create(ctx, o)
	assert.Nil(t, err)
	assert.Equal(t, o.ID, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.Equal(t, o.CreatedAt, actual.CreatedAt)
	assert.Equal(t, o.UpdatedAt, actual.UpdatedAt)
}

func TestDAOUpdate(t *testing.T) {
	now := time.Now().UnixMilli()
	d := NewOrgDAO(logrus.StandardLogger())
	ctx := context.Background()
	o := Org{
		ID:         "foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	actual, err := d.Update(ctx, o)
	assert.Nil(t, err)
	assert.Equal(t, o.ID, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.Equal(t, o.CreatedAt, actual.CreatedAt)
	assert.Equal(t, o.UpdatedAt, actual.UpdatedAt)
}

func TestDAODelete(t *testing.T) {
	d := NewOrgDAO(logrus.StandardLogger())
	ctx := context.Background()
	id := "foo"
	err := d.Delete(ctx, id)
	assert.Nil(t, err)
}
