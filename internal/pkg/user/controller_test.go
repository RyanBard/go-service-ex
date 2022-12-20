package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockSVC struct {
	mock.Mock
}

// https://stackoverflow.com/questions/45126312/how-do-i-test-an-error-on-reading-from-a-request-body
type errReader int

const mockIOErrMsg = "unit-test mock io error"

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New(mockIOErrMsg)
}

func initCTRL() (c *ctrl, ms *mockSVC) {
	log := logrus.StandardLogger()
	log.SetLevel(logrus.PanicLevel)
	v := validator.New()
	ms = new(mockSVC)
	c = NewController(log, v, ms)
	return c, ms
}

func ginCtx(url string) (*gin.Context, *httptest.ResponseRecorder, error) {
	return ginCtxWithBody(url, nil)
}

func ginCtxWithBody(url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder, error) {
	var bodyStr *string
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, httptest.NewRecorder(), err
		}
		jsonStr := string(jsonBytes)
		bodyStr = &jsonStr
	}
	return ginCtxWithStrBody(url, bodyStr)
}

func ginCtxWithStrBody(url string, body *string) (*gin.Context, *httptest.ResponseRecorder, error) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	u := url
	if url == "" {
		u = "/"
	}
	var rdr io.Reader
	if body != nil {
		rdr = strings.NewReader(*body)
	}
	req, err := http.NewRequest("POST", u, rdr)
	if err != nil {
		return nil, w, err
	}
	gc.Request = req
	return gc, w, nil
}

func ginCtxWithIOErr(url string) (*gin.Context, *httptest.ResponseRecorder, error) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	u := url
	if url == "" {
		u = "/"
	}
	req, err := http.NewRequest("POST", u, errReader(0))
	if err != nil {
		return nil, w, err
	}
	gc.Request = req
	return gc, w, nil
}

func TestCTRLGetByID(t *testing.T) {
	id := "foo-id"

	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	mockRes := User{
		ID:        id,
		Name:      "foo-name",
		Email:     "foo@bar.com",
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(200),
	}
	ms.On("GetByID", mock.Anything, id).Return(mockRes, nil)

	c.GetByID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual User
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLGetByID_NotFoundError(t *testing.T) {
	id := "foo-id"

	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	mockErr := NotFoundErr{ID: id}
	ms.On("GetByID", mock.Anything, id).Return(User{}, mockErr)

	c.GetByID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 404, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLGetByID_ServiceError(t *testing.T) {
	id := "foo-id"

	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	mockErr := errors.New("unit-test mock service error")
	ms.On("GetByID", mock.Anything, id).Return(User{}, mockErr)

	c.GetByID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 500, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLGetAll(t *testing.T) {
	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	mockRes := []User{
		{
			ID:        "foo-id",
			Name:      "foo-name",
			Email:     "foo@bar.com",
			CreatedAt: time.UnixMilli(100),
			UpdatedAt: time.UnixMilli(200),
		},
	}
	ms.On("GetAll", mock.Anything).Return(mockRes, nil)

	c.GetAll(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual []User
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLGetAll_ServiceError(t *testing.T) {
	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	mockErr := errors.New("unit-test mock service error")
	ms.On("GetAll", mock.Anything).Return([]User{}, mockErr)

	c.GetAll(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 500, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLGetAllByOrgID(t *testing.T) {
	orgID := "foo-id"
	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: orgID,
		},
	}

	mockRes := []User{
		{
			ID:        "foo-id",
			OrgID:     orgID,
			Name:      "foo-name",
			Email:     "foo@bar.com",
			CreatedAt: time.UnixMilli(100),
			UpdatedAt: time.UnixMilli(200),
		},
	}
	ms.On("GetAllByOrgID", mock.Anything, orgID).Return(mockRes, nil)

	c.GetAllByOrgID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual []User
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLGetAllByOrgID_ServiceError(t *testing.T) {
	orgID := "foo-id"
	c, ms := initCTRL()
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: orgID,
		},
	}

	mockErr := errors.New("unit-test mock service error")
	ms.On("GetAllByOrgID", mock.Anything, orgID).Return([]User{}, mockErr)

	c.GetAllByOrgID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 500, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLSave(t *testing.T) {
	u := User{
		ID:    "body-foo-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	mockRes := User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(200),
	}
	ms.On("Save", mock.Anything, u).Return(mockRes, nil)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual User
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLSave_PathID(t *testing.T) {
	id := "foo-id"
	u := User{
		ID:    "body-foo-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	expectedUser := User{
		ID:    id,
		Name:  u.Name,
		Email: u.Email,
	}
	mockRes := User{
		ID:        id,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(200),
	}
	ms.On("Save", mock.Anything, expectedUser).Return(mockRes, nil)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual User
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLSave_ReadError(t *testing.T) {
	c, _ := initCTRL()
	gc, w, err := ginCtxWithIOErr("/")
	assert.Nil(t, err)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	assert.Equal(t, mockIOErrMsg, actual["message"])
}

func TestCTRLSave_UnmarshalError(t *testing.T) {
	malformedBody := "{"

	c, _ := initCTRL()
	gc, w, err := ginCtxWithStrBody("/", &malformedBody)
	assert.Nil(t, err)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	assert.Equal(t, "unexpected end of JSON input", actual["message"])
}

func TestCTRLSave_ValidationError_MissingName(t *testing.T) {
	u := User{
		Name:  "",
		Email: "foo@bar.com",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	errMsg := actual["message"]
	assert.Contains(t, errMsg, "required")
	assert.Contains(t, errMsg, "Name")
}

func TestCTRLSave_ValidationError_MissingDesc(t *testing.T) {
	u := User{
		Name:  "foo-name",
		Email: "",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	errMsg := actual["message"]
	assert.Contains(t, errMsg, "required")
	assert.Contains(t, errMsg, "Email")
}

func TestCTRLSave_ValidationError_MissingNameAndDesc(t *testing.T) {
	u := User{
		Name:  "",
		Email: "",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	errMsg := actual["message"]
	assert.Contains(t, errMsg, "required")
	assert.Contains(t, errMsg, "Name")
	assert.Contains(t, errMsg, "Email")
}

func TestCTRLSave_NotFoundError(t *testing.T) {
	u := User{
		ID:    "body-foo-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	mockErr := NotFoundErr{ID: u.ID}
	ms.On("Save", mock.Anything, u).Return(User{}, mockErr)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 404, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLSave_CannotModifySysUserError(t *testing.T) {
	u := User{
		ID:    "body-foo-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	mockErr := CannotModifySysUserErr{ID: u.ID}
	ms.On("Save", mock.Anything, u).Return(User{}, mockErr)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 403, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLSave_CannotAssociateSysOrgError(t *testing.T) {
	u := User{
		ID:    "body-foo-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	mockErr := CannotAssociateSysOrgErr{UserID: u.ID, OrgID: u.OrgID}
	ms.On("Save", mock.Anything, u).Return(User{}, mockErr)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLSave_ServiceError(t *testing.T) {
	u := User{
		ID:    "body-foo-id",
		Name:  "foo-name",
		Email: "foo@bar.com",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	mockErr := errors.New("unit-test mock service error")
	ms.On("Save", mock.Anything, u).Return(User{}, mockErr)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 500, gc.Writer.Status())
	assert.Equal(t, mockErr.Error(), actual["message"])
}

func TestCTRLDelete(t *testing.T) {
	u := DeleteUser{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: u.ID,
		},
	}

	ms.On("Delete", mock.Anything, u).Return(nil)

	c.Delete(gc)
	assert.Equal(t, 204, gc.Writer.Status())
}

func TestCTRLDelete_ReadError(t *testing.T) {
	c, _ := initCTRL()
	gc, w, err := ginCtxWithIOErr("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: "foo-id",
		},
	}

	c.Delete(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	assert.Equal(t, mockIOErrMsg, actual["message"])
}

func TestCTRLDelete_UnmarshalError(t *testing.T) {
	malformedBody := "{"

	c, _ := initCTRL()
	gc, w, err := ginCtxWithStrBody("/", &malformedBody)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: "foo-id",
		},
	}

	c.Delete(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	assert.Equal(t, "unexpected end of JSON input", actual["message"])
}

func TestCTRLDelete_ValidationError_MissingVersion(t *testing.T) {
	u := DeleteUser{
		ID: "foo-id",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: u.ID,
		},
	}

	c.Delete(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual map[string]string
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 400, gc.Writer.Status())
	errMsg := actual["message"]
	assert.Contains(t, errMsg, "required")
	assert.Contains(t, errMsg, "Version")
}

func TestCTRLDelete_NotFoundError(t *testing.T) {
	u := DeleteUser{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: u.ID,
		},
	}

	mockErr := NotFoundErr{ID: u.ID}
	ms.On("Delete", mock.Anything, u).Return(mockErr)

	c.Delete(gc)
	assert.Equal(t, 404, gc.Writer.Status())
}

func TestCTRLDelete_CannotModifySysUserError(t *testing.T) {
	u := DeleteUser{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: u.ID,
		},
	}

	mockErr := CannotModifySysUserErr{ID: u.ID}
	ms.On("Delete", mock.Anything, u).Return(mockErr)

	c.Delete(gc)
	assert.Equal(t, 403, gc.Writer.Status())
}

func TestCTRLDelete_ServiceError(t *testing.T) {
	u := DeleteUser{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", u)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: u.ID,
		},
	}

	mockErr := errors.New("unit-test mock service error")
	ms.On("Delete", mock.Anything, u).Return(mockErr)

	c.Delete(gc)
	assert.Equal(t, 500, gc.Writer.Status())
}

func (m *mockSVC) GetByID(ctx context.Context, id string) (User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(User), args.Error(1)
}

func (m *mockSVC) GetAll(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]User), args.Error(1)
}

func (m *mockSVC) GetAllByOrgID(ctx context.Context, orgID string) ([]User, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]User), args.Error(1)
}

func (m *mockSVC) Save(ctx context.Context, u User) (User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(User), args.Error(1)
}

func (m *mockSVC) Delete(ctx context.Context, u DeleteUser) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
