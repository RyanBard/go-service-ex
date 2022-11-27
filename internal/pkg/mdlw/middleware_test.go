package mdlw

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ginContext(headers map[string]string) (*gin.Context, *httptest.ResponseRecorder, error) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		return nil, w, err
	}
	gc.Request = req
	for k, v := range headers {
		gc.Request.Header.Set(k, v)
	}
	return gc, w, nil
}

func TestReqID_NoHeader(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	mw := ReqID(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value("reqID")
	assert.Contains(t, actual, "generated-")
}

func TestReqID_Header(t *testing.T) {
	expected := "expected-req-id"

	gc, w, err := ginContext(map[string]string{"X-Request-Id": expected})
	assert.Nil(t, err)

	mw := ReqID(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value("reqID")
	assert.Equal(t, expected, actual)
}

func TestAuth_NoHeader(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	mw := Auth(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value("userID")
	assert.Nil(t, actual)
}

func TestAuth_NotBearer(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": "Basic foo"})
	assert.Nil(t, err)

	mw := Auth(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value("userID")
	assert.Nil(t, actual)
}

func TestAuth_MalformedTooManyParts(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": "Bearer Bearer foo"})
	assert.Nil(t, err)

	mw := Auth(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value("userID")
	assert.Nil(t, actual)
}

func TestAuth_InvalidToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": "Bearer bar"})
	assert.Nil(t, err)

	mw := Auth(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value("userID")
	assert.Nil(t, actual)
}

func TestAuth_ValidToken(t *testing.T) {
	expected := "TODO"

	gc, w, err := ginContext(map[string]string{"Authorization": "Bearer foo"})
	assert.Nil(t, err)

	mw := Auth(logrus.StandardLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value("userID")
	assert.Equal(t, expected, actual)
}
