package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/RyanBard/go-service-ex/pkg/user"
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
	orgID       = "foo-org-id"
	name        = "foo-name"
	email       = "foo@bar.com"
	isAdmin     = true
	createdBy   = "logged-in-user-id"
	updatedBy   = "logged-in-user-id"
	version     = int64(3)
)

func getRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"org_id",
		"name",
		"email",
		"is_admin",
		"created_at",
		"created_by",
		"updated_at",
		"updated_by",
		"version",
	}).AddRow(
		id,
		orgID,
		name,
		email,
		isAdmin,
		createdAt,
		createdBy,
		updatedAt,
		updatedBy,
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
	assert.Equal(t, orgID, actual.OrgID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, email, actual.Email)
	assert.Equal(t, isAdmin, actual.IsAdmin)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, createdBy, actual.CreatedBy)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, updatedBy, actual.UpdatedBy)
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
	assert.Equal(t, orgID, actual.OrgID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, email, actual.Email)
	assert.Equal(t, isAdmin, actual.IsAdmin)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, createdBy, actual.CreatedBy)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, updatedBy, actual.UpdatedBy)
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

func TestDAOGetAllByOrgID(t *testing.T) {
	d, _, md := initDAO()

	md.ExpectQuery(regexp.QuoteMeta(getAllByOrgIDQuery)).
		WithArgs(orgID).
		WillReturnRows(getRows())

	actuals, err := d.GetAllByOrgID(ctx, orgID)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, orgID, actual.OrgID)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, email, actual.Email)
	assert.Equal(t, isAdmin, actual.IsAdmin)
	assert.Equal(t, createdAt, actual.CreatedAt)
	assert.Equal(t, createdBy, actual.CreatedBy)
	assert.Equal(t, updatedAt, actual.UpdatedAt)
	assert.Equal(t, updatedBy, actual.UpdatedBy)
	assert.Equal(t, version, actual.Version)
}

func TestDAOGetAllByOrgID_Error(t *testing.T) {
	d, _, md := initDAO()

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectQuery(regexp.QuoteMeta(getAllByOrgIDQuery)).
		WithArgs(orgID).
		WillReturnError(&mockErr)

	_, err := d.GetAllByOrgID(ctx, orgID)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOCreate(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("INSERT INTO users").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
}

func TestDAOCreate_EmailALreadyInUseErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error", Constraint: "users_email_uk"}
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO users").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.NotNil(t, err)
	var dupEmailErr EmailAlreadyInUseErr
	assert.True(t, errors.As(err, &dupEmailErr))
	assert.Contains(t, dupEmailErr.Error(), "email")
	assert.Contains(t, dupEmailErr.Error(), "in use")
	assert.Contains(t, dupEmailErr.Error(), email)
}

func TestDAOCreate_OtherPQErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO users").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOCreate_OtherNonPQErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("INSERT INTO users").
		WillReturnError(mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, mockErr, err)
}

func TestDAOCreate_TooManyRowsAffected(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("INSERT INTO users").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Create(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "unexpected number of rows affected")
}

func TestDAOUpdate(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE users").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	actual, err := d.Update(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, u.ID, actual.ID)
	assert.Equal(t, u.OrgID, actual.OrgID)
	assert.Equal(t, u.Name, actual.Name)
	assert.Equal(t, u.Email, actual.Email)
	assert.Equal(t, u.IsAdmin, actual.IsAdmin)
	assert.Equal(t, u.CreatedAt, actual.CreatedAt)
	assert.Equal(t, u.CreatedBy, actual.CreatedBy)
	assert.Equal(t, u.UpdatedAt, actual.UpdatedAt)
	assert.Equal(t, u.UpdatedBy, actual.UpdatedBy)
	assert.Equal(t, u.Version+1, actual.Version)
}

func TestDAOUpdate_OptimisticLockErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE users").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	var expected OptimisticLockErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "modified")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", version))
}

func TestDAOUpdate_EmailAlreadyInUseErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error", Constraint: "users_email_uk"}
	md.ExpectBegin()
	md.ExpectExec("UPDATE users").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.NotNil(t, err)
	var dupEmailErr EmailAlreadyInUseErr
	assert.True(t, errors.As(err, &dupEmailErr))
	assert.Contains(t, dupEmailErr.Error(), "email")
	assert.Contains(t, dupEmailErr.Error(), "in use")
	assert.Contains(t, dupEmailErr.Error(), email)
}

func TestDAOUpdate_OtherPQErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectBegin()
	md.ExpectExec("UPDATE users").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAOUpdate_OtherNonPQErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	mockErr := errors.New("unit-test mock error")
	md.ExpectBegin()
	md.ExpectExec("UPDATE users").
		WillReturnError(mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, mockErr, err)
}

func TestDAOUpdate_TooManyRowsAffected(t *testing.T) {
	d, db, md := initDAO()

	u := user.User{
		ID:        id,
		OrgID:     orgID,
		Name:      name,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: createdAt,
		CreatedBy: createdBy,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Version:   version,
	}

	md.ExpectBegin()
	md.ExpectExec("UPDATE users").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	_, err = d.Update(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "unexpected number of rows affected")
}

func TestDAODelete(t *testing.T) {
	d, db, md := initDAO()

	u := user.DeleteUser{
		ID:      id,
		Version: version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM users").
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Nil(t, err)
}

func TestDAODelete_OptimisticLockErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.DeleteUser{
		ID:      id,
		Version: version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM users").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	var expected OptimisticLockErr
	assert.True(t, errors.As(err, &expected))
	assert.Contains(t, err.Error(), "modified")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", version))
}

func TestDAODelete_OtherErr(t *testing.T) {
	d, db, md := initDAO()

	u := user.DeleteUser{
		ID:      id,
		Version: version,
	}

	mockErr := pq.Error{Message: "unit-test mock error"}
	md.ExpectBegin()
	md.ExpectExec("DELETE FROM users").
		WillReturnError(&mockErr)

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Equal(t, &mockErr, err)
}

func TestDAODelete_TooManyRowsAffected(t *testing.T) {
	d, db, md := initDAO()

	u := user.DeleteUser{
		ID:      id,
		Version: version,
	}

	md.ExpectBegin()
	md.ExpectExec("DELETE FROM users").
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := db.Beginx()
	assert.Nil(t, err)

	err = d.Delete(ctx, tx, u)

	assert.Nil(t, md.ExpectationsWereMet())
	assert.Contains(t, err.Error(), "unexpected number of rows affected")
}
