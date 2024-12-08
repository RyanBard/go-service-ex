package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Payload struct {
	Foo string `json:"foo"`
}

type errReader int

const mockIOErrMsg = "unit-test mock io error"

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New(mockIOErrMsg)
}

func (errReader) Close() error {
	return nil
}

func initClient(f func(w http.ResponseWriter, r *http.Request)) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(f))
	client := NewClient(http.Client{})
	return client, server
}

func TestNewHTTPError_IOErr(t *testing.T) {
	statusCode := 400
	err := newHTTPError(statusCode, errReader(0))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", statusCode))
	var httpErr HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, statusCode, httpErr.StatusCode)
	assert.Equal(t, "", httpErr.ErrMessage)
}

func TestNewHTTPError_StrResp(t *testing.T) {
	statusCode := 400
	errMsg := "unit-test err message"
	err := newHTTPError(statusCode, io.NopCloser(strings.NewReader(errMsg)))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", statusCode))
	assert.Contains(t, err.Error(), errMsg)
	var httpErr HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, statusCode, httpErr.StatusCode)
	assert.Equal(t, errMsg, httpErr.ErrMessage)
}

func TestNewHTTPError_EmptyStrResp(t *testing.T) {
	statusCode := 400
	errMsg := ""
	err := newHTTPError(statusCode, io.NopCloser(strings.NewReader(errMsg)))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", statusCode))
	assert.Contains(t, err.Error(), errMsg)
	var httpErr HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, statusCode, httpErr.StatusCode)
	assert.Equal(t, errMsg, httpErr.ErrMessage)
}

func TestNewHTTPError_JSONRespWithMessage(t *testing.T) {
	statusCode := 400
	errMsg := "unit-test err message"
	jsonStr := fmt.Sprintf(`{"message": "%s"}`, errMsg)
	err := newHTTPError(statusCode, io.NopCloser(strings.NewReader(jsonStr)))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", statusCode))
	assert.Contains(t, err.Error(), errMsg)
	var httpErr HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, statusCode, httpErr.StatusCode)
	assert.Equal(t, errMsg, httpErr.ErrMessage)
}

func TestNewHTTPError_JSONRespWithError(t *testing.T) {
	statusCode := 400
	errMsg := "unit-test err message"
	jsonStr := fmt.Sprintf(`{"error": "%s"}`, errMsg)
	err := newHTTPError(statusCode, io.NopCloser(strings.NewReader(jsonStr)))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", statusCode))
	assert.Contains(t, err.Error(), errMsg)
	var httpErr HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, statusCode, httpErr.StatusCode)
	assert.Equal(t, errMsg, httpErr.ErrMessage)
}

func TestNewHTTPError_JSONResp(t *testing.T) {
	statusCode := 400
	errMsg := "unit-test err message"
	jsonStr := fmt.Sprintf(`{"other": "%s"}`, errMsg)
	err := newHTTPError(statusCode, io.NopCloser(strings.NewReader(jsonStr)))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%d", statusCode))
	var httpErr HTTPError
	assert.True(t, errors.As(err, &httpErr))
	assert.Equal(t, statusCode, httpErr.StatusCode)
	assert.Equal(t, "", httpErr.ErrMessage)
}

func TestIsSuccess(t *testing.T) {
	assert.True(t, isSuccess(200))
	assert.True(t, isSuccess(201))
	assert.True(t, isSuccess(202))
	assert.True(t, isSuccess(204))
	assert.False(t, isSuccess(400))
	assert.False(t, isSuccess(401))
	assert.False(t, isSuccess(403))
	assert.False(t, isSuccess(404))
	assert.False(t, isSuccess(409))
	assert.False(t, isSuccess(500))
	assert.False(t, isSuccess(501))
	assert.False(t, isSuccess(502))
	assert.False(t, isSuccess(503))
	assert.False(t, isSuccess(504))
}

func TestRenderPath(t *testing.T) {
	inputPath := "/foo/:fooID/bar/:barID/baz/:bazID"
	pathParams := map[string]string{
		"fooID": "a/bc",
		"barID": "def",
		"bazID": "ghi",
	}
	path, escapedPath, err := renderPath(inputPath, pathParams)
	assert.Nil(t, err)
	assert.Equal(t, "/foo/a/bc/bar/def/baz/ghi", path)
	assert.Equal(t, "/foo/a%2Fbc/bar/def/baz/ghi", escapedPath)
}

func TestRenderPath_TrailingSlash(t *testing.T) {
	inputPath := "/foo/:fooID/bar/:barID/baz/:bazID/"
	pathParams := map[string]string{
		"fooID": "a/bc",
		"barID": "def",
		"bazID": "ghi",
	}
	path, escapedPath, err := renderPath(inputPath, pathParams)
	assert.Nil(t, err)
	assert.Equal(t, "/foo/a/bc/bar/def/baz/ghi/", path)
	assert.Equal(t, "/foo/a%2Fbc/bar/def/baz/ghi/", escapedPath)
}

func TestRenderPath_ParamNotFound(t *testing.T) {
	inputPath := "/foo/:fooID/bar/:barID/baz/:bazID"
	pathParams := map[string]string{
		"barID": "def",
		"bazID": "ghi",
	}
	path, escapedPath, err := renderPath(inputPath, pathParams)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fooID")
	assert.Contains(t, err.Error(), "not found")
	assert.Equal(t, "", path)
	assert.Equal(t, "", escapedPath)
}

func TestWithQueryParams(t *testing.T) {
	hc, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	hb := hc.newBuilder("", "", map[string]string{}).
		WithQueryParams(map[string][]string{"foo": {"bar"}})
	assert.Len(t, hb.queryParams, 1)
	assert.Len(t, hb.queryParams["foo"], 1)
	assert.Equal(t, "bar", hb.queryParams["foo"][0])
}

func TestWithHeaders(t *testing.T) {
	hc, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	hb := hc.newBuilder("", "", map[string]string{}).
		WithHeaders(map[string][]string{"foo": {"bar"}})
	assert.Len(t, hb.headers, 1)
	assert.Len(t, hb.headers["foo"], 1)
	assert.Equal(t, "bar", hb.headers["foo"][0])
}

func TestWithAccept(t *testing.T) {
	hc, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	hb := hc.newBuilder("", "", map[string]string{}).
		WithAccept("foo")
	assert.Equal(t, "foo", hb.accept)
}

func TestWithContentType(t *testing.T) {
	hc, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	hb := hc.newBuilder("", "", map[string]string{}).
		WithContentType("foo")
	assert.Equal(t, "foo", hb.contentType)
}

func TestWithBody_Str(t *testing.T) {
	hc, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	hb := hc.newBuilder("", "", map[string]string{}).
		WithBody("foo")
	assert.Equal(t, "foo", hb.body)
}

func TestWithBody_Struct(t *testing.T) {
	foo := Payload{
		Foo: "foobar",
	}
	hc, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	hb := hc.newBuilder("", "", map[string]string{}).
		WithBody(foo)
	assert.Equal(t, foo, hb.body)
}

func TestRetrieveStr(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockResp))
	})
	var s string
	statusCode, err := client.Get(server.URL, map[string]string{}).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, mockResp, s)
}

func TestRetrieveStrWithContext_TimeoutErr(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte(mockResp))
	})
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Millisecond)
	var s string
	_, err := client.Get(server.URL, map[string]string{}).
		RetrieveStrWithContext(ctx, &s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestRetrieveStr_Head(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{})
	})
	var s string
	statusCode, err := client.Head(server.URL, map[string]string{}).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, "", s)
}

func TestRetrieveStr_NoContent(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	var s string
	statusCode, err := client.Get(server.URL, map[string]string{}).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 204, statusCode)
	assert.Equal(t, "", s)
}

func TestRetrieveStr_NilOut(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockResp))
	})
	statusCode, err := client.Get(server.URL, map[string]string{}).
		RetrieveStr(nil)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
}

func TestRetrieve(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockResp))
	})
	var r Payload
	statusCode, err := client.Get(server.URL, map[string]string{}).
		Retrieve(&r)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, "bar", r.Foo)
}

func TestRetrieveWithContext_TimeoutErr(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte(mockResp))
	})
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Millisecond)
	var r Payload
	_, err := client.Get(server.URL, map[string]string{}).
		RetrieveWithContext(ctx, &r)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestRetrieve_Head(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{})
	})
	var r Payload
	statusCode, err := client.Head(server.URL, map[string]string{}).
		Retrieve(&r)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, "", r.Foo)
}

func TestRetrieve_NoContent(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	var r Payload
	statusCode, err := client.Get(server.URL, map[string]string{}).
		Retrieve(&r)
	assert.Nil(t, err)
	assert.Equal(t, 204, statusCode)
	assert.Equal(t, "", r.Foo)
}

func TestRetrieve_NilOut(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockResp))
	})
	statusCode, err := client.Get(server.URL, map[string]string{}).
		Retrieve(nil)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
}

func TestHead(t *testing.T) {
	path := "/api/foo"
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte{}, b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "x=1&x=2&x=3&y=4&z=5", r.URL.RawQuery)
		assert.Equal(t, []string{"1", "2", "3"}, r.Header.Values("x-foo"))
		assert.Equal(t, []string{"4"}, r.Header.Values("x-bar"))
		assert.Equal(t, []string{"5"}, r.Header.Values("x-baz"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/json", r.Header.Get("content-Type"))
		w.WriteHeader(204)
	})
	var s string
	statusCode, err := client.Head(server.URL+path, map[string]string{}).
		WithQueryParams(map[string][]string{
			"x": {"1", "2", "3"},
			"y": {"4"},
			"z": {"5"},
		}).
		WithHeaders(map[string][]string{
			"x-foo": {"1", "2", "3"},
			"x-bar": {"4"},
			"x-baz": {"5"},
		}).
		WithAccept("application/json").
		WithContentType("application/json").
		// Body should not be sent on HEAD requests, so we're verifying this is ignored
		WithBody("foobarbaz").
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 204, statusCode)
	assert.Equal(t, "", s)
}

func TestGet(t *testing.T) {
	path := "/api/foo"
	mockResp := `{"foo":"bar"}`
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte{}, b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "x=1&x=2&x=3&y=4&z=5", r.URL.RawQuery)
		assert.Equal(t, []string{"1", "2", "3"}, r.Header.Values("x-foo"))
		assert.Equal(t, []string{"4"}, r.Header.Values("x-bar"))
		assert.Equal(t, []string{"5"}, r.Header.Values("x-baz"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/json", r.Header.Get("content-type"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	var s string
	statusCode, err := client.Get(server.URL+path, map[string]string{}).
		WithQueryParams(map[string][]string{
			"x": {"1", "2", "3"},
			"y": {"4"},
			"z": {"5"},
		}).
		WithHeaders(map[string][]string{
			"x-foo": {"1", "2", "3"},
			"x-bar": {"4"},
			"x-baz": {"5"},
		}).
		WithAccept("application/json").
		WithContentType("application/json").
		// Body should not be sent on HEAD requests, so we're verifying this is ignored
		WithBody("foobarbaz").
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, mockResp, s)
}

func TestPost(t *testing.T) {
	path := "/api/foo"
	mockResp := `{"foo":"bar"}`
	inputBody := "foobarbaz"
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(inputBody), b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "x=1&x=2&x=3&y=4&z=5", r.URL.RawQuery)
		assert.Equal(t, []string{"1", "2", "3"}, r.Header.Values("x-foo"))
		assert.Equal(t, []string{"4"}, r.Header.Values("x-bar"))
		assert.Equal(t, []string{"5"}, r.Header.Values("x-baz"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "text/plain", r.Header.Get("content-type"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	var s string
	statusCode, err := client.Post(server.URL+path, map[string]string{}).
		WithQueryParams(map[string][]string{
			"x": {"1", "2", "3"},
			"y": {"4"},
			"z": {"5"},
		}).
		WithHeaders(map[string][]string{
			"x-foo": {"1", "2", "3"},
			"x-bar": {"4"},
			"x-baz": {"5"},
		}).
		WithAccept("application/json").
		WithContentType("text/plain").
		WithBody(inputBody).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, mockResp, s)
}

func TestPut(t *testing.T) {
	path := "/api/foo"
	mockResp := `{"foo":"bar"}`
	inputBody := "foobarbaz"
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(inputBody), b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "x=1&x=2&x=3&y=4&z=5", r.URL.RawQuery)
		assert.Equal(t, []string{"1", "2", "3"}, r.Header.Values("x-foo"))
		assert.Equal(t, []string{"4"}, r.Header.Values("x-bar"))
		assert.Equal(t, []string{"5"}, r.Header.Values("x-baz"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "text/plain", r.Header.Get("content-type"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	var s string
	statusCode, err := client.Put(server.URL+path, map[string]string{}).
		WithQueryParams(map[string][]string{
			"x": {"1", "2", "3"},
			"y": {"4"},
			"z": {"5"},
		}).
		WithHeaders(map[string][]string{
			"x-foo": {"1", "2", "3"},
			"x-bar": {"4"},
			"x-baz": {"5"},
		}).
		WithAccept("application/json").
		WithContentType("text/plain").
		WithBody(inputBody).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, mockResp, s)
}

func TestDelete(t *testing.T) {
	path := "/api/foo"
	mockResp := `{"foo":"bar"}`
	inputBody := "foobarbaz"
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(inputBody), b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "x=1&x=2&x=3&y=4&z=5", r.URL.RawQuery)
		assert.Equal(t, []string{"1", "2", "3"}, r.Header.Values("x-foo"))
		assert.Equal(t, []string{"4"}, r.Header.Values("x-bar"))
		assert.Equal(t, []string{"5"}, r.Header.Values("x-baz"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "text/plain", r.Header.Get("content-type"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	var s string
	statusCode, err := client.Delete(server.URL+path, map[string]string{}).
		WithQueryParams(map[string][]string{
			"x": {"1", "2", "3"},
			"y": {"4"},
			"z": {"5"},
		}).
		WithHeaders(map[string][]string{
			"x-foo": {"1", "2", "3"},
			"x-bar": {"4"},
			"x-baz": {"5"},
		}).
		WithAccept("application/json").
		WithContentType("text/plain").
		WithBody(inputBody).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, mockResp, s)
}

func TestJSONInput_Valid(t *testing.T) {
	mockResp := `{"foo":"bar"}`
	input := Payload{
		Foo: "foobarbaz",
	}
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(`{"foo":"foobarbaz"}`), b)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/json", r.Header.Get("content-type"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	var s string
	statusCode, err := client.Post(server.URL, map[string]string{}).
		WithAccept("application/json").
		WithContentType("application/json").
		WithBody(input).
		RetrieveStr(&s)
	assert.Nil(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, mockResp, s)
}

func TestJSONInput_Invalid(t *testing.T) {
	inputBody := func() {}
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {})
	var s string
	_, err := client.Post(server.URL, map[string]string{}).
		WithContentType("application/json").
		WithBody(inputBody).
		RetrieveStr(&s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsupported type")
}

func TestURI_Invalid(t *testing.T) {
	client, _ := initClient(func(w http.ResponseWriter, r *http.Request) {})
	var s string
	_, err := client.Post(":invalid-uri", map[string]string{}).
		RetrieveStr(&s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")
}

func TestPath_MissingPathParam(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {})
	var s string
	_, err := client.Post(server.URL+"/api/foo/:fooID/bar/:barID", map[string]string{"fooID": "123"}).
		RetrieveStr(&s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "barID")
	assert.Contains(t, err.Error(), "not found")
}

func Test_HTTPError(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	var s string
	statusCode, err := client.Post(server.URL, map[string]string{}).
		RetrieveStr(&s)
	assert.Equal(t, 401, statusCode)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestStructInput_NotJSONContentType(t *testing.T) {
	client, server := initClient(func(w http.ResponseWriter, r *http.Request) {})
	foo := Payload{
		Foo: "foobar",
	}
	var s string
	_, err := client.Post(server.URL, map[string]string{}).
		WithContentType("text/plain").
		WithBody(foo).
		RetrieveStr(&s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "contentType")
	assert.Contains(t, err.Error(), "body was not a string")
}
