package org

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

	mockRes := Org{
		ID:        id,
		Name:      "foo-name",
		Desc:      "foo-desc",
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(200),
	}
	ms.On("GetByID", mock.Anything, id).Return(mockRes, nil)

	c.GetByID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual Org
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
	ms.On("GetByID", mock.Anything, id).Return(Org{}, mockErr)

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
	ms.On("GetByID", mock.Anything, id).Return(Org{}, mockErr)

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

	mockRes := []Org{
		{
			ID:        "foo-id",
			Name:      "foo-name",
			Desc:      "foo-desc",
			CreatedAt: time.UnixMilli(100),
			UpdatedAt: time.UnixMilli(200),
		},
	}
	ms.On("GetAll", mock.Anything, "").Return(mockRes, nil)

	c.GetAll(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual []Org
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLGetAll_NameSpecified(t *testing.T) {
	name := "foo-name"

	c, ms := initCTRL()
	gc, w, err := ginCtx("/?name=" + name)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "name",
			Value: name,
		},
	}

	mockRes := []Org{
		{
			ID:        "foo-id",
			Name:      name,
			Desc:      "foo-desc",
			CreatedAt: time.UnixMilli(100),
			UpdatedAt: time.UnixMilli(200),
		},
	}
	ms.On("GetAll", mock.Anything, name).Return(mockRes, nil)

	c.GetAll(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual []Org
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
	ms.On("GetAll", mock.Anything, "").Return([]Org{}, mockErr)

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

func TestCTRLSave(t *testing.T) {
	o := Org{
		ID:   "body-foo-id",
		Name: "foo-name",
		Desc: "foo-desc",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	mockRes := Org{
		ID:        o.ID,
		Name:      o.Name,
		Desc:      o.Desc,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(200),
	}
	ms.On("Save", mock.Anything, o).Return(mockRes, nil)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual Org
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, mockRes, actual)
}

func TestCTRLSave_PathID(t *testing.T) {
	id := "foo-id"
	o := Org{
		ID:   "body-foo-id",
		Name: "foo-name",
		Desc: "foo-desc",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	expectedOrg := Org{
		ID:   id,
		Name: o.Name,
		Desc: o.Desc,
	}
	mockRes := Org{
		ID:        id,
		Name:      o.Name,
		Desc:      o.Desc,
		CreatedAt: time.UnixMilli(100),
		UpdatedAt: time.UnixMilli(200),
	}
	ms.On("Save", mock.Anything, expectedOrg).Return(mockRes, nil)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual Org
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
	o := Org{
		Name: "",
		Desc: "foo-desc",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
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
	o := Org{
		Name: "foo-name",
		Desc: "",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
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
	assert.Contains(t, errMsg, "Desc")
}

func TestCTRLSave_ValidationError_MissingNameAndDesc(t *testing.T) {
	o := Org{
		Name: "",
		Desc: "",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
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
	assert.Contains(t, errMsg, "Desc")
}

func TestCTRLSave_NotFoundError(t *testing.T) {
	o := Org{
		ID:   "body-foo-id",
		Name: "foo-name",
		Desc: "foo-desc",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	mockErr := NotFoundErr{ID: o.ID}
	ms.On("Save", mock.Anything, o).Return(Org{}, mockErr)

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

func TestCTRLSave_CannotModifySysOrgError(t *testing.T) {
	o := Org{
		ID:   "body-foo-id",
		Name: "foo-name",
		Desc: "foo-desc",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	mockErr := CannotModifySysOrgErr{ID: o.ID}
	ms.On("Save", mock.Anything, o).Return(Org{}, mockErr)

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

func TestCTRLSave_ServiceError(t *testing.T) {
	o := Org{
		ID:   "body-foo-id",
		Name: "foo-name",
		Desc: "foo-desc",
	}

	c, ms := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	mockErr := errors.New("unit-test mock service error")
	ms.On("Save", mock.Anything, o).Return(Org{}, mockErr)

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
	o := DeleteOrg{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: o.ID,
		},
	}

	ms.On("Delete", mock.Anything, o).Return(nil)

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
	o := DeleteOrg{
		ID: "foo-id",
	}

	c, _ := initCTRL()
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: o.ID,
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

func TestCTRLDelete_NotFoundErr(t *testing.T) {
	o := DeleteOrg{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: o.ID,
		},
	}

	mockErr := NotFoundErr{ID: o.ID}
	ms.On("Delete", mock.Anything, o).Return(mockErr)

	c.Delete(gc)
	assert.Equal(t, 204, gc.Writer.Status())
}

func TestCTRLDelete_CannotModifySysOrgErr(t *testing.T) {
	o := DeleteOrg{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: o.ID,
		},
	}

	mockErr := CannotModifySysOrgErr{ID: o.ID}
	ms.On("Delete", mock.Anything, o).Return(mockErr)

	c.Delete(gc)
	assert.Equal(t, 403, gc.Writer.Status())
}

func TestCTRLDelete_ServiceError(t *testing.T) {
	o := DeleteOrg{
		ID:      "body-foo-id",
		Version: 1,
	}

	c, ms := initCTRL()
	gc, _, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: o.ID,
		},
	}

	mockErr := errors.New("unit-test mock service error")
	ms.On("Delete", mock.Anything, o).Return(mockErr)

	c.Delete(gc)
	assert.Equal(t, 500, gc.Writer.Status())
}

func (m *mockSVC) GetByID(ctx context.Context, id string) (Org, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Org), args.Error(1)
}

func (m *mockSVC) GetAll(ctx context.Context, name string) ([]Org, error) {
	args := m.Called(ctx, name)
	return args.Get(0).([]Org), args.Error(1)
}

func (m *mockSVC) Save(ctx context.Context, o Org) (Org, error) {
	args := m.Called(ctx, o)
	return args.Get(0).(Org), args.Error(1)
}

func (m *mockSVC) Delete(ctx context.Context, o DeleteOrg) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}
