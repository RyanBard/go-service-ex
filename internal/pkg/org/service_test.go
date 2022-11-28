package org

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockOrgDAO struct {
	mock.Mock
}

type mockTimer struct {
	mock.Mock
}

type mockIDGen struct {
	mock.Mock
}

func initSVC() (s *service, md *mockOrgDAO, mt *mockTimer, mi *mockIDGen) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.PanicLevel)
	md = new(mockOrgDAO)
	mt = new(mockTimer)
	mi = new(mockIDGen)
	s = NewOrgService(log, md, mt, mi)
	return s, md, mt, mi
}

func TestSVCGetByID(t *testing.T) {
	s, md, _, _ := initSVC()

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
	s, md, _, _ := initSVC()

	ctx := context.Background()
	id := "foo-id"

	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, id).Return(Org{}, mockErr)

	actual, err := s.GetByID(ctx, id)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCGetAll(t *testing.T) {
	s, md, _, _ := initSVC()

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
	s, md, _, _ := initSVC()

	ctx := context.Background()
	name := ""

	mockErr := errors.New("unit-test mock error")
	md.On("GetAll", ctx).Return([]Org{}, mockErr)

	actual, err := s.GetAll(ctx, name)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []Org{}, actual)
}

func TestSVCGetAll_NameSpecified(t *testing.T) {
	s, md, _, _ := initSVC()

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
	s, md, _, _ := initSVC()

	ctx := context.Background()
	name := "foo"

	mockErr := errors.New("unit-test mock error")
	md.On("SearchByName", ctx, name).Return([]Org{}, mockErr)

	actual, err := s.GetAll(ctx, name)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []Org{}, actual)
}

func TestSVCSave_NoID(t *testing.T) {
	s, md, mt, mi := initSVC()

	ctx := context.Background()
	o := Org{
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
	}

	now := int64(100)
	mt.On("Now").Return(now)

	id := "foo-id"
	mi.On("GenID").Return(id)

	expectedOrg := Org{
		ID:         id,
		Name:       o.Name,
		Desc:       o.Desc,
		IsArchived: o.IsArchived,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	md.On("Create", ctx, expectedOrg).Return(expectedOrg, nil)

	actual, err := s.Save(ctx, o)

	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, actual)
}

func TestSVCSave_NoID_DAOErr(t *testing.T) {
	s, md, mt, mi := initSVC()

	ctx := context.Background()
	o := Org{
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
	}

	now := int64(100)
	mt.On("Now").Return(now)

	id := "foo-id"
	mi.On("GenID").Return(id)

	expectedOrg := Org{
		ID:         id,
		Name:       o.Name,
		Desc:       o.Desc,
		IsArchived: o.IsArchived,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	mockErr := errors.New("unit-test mock error")
	md.On("Create", ctx, expectedOrg).Return(Org{}, mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCSave_ID(t *testing.T) {
	s, md, mt, _ := initSVC()

	ctx := context.Background()
	o := Org{
		ID:         "foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		// TODO - change to look up from existingq
		CreatedAt: int64(100),
	}

	now := int64(200)
	mt.On("Now").Return(now)

	expectedOrg := Org{
		ID:         o.ID,
		Name:       o.Name,
		Desc:       o.Desc,
		IsArchived: o.IsArchived,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  now,
	}
	md.On("Update", ctx, expectedOrg).Return(expectedOrg, nil)

	actual, err := s.Save(ctx, o)

	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, actual)
}

func TestSVCSave_ID_DAOErr(t *testing.T) {
	s, md, mt, _ := initSVC()

	ctx := context.Background()
	o := Org{
		ID:         "foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		// TODO - change to look up from existingq
		CreatedAt: int64(100),
	}

	now := int64(200)
	mt.On("Now").Return(now)

	expectedOrg := Org{
		ID:         o.ID,
		Name:       o.Name,
		Desc:       o.Desc,
		IsArchived: o.IsArchived,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  now,
	}
	mockErr := errors.New("unit-test mock error")
	md.On("Update", ctx, expectedOrg).Return(Org{}, mockErr)

	actual, err := s.Save(ctx, o)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, Org{}, actual)
}

func TestSVCDelete(t *testing.T) {
	s, md, _, _ := initSVC()

	ctx := context.Background()
	id := "foo"

	md.On("Delete", ctx, id).Return(nil)

	err := s.Delete(ctx, id)
	assert.Nil(t, err)
}

func TestSVCDelete_DAOErr(t *testing.T) {
	s, md, _, _ := initSVC()

	ctx := context.Background()
	id := "foo"

	mockErr := errors.New("unit-test mock error")
	md.On("Delete", ctx, id).Return(mockErr)

	err := s.Delete(ctx, id)
	assert.Equal(t, mockErr, err)
}

func (d *mockOrgDAO) GetByID(ctx context.Context, id string) (Org, error) {
	args := d.Called(ctx, id)
	return args.Get(0).(Org), args.Error(1)
}

func (d *mockOrgDAO) GetAll(ctx context.Context) ([]Org, error) {
	args := d.Called(ctx)
	return args.Get(0).([]Org), args.Error(1)
}

func (d *mockOrgDAO) SearchByName(ctx context.Context, name string) ([]Org, error) {
	args := d.Called(ctx, name)
	return args.Get(0).([]Org), args.Error(1)
}

func (d *mockOrgDAO) Create(ctx context.Context, o Org) (Org, error) {
	args := d.Called(ctx, o)
	return args.Get(0).(Org), args.Error(1)
}

func (d *mockOrgDAO) Update(ctx context.Context, o Org) (Org, error) {
	args := d.Called(ctx, o)
	return args.Get(0).(Org), args.Error(1)
}

func (d *mockOrgDAO) Delete(ctx context.Context, id string) error {
	args := d.Called(ctx, id)
	return args.Error(0)
}

func (t *mockTimer) Now() int64 {
	args := t.Called()
	return args.Get(0).(int64)
}

func (t *mockIDGen) GenID() string {
	args := t.Called()
	return args.String(0)
}
