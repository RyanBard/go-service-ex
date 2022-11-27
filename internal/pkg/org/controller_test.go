package org

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

func TestCTRLGetByID(t *testing.T) {
	id := "foo-id"
	now := time.Now().UnixMilli()

	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	c.GetByID(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual Org
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, "Foo Name", actual.Name)
	assert.Equal(t, "Foo Desc", actual.Desc)
	assert.Equal(t, false, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestCTRLGetByID_ServiceError(t *testing.T) {
	// TODO
}

func TestCTRLGetAll(t *testing.T) {
	now := time.Now().UnixMilli()

	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
	gc, w, err := ginCtx("/")
	assert.Nil(t, err)

	c.GetAll(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actuals []Org
	err = json.Unmarshal(bytes, &actuals)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, 200, gc.Writer.Status())
	assert.NotEqual(t, "", actual.ID)
	assert.Equal(t, "Foo Name", actual.Name)
	assert.Equal(t, "Foo Desc", actual.Desc)
	assert.Equal(t, false, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestCTRLGetAll_NameSpecified(t *testing.T) {
	name := "foo-name"
	now := time.Now().UnixMilli()

	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
	gc, w, err := ginCtx("/?name=" + name)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "name",
			Value: name,
		},
	}

	c.GetAll(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actuals []Org
	err = json.Unmarshal(bytes, &actuals)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 1, len(actuals))
	actual := actuals[0]
	assert.Equal(t, 200, gc.Writer.Status())
	assert.NotEqual(t, "", actual.ID)
	assert.Equal(t, name+"x Name", actual.Name)
	assert.Equal(t, "Foo Desc", actual.Desc)
	assert.Equal(t, false, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestCTRLGetAll_ServiceError(t *testing.T) {
	// TODO
}

func TestCTRLSave(t *testing.T) {
	now := time.Now().UnixMilli()
	o := Org{
		ID:         "body-foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual Org
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, o.ID, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestCTRLSave_PathID(t *testing.T) {
	id := "foo-id"
	now := time.Now().UnixMilli()
	o := Org{
		ID:         "body-foo-id",
		Name:       "foo-name",
		Desc:       "foo-desc",
		IsArchived: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
	gc, w, err := ginCtxWithBody("/", o)
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	c.Save(gc)
	res := w.Result()
	bytes, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	var actual Org
	err = json.Unmarshal(bytes, &actual)
	assert.Nil(t, err)
	defer res.Body.Close()
	assert.Equal(t, 200, gc.Writer.Status())
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, o.Name, actual.Name)
	assert.Equal(t, o.Desc, actual.Desc)
	assert.Equal(t, o.IsArchived, actual.IsArchived)
	assert.GreaterOrEqual(t, actual.CreatedAt, now)
	assert.GreaterOrEqual(t, actual.UpdatedAt, now)
}

func TestCTRLSave_ReadError(t *testing.T) {
	// TODO
}

func TestCTRLSave_UnmarshalError(t *testing.T) {
	malformedBody := "{"
	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
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

func TestCTRLSave_ServiceError(t *testing.T) {
	// TODO
}

func TestCTRLDelete(t *testing.T) {
	id := "foo-id"

	d := NewOrgDAO(logrus.StandardLogger())
	s := NewOrgService(logrus.StandardLogger(), d)
	c := NewOrgController(logrus.StandardLogger(), s)
	gc, _, err := ginCtx("/")
	assert.Nil(t, err)

	gc.Params = []gin.Param{
		{
			Key:   "id",
			Value: id,
		},
	}

	c.Delete(gc)
	assert.Equal(t, 204, gc.Writer.Status())
}

func TestCTRLDelete_ServiceError(t *testing.T) {
	// TODO
}
