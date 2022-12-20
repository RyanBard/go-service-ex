package org

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type mockDAO struct {
	mock.Mock
}

type mockTXManager struct {
	mock.Mock
}

type mockTimer struct {
	mock.Mock
}

type mockIDGen struct {
	mock.Mock
}

func initSVC() (s *service, md *mockDAO, mm *mockTXManager, mt *mockTimer, mi *mockIDGen) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.PanicLevel)
	md = new(mockDAO)
	mm = new(mockTXManager)
	mt = new(mockTimer)
	mi = new(mockIDGen)
	s = NewService(log, md, mm, mt, mi)
	return s, md, mm, mt, mi
}

func TestSVCGetByID(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	id := "foo-id"

	mockRes := Org{
		ID:   id,
		Name: "foo-name",
		Desc: "foo-desc",
	}
	md.On("GetByID", ctx, id).Return(mockRes, nil)

	actual, err := s.GetByID(ctx, id)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetByID_DAOErr(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	id := "foo-id"

	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, id).Return(Org{}, mockErr)

	actual, err := s.GetByID(ctx, id)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCGetAll(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	name := ""

	mockRes := []Org{
		{
			ID:   "foo-id",
			Name: "foo-name",
			Desc: "foo-desc",
		},
	}
	md.On("GetAll", ctx).Return(mockRes, nil)

	actual, err := s.GetAll(ctx, name)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetAll_DAOErr(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	name := ""

	mockErr := errors.New("unit-test mock error")
	md.On("GetAll", ctx).Return([]Org{}, mockErr)

	actual, err := s.GetAll(ctx, name)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []Org{}, actual)
}

func TestSVCGetAll_NameSpecified(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	name := "foo"

	mockRes := []Org{
		{
			ID:   "foo-id",
			Name: "foo-name",
			Desc: "foo-desc",
		},
	}
	md.On("SearchByName", ctx, name).Return(mockRes, nil)

	actual, err := s.GetAll(ctx, name)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetAll_NameSpecified_DAOErr(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	name := "foo"

	mockErr := errors.New("unit-test mock error")
	md.On("SearchByName", ctx, name).Return([]Org{}, mockErr)

	actual, err := s.GetAll(ctx, name)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []Org{}, actual)
}

func TestSVCSave_NoID(t *testing.T) {
	s, md, _, mt, mi := initSVC()

	ctx := context.Background()
	o := Org{
		Name:      "foo-name",
		Desc:      "foo-desc",
		CreatedAt: time.UnixMilli(100),
		CreatedBy: "ignored",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   99,
	}

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	id := "foo-id"
	mi.On("GenID").Return(id)

	expectedOrg := Org{
		ID:        id,
		Name:      o.Name,
		Desc:      o.Desc,
		CreatedAt: now,
		CreatedBy: "TODO",
		UpdatedAt: now,
		UpdatedBy: "TODO",
		Version:   1,
	}
	var expectedTX *sqlx.Tx
	md.On("Create", ctx, expectedTX, expectedOrg).Return(nil)

	actual, err := s.Save(ctx, o)

	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, actual)
}

func TestSVCSave_NoID_DAOErr(t *testing.T) {
	s, md, _, mt, mi := initSVC()

	ctx := context.Background()
	o := Org{
		Name: "foo-name",
		Desc: "foo-desc",
	}

	now := time.UnixMilli(100)
	mt.On("Now").Return(now)

	id := "foo-id"
	mi.On("GenID").Return(id)

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("Create", ctx, expectedTX, mock.Anything).Return(mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCSave_ID(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	ctx := context.Background()
	o := Org{
		ID:        "foo-id",
		Name:      "foo-name",
		Desc:      "foo-desc",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	expectedOrg := Org{
		ID:        o.ID,
		Name:      o.Name,
		Desc:      o.Desc,
		UpdatedAt: now,
		UpdatedBy: "TODO",
		Version:   o.Version,
	}
	var expectedTX *sqlx.Tx
	md.On("GetByID", ctx, o.ID).Return(Org{}, nil)
	md.On("Update", ctx, expectedTX, expectedOrg).Return(expectedOrg, nil)

	actual, err := s.Save(ctx, o)

	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, actual)
}

func TestSVCSave_ID_DAOUpdateErr(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	ctx := context.Background()
	o := Org{
		ID:        "foo-id",
		Name:      "foo-name",
		Desc:      "foo-desc",
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(100),
		Version:   1,
	}

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, o.ID).Return(Org{}, nil)
	md.On("Update", ctx, expectedTX, mock.Anything).Return(Org{}, mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCSave_ID_OrgNotFound(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	ctx := context.Background()
	o := Org{
		ID:        "foo-id",
		Name:      "foo-name",
		Desc:      "foo-desc",
		IsSystem:  false,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(100),
		Version:   1,
	}

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	mockErr := NotFoundErr{
		ID: o.ID,
	}
	md.On("GetByID", ctx, o.ID).Return(Org{}, mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCSave_ID_SystemOrgNotAllowed(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	ctx := context.Background()
	o := Org{
		ID:        "foo-id",
		Name:      "foo-name",
		Desc:      "foo-desc",
		IsSystem:  false,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(100),
		Version:   1,
	}

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	md.On("GetByID", ctx, o.ID).Return(Org{IsSystem: true}, nil)

	actual, err := s.Save(ctx, o)

	assert.NotNil(t, err)
	var sysOrgErr CannotModifySysOrgErr
	assert.True(t, errors.As(err, &sysOrgErr))
	assert.Contains(t, err.Error(), "Cannot modify")
	assert.Contains(t, err.Error(), "system org")
	assert.Contains(t, err.Error(), o.ID)
	assert.Equal(t, Org{}, actual)
}

func TestSVCDelete(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	o := DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	var expectedTX *sqlx.Tx
	md.On("GetByID", ctx, o.ID).Return(Org{}, nil)
	md.On("Delete", ctx, expectedTX, o).Return(nil)

	err := s.Delete(ctx, o)
	assert.Nil(t, err)
}

func TestSVCDelete_DAOErr(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	o := DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, o.ID).Return(Org{}, nil)
	md.On("Delete", ctx, expectedTX, mock.Anything).Return(mockErr)

	err := s.Delete(ctx, o)
	assert.Equal(t, mockErr, err)
}

func TestSVCDelete_OrgNotFound(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	o := DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	mockErr := NotFoundErr{
		ID: o.ID,
	}
	md.On("GetByID", ctx, o.ID).Return(Org{}, mockErr)

	err := s.Delete(ctx, o)
	assert.Equal(t, mockErr, err)
}

func TestSVCDelete_SystemOrgNotAllowed(t *testing.T) {
	s, md, _, _, _ := initSVC()

	ctx := context.Background()
	o := DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	md.On("GetByID", ctx, o.ID).Return(Org{IsSystem: true}, nil)

	err := s.Delete(ctx, o)
	assert.NotNil(t, err)
	var sysOrgErr CannotModifySysOrgErr
	assert.True(t, errors.As(err, &sysOrgErr))
	assert.Contains(t, err.Error(), "Cannot modify")
	assert.Contains(t, err.Error(), "system org")
	assert.Contains(t, err.Error(), o.ID)
}

func (d *mockDAO) GetByID(ctx context.Context, id string) (Org, error) {
	args := d.Called(ctx, id)
	return args.Get(0).(Org), args.Error(1)
}

func (d *mockDAO) GetAll(ctx context.Context) ([]Org, error) {
	args := d.Called(ctx)
	return args.Get(0).([]Org), args.Error(1)
}

func (d *mockDAO) SearchByName(ctx context.Context, name string) ([]Org, error) {
	args := d.Called(ctx, name)
	return args.Get(0).([]Org), args.Error(1)
}

func (d *mockDAO) Create(ctx context.Context, tx *sqlx.Tx, o Org) error {
	args := d.Called(ctx, tx, o)
	return args.Error(0)
}

func (d *mockDAO) Update(ctx context.Context, tx *sqlx.Tx, o Org) (Org, error) {
	args := d.Called(ctx, tx, o)
	return args.Get(0).(Org), args.Error(1)
}

func (d *mockDAO) Delete(ctx context.Context, tx *sqlx.Tx, o DeleteOrg) error {
	args := d.Called(ctx, tx, o)
	return args.Error(0)
}

func (m *mockTXManager) Do(ctx context.Context, tx *sqlx.Tx, f func(tx *sqlx.Tx) error) error {
	return f(nil)
}

func (t *mockTimer) Now() time.Time {
	args := t.Called()
	return args.Get(0).(time.Time)
}

func (t *mockIDGen) GenID() string {
	args := t.Called()
	return args.String(0)
}
