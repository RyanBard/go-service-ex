package tx

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
)

var (
	ctx = context.Background()
)

func initMGR() (m *txmgr, dbx *sqlx.DB, md sqlmock.Sqlmock) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)
	db, md, err := sqlmock.New()
	if err != nil {
		log.Fatal("failed to mock db")
	}
	dbx = sqlx.NewDb(db, "sqlmock")
	m = NewTXMGR(log, dbx)
	return m, dbx, md
}

func TestDo_CommitOnSuccess(t *testing.T) {
	m, _, md := initMGR()
	md.ExpectBegin()
	md.ExpectCommit()
	actual := m.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return nil
	})
	assert.Nil(t, actual)
	assert.Nil(t, md.ExpectationsWereMet())
}

func TestDo_RollbackOnError(t *testing.T) {
	m, _, md := initMGR()
	md.ExpectBegin()
	md.ExpectRollback()
	mockErr := errors.New("unit-test error")
	actual := m.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return mockErr
	})
	assert.Equal(t, mockErr, actual)
	assert.Nil(t, md.ExpectationsWereMet())
}

func TestDo_ErrorOnFailedToBegin(t *testing.T) {
	m, _, md := initMGR()
	actual := m.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return nil
	})
	assert.NotNil(t, actual)
	assert.Contains(t, actual.Error(), "call to database transaction Begin was not expected")
	assert.Nil(t, md.ExpectationsWereMet())
}

func TestDo_ErrorOnFailedToCommit(t *testing.T) {
	m, _, md := initMGR()
	md.ExpectBegin()
	actual := m.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return nil
	})
	assert.NotNil(t, actual)
	assert.Contains(t, actual.Error(), "call to Commit transaction was not expected")
	assert.Nil(t, md.ExpectationsWereMet())
}

func TestDo_OriginalErrOnFailedToRollback(t *testing.T) {
	m, _, md := initMGR()
	md.ExpectBegin()
	mockErr := errors.New("unit-test error")
	actual := m.Do(ctx, nil, func(tx *sqlx.Tx) error {
		return mockErr
	})
	assert.Equal(t, mockErr, actual)
	assert.Nil(t, md.ExpectationsWereMet())
}

func TestDo_JoinTX_NoCommitOnSuccess(t *testing.T) {
	m, dbx, md := initMGR()
	md.ExpectBegin()
	tx, err := dbx.Beginx()
	assert.Nil(t, err)
	actual := m.Do(ctx, tx, func(tx *sqlx.Tx) error {
		return nil
	})
	assert.Nil(t, actual)
	assert.Nil(t, md.ExpectationsWereMet())
}

func TestDo_JoinTX_NoRollbackOnError(t *testing.T) {
	m, dbx, md := initMGR()
	md.ExpectBegin()
	tx, err := dbx.Beginx()
	assert.Nil(t, err)
	mockErr := errors.New("unit-test error")
	actual := m.Do(ctx, tx, func(tx *sqlx.Tx) error {
		return mockErr
	})
	assert.Equal(t, mockErr, actual)
	assert.Nil(t, md.ExpectationsWereMet())
}
