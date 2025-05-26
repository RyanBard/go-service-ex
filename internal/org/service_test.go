package org

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	ctxutil "github.com/RyanBard/go-ctx-util/pkg"
	"github.com/RyanBard/go-service-ex/internal/testutil"
	"github.com/RyanBard/go-service-ex/pkg/org"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	log := testutil.GetLogger()
	md = new(mockDAO)
	mm = new(mockTXManager)
	mt = new(mockTimer)
	mi = new(mockIDGen)
	s = NewService(log, md, mm, mt, mi)
	return s, md, mm, mt, mi
}

func TestSVCGetByID(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	id := "foo-id"

	mockRes := org.Org{
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

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	id := "foo-id"

	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, id).Return(org.Org{}, mockErr)

	actual, err := s.GetByID(ctx, id)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, org.Org{}, actual)
}

func TestSVCGetAll(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	name := ""

	mockRes := []org.Org{
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

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	name := ""

	mockErr := errors.New("unit-test mock error")
	md.On("GetAll", ctx).Return([]org.Org{}, mockErr)

	actual, err := s.GetAll(ctx, name)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []org.Org{}, actual)
}

func TestSVCGetAll_NameSpecified(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	name := "foo"

	mockRes := []org.Org{
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

func TestSVCGetAll_UpperCaseNameSpecified(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	name := "FOO"

	mockRes := []org.Org{
		{
			ID:   "foo-id",
			Name: "foo-name",
			Desc: "foo-desc",
		},
	}
	md.On("SearchByName", ctx, strings.ToLower(name)).Return(mockRes, nil)

	actual, err := s.GetAll(ctx, name)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetAll_NameSpecified_DAOErr(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	name := "foo"

	mockErr := errors.New("unit-test mock error")
	md.On("SearchByName", ctx, name).Return([]org.Org{}, mockErr)

	actual, err := s.GetAll(ctx, name)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []org.Org{}, actual)
}

func TestSVCSave_NoID(t *testing.T) {
	s, md, _, mt, mi := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.Org{
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

	expectedOrg := org.Org{
		ID:        id,
		Name:      o.Name,
		Desc:      o.Desc,
		CreatedAt: now,
		CreatedBy: loggedInUserID,
		UpdatedAt: now,
		UpdatedBy: loggedInUserID,
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

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.Org{
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
	assert.Equal(t, org.Org{}, actual)
}

func TestSVCSave_ID(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.Org{
		ID:        "foo-id",
		Name:      "foo-name",
		Desc:      "foo-desc",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	expectedOrg := org.Org{
		ID:        o.ID,
		Name:      o.Name,
		Desc:      o.Desc,
		UpdatedAt: now,
		UpdatedBy: loggedInUserID,
		Version:   o.Version,
	}
	var expectedTX *sqlx.Tx
	md.On("GetByID", ctx, o.ID).Return(org.Org{}, nil)
	md.On("Update", ctx, expectedTX, expectedOrg).Return(expectedOrg, nil)

	actual, err := s.Save(ctx, o)

	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, actual)
}

func TestSVCSave_ID_DAOUpdateErr(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.Org{
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
	md.On("GetByID", ctx, o.ID).Return(org.Org{}, nil)
	md.On("Update", ctx, expectedTX, mock.Anything).Return(org.Org{}, mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, org.Org{}, actual)
}

func TestSVCSave_ID_OrgNotFound(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.Org{
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

	mockErr := ErrNotFound{
		ID: o.ID,
	}
	md.On("GetByID", ctx, o.ID).Return(org.Org{}, mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, org.Org{}, actual)
}

func TestSVCSave_ID_SystemOrgNotAllowed(t *testing.T) {
	s, md, _, mt, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.Org{
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

	md.On("GetByID", ctx, o.ID).Return(org.Org{IsSystem: true}, nil)

	actual, err := s.Save(ctx, o)

	assert.NotNil(t, err)
	var sysOrgErr ErrCannotModifySysOrg
	assert.True(t, errors.As(err, &sysOrgErr))
	assert.Contains(t, err.Error(), "Cannot modify")
	assert.Contains(t, err.Error(), "system org")
	assert.Contains(t, err.Error(), o.ID)
	assert.Equal(t, org.Org{}, actual)
}

func TestSVCSave_ErrIfNoAuditInfo(t *testing.T) {
	s, _, _, _, _ := initSVC()

	ctx := context.Background()
	o := org.Org{
		ID:        "foo-id",
		Name:      "foo-name",
		Desc:      "foo-desc",
		IsSystem:  false,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(100),
		Version:   1,
	}

	actual, err := s.Save(ctx, o)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "user not logged in")
	assert.Equal(t, org.Org{}, actual)
}

func TestSVCDelete(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	var expectedTX *sqlx.Tx
	md.On("GetByID", ctx, o.ID).Return(org.Org{}, nil)
	md.On("Delete", ctx, expectedTX, o).Return(nil)

	err := s.Delete(ctx, o)
	assert.Nil(t, err)
}

func TestSVCDelete_DAOErr(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, o.ID).Return(org.Org{}, nil)
	md.On("Delete", ctx, expectedTX, mock.Anything).Return(mockErr)

	err := s.Delete(ctx, o)
	assert.Equal(t, mockErr, err)
}

func TestSVCDelete_OrgNotFound(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	mockErr := ErrNotFound{
		ID: o.ID,
	}
	md.On("GetByID", ctx, o.ID).Return(org.Org{}, mockErr)

	err := s.Delete(ctx, o)
	assert.Equal(t, mockErr, err)
}

func TestSVCDelete_SystemOrgNotAllowed(t *testing.T) {
	s, md, _, _, _ := initSVC()

	loggedInUserID := "logged-in-user-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, loggedInUserID)
	o := org.DeleteOrg{
		ID:      "foo-id",
		Version: 2,
	}

	md.On("GetByID", ctx, o.ID).Return(org.Org{IsSystem: true}, nil)

	err := s.Delete(ctx, o)
	assert.NotNil(t, err)
	var sysOrgErr ErrCannotModifySysOrg
	assert.True(t, errors.As(err, &sysOrgErr))
	assert.Contains(t, err.Error(), "Cannot modify")
	assert.Contains(t, err.Error(), "system org")
	assert.Contains(t, err.Error(), o.ID)
}

func (d *mockDAO) GetByID(ctx context.Context, id string) (org.Org, error) {
	args := d.Called(ctx, id)
	return args.Get(0).(org.Org), args.Error(1)
}

func (d *mockDAO) GetAll(ctx context.Context) ([]org.Org, error) {
	args := d.Called(ctx)
	return args.Get(0).([]org.Org), args.Error(1)
}

func (d *mockDAO) SearchByName(ctx context.Context, name string) ([]org.Org, error) {
	args := d.Called(ctx, name)
	return args.Get(0).([]org.Org), args.Error(1)
}

func (d *mockDAO) Create(ctx context.Context, tx *sqlx.Tx, o org.Org) error {
	args := d.Called(ctx, tx, o)
	return args.Error(0)
}

func (d *mockDAO) Update(ctx context.Context, tx *sqlx.Tx, o org.Org) (org.Org, error) {
	args := d.Called(ctx, tx, o)
	return args.Get(0).(org.Org), args.Error(1)
}

func (d *mockDAO) Delete(ctx context.Context, tx *sqlx.Tx, o org.DeleteOrg) error {
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
