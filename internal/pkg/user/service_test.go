package user

import (
	"context"
	"errors"
	"github.com/RyanBard/gin-ex/internal/pkg/org"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type mockOrgSVC struct {
	mock.Mock
}

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

func initSVC() (s *service, ms *mockOrgSVC, md *mockDAO, mm *mockTXManager, mt *mockTimer, mi *mockIDGen) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.PanicLevel)
	ms = new(mockOrgSVC)
	md = new(mockDAO)
	mm = new(mockTXManager)
	mt = new(mockTimer)
	mi = new(mockIDGen)
	s = NewService(log, ms, md, mm, mt, mi)
	return s, ms, md, mm, mt, mi
}

func TestSVCGetByID(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	id := "foo-id"

	mockRes := User{
		ID:    id,
		Name:  "foo-name",
		Email: "foo@bar.com",
	}
	md.On("GetByID", ctx, id).Return(mockRes, nil)

	actual, err := s.GetByID(ctx, id)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetByID_DAOErr(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	id := "foo-id"

	mockErr := errors.New("unit-test mock error")
	md.On("GetByID", ctx, id).Return(User{}, mockErr)

	actual, err := s.GetByID(ctx, id)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, User{}, actual)
}

func TestSVCGetAll(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()

	mockRes := []User{
		{
			ID:    "foo-id",
			Name:  "foo-name",
			Email: "foo@bar.com",
		},
	}
	md.On("GetAll", ctx).Return(mockRes, nil)

	actual, err := s.GetAll(ctx)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetAll_DAOErr(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()

	mockErr := errors.New("unit-test mock error")
	md.On("GetAll", ctx).Return([]User{}, mockErr)

	actual, err := s.GetAll(ctx)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []User{}, actual)
}

func TestSVCGetAllByOrgID(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	orgID := "foo-org-id"

	mockRes := []User{
		{
			ID:    "foo-id",
			OrgID: orgID,
			Name:  "foo-name",
			Email: "foo@bar.com",
		},
	}
	md.On("GetAllByOrgID", ctx, orgID).Return(mockRes, nil)

	actual, err := s.GetAllByOrgID(ctx, orgID)

	assert.Nil(t, err)
	assert.Equal(t, mockRes, actual)
}

func TestSVCGetAllByOrgID_DAOErr(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	orgID := "foo-org-id"

	mockErr := errors.New("unit-test mock error")
	md.On("GetAllByOrgID", ctx, orgID).Return([]User{}, mockErr)

	actual, err := s.GetAllByOrgID(ctx, orgID)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, []User{}, actual)
}

func TestSVCSave_NoID(t *testing.T) {
	s, ms, md, _, mt, mi := initSVC()

	ctx := context.Background()
	u := User{
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		CreatedAt: time.UnixMilli(100),
		CreatedBy: "ignored",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   99,
	}

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID}, nil)

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	id := "foo-id"
	mi.On("GenID").Return(id)

	expectedUser := User{
		ID:        id,
		OrgID:     u.OrgID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: now,
		CreatedBy: "TODO",
		UpdatedAt: now,
		UpdatedBy: "TODO",
		Version:   1,
	}
	var expectedTX *sqlx.Tx
	md.On("Create", ctx, expectedTX, expectedUser).Return(nil)

	actual, err := s.Save(ctx, u)

	assert.Nil(t, err)
	assert.Equal(t, expectedUser, actual)
}

func TestSVCSave_NoID_OrgNotFound(t *testing.T) {
	s, ms, _, _, _, _ := initSVC()

	ctx := context.Background()
	u := User{
		OrgID: "foo-org-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	mockErr := errors.New("unit-test org not found")
	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{}, mockErr)

	actual, err := s.Save(ctx, u)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_NoID_CannotAssociateSysOrg(t *testing.T) {
	s, ms, _, _, _, _ := initSVC()

	ctx := context.Background()
	u := User{
		OrgID: "foo-org-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID, IsSystem: true}, nil)

	actual, err := s.Save(ctx, u)

	assert.NotNil(t, err)
	var orgErr CannotAssociateSysOrgErr
	assert.True(t, errors.As(err, &orgErr))
	assert.Contains(t, err.Error(), "Cannot associate user")
	assert.Contains(t, err.Error(), "system org")
	assert.Contains(t, err.Error(), u.ID)
	assert.Contains(t, err.Error(), u.OrgID)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_NoID_DAOErr(t *testing.T) {
	s, ms, md, _, mt, mi := initSVC()

	ctx := context.Background()
	u := User{
		OrgID: "foo-org-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID}, nil)

	now := time.UnixMilli(100)
	mt.On("Now").Return(now)

	id := "foo-id"
	mi.On("GenID").Return(id)

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("Create", ctx, expectedTX, mock.Anything).Return(mockErr)

	actual, err := s.Save(ctx, u)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_ID(t *testing.T) {
	s, ms, md, _, mt, _ := initSVC()

	ctx := context.Background()
	u := User{
		ID:        "foo-id",
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID}, nil)

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID}, nil)

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	expectedUser := User{
		ID:        u.ID,
		OrgID:     u.OrgID,
		Name:      u.Name,
		Email:     u.Email,
		UpdatedAt: now,
		UpdatedBy: "TODO",
		Version:   u.Version,
	}
	var expectedTX *sqlx.Tx
	md.On("Update", ctx, expectedTX, expectedUser).Return(expectedUser, nil)

	actual, err := s.Save(ctx, u)

	assert.Nil(t, err)
	assert.Equal(t, expectedUser, actual)
}

func TestSVCSave_ID_OrgNotFound(t *testing.T) {
	s, ms, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := User{
		ID:        "foo-id",
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID}, nil)

	mockErr := errors.New("unit-test org not found")
	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{}, mockErr)

	actual, err := s.Save(ctx, u)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_ID_CannotAssociateSysOrg(t *testing.T) {
	s, ms, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := User{
		ID:        "foo-id",
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID}, nil)

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID, IsSystem: true}, nil)

	actual, err := s.Save(ctx, u)

	assert.NotNil(t, err)
	var orgErr CannotAssociateSysOrgErr
	assert.True(t, errors.As(err, &orgErr))
	assert.Contains(t, err.Error(), "Cannot associate user")
	assert.Contains(t, err.Error(), "system org")
	assert.Contains(t, err.Error(), u.ID)
	assert.Contains(t, err.Error(), u.OrgID)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_ID_CannotModifySysUser(t *testing.T) {
	s, ms, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := User{
		ID:        "foo-id",
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID, IsSystem: true}, nil)

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID}, nil)

	actual, err := s.Save(ctx, u)

	assert.NotNil(t, err)
	var sysUserErr CannotModifySysUserErr
	assert.True(t, errors.As(err, &sysUserErr))
	assert.Contains(t, err.Error(), "Cannot modify")
	assert.Contains(t, err.Error(), "system user")
	assert.Contains(t, err.Error(), u.ID)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_ID_UserNotFound(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := User{
		ID:        "foo-id",
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	mockErr := NotFoundErr{ID: u.ID}
	md.On("GetByID", ctx, u.ID).Return(User{}, mockErr)

	actual, err := s.Save(ctx, u)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, User{}, actual)
}

func TestSVCSave_ID_DAOUpdateErr(t *testing.T) {
	s, ms, md, _, mt, _ := initSVC()

	ctx := context.Background()
	u := User{
		ID:        "foo-id",
		OrgID:     "foo-org-id",
		Name:      "foo-name",
		Email:     "foo@bar.com",
		UpdatedAt: time.UnixMilli(100),
		UpdatedBy: "ignored",
		Version:   1,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID}, nil)

	ms.On("GetByID", ctx, u.OrgID).Return(org.Org{ID: u.OrgID}, nil)

	now := time.UnixMilli(200)
	mt.On("Now").Return(now)

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("Update", ctx, expectedTX, mock.Anything).Return(User{}, mockErr)

	actual, err := s.Save(ctx, u)

	assert.Equal(t, mockErr, err)
	assert.Equal(t, User{}, actual)
}

func TestSVCDelete(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := DeleteUser{
		ID:      "foo-id",
		Version: 2,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID}, nil)

	var expectedTX *sqlx.Tx
	md.On("Delete", ctx, expectedTX, u).Return(nil)

	err := s.Delete(ctx, u)
	assert.Nil(t, err)
}

func TestSVCDelete_UserNotFound(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := DeleteUser{
		ID:      "foo-id",
		Version: 2,
	}

	mockErr := NotFoundErr{ID: u.ID}
	md.On("GetByID", ctx, u.ID).Return(User{}, mockErr)

	err := s.Delete(ctx, u)
	assert.Equal(t, mockErr, err)
}

func TestSVCDelete_CannotModifySysUser(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := DeleteUser{
		ID:      "foo-id",
		Version: 2,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID, IsSystem: true}, nil)

	err := s.Delete(ctx, u)
	assert.NotNil(t, err)
	var sysUserErr CannotModifySysUserErr
	assert.True(t, errors.As(err, &sysUserErr))
	assert.Contains(t, err.Error(), "Cannot modify")
	assert.Contains(t, err.Error(), "system user")
	assert.Contains(t, err.Error(), u.ID)
}

func TestSVCDelete_DAOErr(t *testing.T) {
	s, _, md, _, _, _ := initSVC()

	ctx := context.Background()
	u := DeleteUser{
		ID:      "foo-id",
		Version: 2,
	}

	md.On("GetByID", ctx, u.ID).Return(User{ID: u.ID}, nil)

	var expectedTX *sqlx.Tx
	mockErr := errors.New("unit-test mock error")
	md.On("Delete", ctx, expectedTX, mock.Anything).Return(mockErr)

	err := s.Delete(ctx, u)
	assert.Equal(t, mockErr, err)
}

func (d *mockOrgSVC) GetByID(ctx context.Context, id string) (org.Org, error) {
	args := d.Called(ctx, id)
	return args.Get(0).(org.Org), args.Error(1)
}

func (d *mockDAO) GetByID(ctx context.Context, id string) (User, error) {
	args := d.Called(ctx, id)
	return args.Get(0).(User), args.Error(1)
}

func (d *mockDAO) GetAll(ctx context.Context) ([]User, error) {
	args := d.Called(ctx)
	return args.Get(0).([]User), args.Error(1)
}

func (d *mockDAO) GetAllByOrgID(ctx context.Context, orgID string) ([]User, error) {
	args := d.Called(ctx, orgID)
	return args.Get(0).([]User), args.Error(1)
}

func (d *mockDAO) Create(ctx context.Context, tx *sqlx.Tx, u User) error {
	args := d.Called(ctx, tx, u)
	return args.Error(0)
}

func (d *mockDAO) Update(ctx context.Context, tx *sqlx.Tx, u User) (User, error) {
	args := d.Called(ctx, tx, u)
	return args.Get(0).(User), args.Error(1)
}

func (d *mockDAO) Delete(ctx context.Context, tx *sqlx.Tx, u DeleteUser) error {
	args := d.Called(ctx, tx, u)
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
