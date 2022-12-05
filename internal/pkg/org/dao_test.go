package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"regexp"
	"testing"
	"time"
)

var (
	ctx       = context.Background()
	createdAt = time.UnixMilli(100)
	updatedAt = time.UnixMilli(200)
)

const (
	partialName = "foo"
	id          = "foo-id"
	name        = "foo-name"
	desc        = "foo-desc"
	isArchived  = true
	version     = int64(3)
)

func getRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"org_id",
		"name",
		"description",
		"is_archived",
		"created_at",
		"updated_at",
		"version",
	}).AddRow(
		id,
		name,
		desc,
		isArchived,
		createdAt,
		updatedAt,
		version,
	)
}

func initDAO() (d *dao, md sqlmock.Sqlmock) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.PanicLevel)
	db, md, err := sqlmock.New()
	if err != nil {
		log.Fatal("failed to mock db")
	}
	d = NewOrgDAO(log, sqlx.NewDb(db, "sqlmock"))
	return d, md
}

func TestDAOGetByID(t *testing.T) {
	d, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
		WithArgs(id).
		WillReturnRows(getRows())

	actual, err := d.GetByID(ctx, id)

	assert.Nil(t, err)
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, desc, actual.Desc)
	assert.Equal(t, isArchived, actual.IsArchived)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, version, actual.Version)
}

func TestDAOGetByID_NotFoundErr(t *testing.T) {
	d, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	_, err := d.GetByID(ctx, id)

	var expected NotFoundErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), id)
}

func TestDAOGetByID_OtherErr(t *testing.T) {
	d, md := initDAO()

	mockErr := errors.New("unit-test mock error")
	md.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
		WithArgs(id).
		WillReturnError(mockErr)

	_, err := d.GetByID(ctx, id)

	assert.Equal(t, mockErr, err)
}

func TestDAOGetAll(t *testing.T) {
	d, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getAllQuery)).
		WillReturnRows(getRows())

	actuals, err := d.GetAll(ctx)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, desc, actual.Desc)
	assert.Equal(t, isArchived, actual.IsArchived)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, version, actual.Version)
}

func TestDAOGetAll_Error(t *testing.T) {
	d, md := initDAO()

	mockErr := errors.New("unit-test mock error")
	md.ExpectQuery(regexp.QuoteMeta(getAllQuery)).
		WillReturnError(mockErr)

	_, err := d.GetAll(ctx)

	assert.Equal(t, mockErr, err)
}

func TestDAOSearchByName(t *testing.T) {
	d, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(searchByNameQuery)).
		WithArgs("%" + partialName + "%").
		WillReturnRows(getRows())

	actuals, err := d.SearchByName(ctx, partialName)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, desc, actual.Desc)
	assert.Equal(t, isArchived, actual.IsArchived)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, version, actual.Version)
}

func TestDAOSearchByName_Error(t *testing.T) {
	d, md := initDAO()

	mockErr := errors.New("unit-test mock error")
	md.ExpectQuery(regexp.QuoteMeta(searchByNameQuery)).
		WithArgs("%" + partialName + "%").
		WillReturnError(mockErr)

	_, err := d.SearchByName(ctx, partialName)

	assert.Equal(t, mockErr, err)
}

func TestDAOCreate(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnResult(sqlmock.NewResult(1, 1))
	md.ExpectCommit()

	err := d.Create(ctx, o)

	assert.Nil(t, err)
}

func TestDAOCreate_Err(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnError(mockErr)
	md.ExpectRollback()

	err := d.Create(ctx, o)

	assert.Equal(t, mockErr, err)
}

func TestDAOUpdate(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnResult(sqlmock.NewResult(1, 1))
	md.ExpectCommit()

	actual, err := d.Update(ctx, o)

	assert.Nil(t, err)
	assert.Equal(t, o.ID, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.Equal(t, o.CreatedAt, actual.CreatedAt)
	assert.Equal(t, o.UpdatedAt, actual.UpdatedAt)
	assert.Equal(t, o.Version+1, actual.Version)
}

func TestDAOUpdate_OptimisticLockErr(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnResult(sqlmock.NewResult(0, 0))
	md.ExpectRollback()

	_, err := d.Update(ctx, o)

	var expected OptimisticLockErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "modified")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", version))
}

func TestDAOUpdate_OtherErr(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnError(mockErr)
	md.ExpectRollback()

	_, err := d.Update(ctx, o)

	assert.Equal(t, mockErr, err)
}

func TestDAODelete(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnResult(sqlmock.NewResult(1, 1))
	md.ExpectCommit()

	err := d.Delete(ctx, o)

	assert.Nil(t, err)
}

func TestDAODelete_OptimisticLockErr(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnResult(sqlmock.NewResult(0, 0))
	md.ExpectRollback()

	err := d.Delete(ctx, o)

	var expected OptimisticLockErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "modified")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", version))
}

func TestDAODelete_OtherErr(t *testing.T) {
	d, md := initDAO()

	o := Org{
		ID:         id,
		Name:       name,
		Desc:       desc,
		IsArchived: isArchived,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		Version:    version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnError(mockErr)
	md.ExpectRollback()

	err := d.Delete(ctx, o)

	assert.Equal(t, mockErr, err)
}
