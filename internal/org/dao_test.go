package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/RyanBard/gin-ex/pkg/org"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	version     = int64(3)
)

func getRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"name",
		"description",
		"created_at",
		"updated_at",
		"version",
	}).AddRow(
		id,
		name,
		desc,
		createdAt,
		updatedAt,
		version,
	)
}

func initDAO() (d *dao, dbx *sqlx.DB, md sqlmock.Sqlmock) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.PanicLevel)
	db, md, err := sqlmock.New()
	if err != nil {
		log.Fatal("failed to mock db")
	}
	dbx = sqlx.NewDb(db, "sqlmock")
	queryTimeout := 30 * time.Second
	d = NewDAO(log, queryTimeout, dbx)
	return d, dbx, md
}

func TestDAOGetByID(t *testing.T) {
	d, _, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
		WithArgs(id).
		WillReturnRows(getRows())

	actual, err := d.GetByID(ctx, id)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, desc, actual.Desc)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, version, actual.Version)
}

func TestDAOGetByID_NotFoundErr(t *testing.T) {
	d, _, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	_, err := d.GetByID(ctx, id)

	assert.Nil(t, md.ExpectationsWereMet())
	var expected NotFoundErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), id)
}

func TestDAOGetByID_OtherErr(t *testing.T) {
	d, _, md := initDAO()

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
		WithArgs(id).
		WillReturnError(&mockErr)

	_, err := d.GetByID(ctx, id)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOGetAll(t *testing.T) {
	d, _, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getAllQuery)).
		WillReturnRows(getRows())

	actuals, err := d.GetAll(ctx)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, desc, actual.Desc)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, version, actual.Version)
}

func TestDAOGetAll_Error(t *testing.T) {
	d, _, md := initDAO()

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectQuery(regexp.QuoteMeta(getAllQuery)).
		WillReturnError(&mockErr)

	_, err := d.GetAll(ctx)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOSearchByName(t *testing.T) {
	d, _, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(searchByNameQuery)).
		WithArgs("%" + partialName + "%").
		WillReturnRows(getRows())

	actuals, err := d.SearchByName(ctx, partialName)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, desc, actual.Desc)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, version, actual.Version)
}

func TestDAOSearchByName_Error(t *testing.T) {
	d, _, md := initDAO()

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectQuery(regexp.QuoteMeta(searchByNameQuery)).
		WithArgs("%" + partialName + "%").
		WillReturnError(&mockErr)

	_, err := d.SearchByName(ctx, partialName)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOCreate(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
}

func TestDAOCreate_NameUniqueConstraintErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error", Constraint: "orgs_name_uk"}
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.NotNil(t, err)
	var dupNameErr NameAlreadyInUseErr
	assert.True(t, errors.As(err, &dupNameErr))
	assert.Contains(t, dupNameErr.Error(), "name")
	assert.Contains(t, dupNameErr.Error(), "in use")
	assert.Contains(t, dupNameErr.Error(), name)
}

func TestDAOCreate_OtherPQErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOCreate_OtherNonPQErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnError(mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, mockErr, err)
}

func TestDAOCreate_TooManyRowsAffected(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("INSERT INTO orgs").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "unexpected number of rows affected")
}

func TestDAOUpdate(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	actual, err := d.Update(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, o.ID, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.CreatedAt, actual.CreatedAt)
	assert.Equal(t, o.UpdatedAt, actual.UpdatedAt)
	assert.Equal(t, o.Version+1, actual.Version)
}

func TestDAOUpdate_OptimisticLockErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	var expected OptimisticLockErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "modified")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", version))
}

func TestDAOUpdate_NameAlreadyInUseErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error", Constraint: "orgs_name_uk"}
	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.NotNil(t, err)
	var dupNameErr NameAlreadyInUseErr
	assert.True(t, errors.As(err, &dupNameErr))
	assert.Contains(t, dupNameErr.Error(), "name")
	assert.Contains(t, dupNameErr.Error(), "in use")
	assert.Contains(t, dupNameErr.Error(), name)
}

func TestDAOUpdate_OtherPQErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOUpdate_OtherNonPQErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnError(mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, mockErr, err)
}

func TestDAOUpdate_TooManyRowsAffected(t *testing.T) {
	d, db, md := initDAO()

	o := org.Org{
		ID:        id,
		Name:      name,
		Desc:      desc,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE orgs").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "unexpected number of rows affected")
}

func TestDAODelete(t *testing.T) {
	d, db, md := initDAO()

	o := org.DeleteOrg{
		ID:      id,
		Version: version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
}

func TestDAODelete_OptimisticLockErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.DeleteOrg{
		ID:      id,
		Version: version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	var expected OptimisticLockErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "modified")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", version))
}

func TestDAODelete_OtherErr(t *testing.T) {
	d, db, md := initDAO()

	o := org.DeleteOrg{
		ID:      id,
		Version: version,
	}

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAODelete_TooManyRowsAffected(t *testing.T) {
	d, db, md := initDAO()

	o := org.DeleteOrg{
		ID:      id,
		Version: version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM orgs").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, o)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "unexpected number of rows affected")
}
