package apiclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/RyanBard/gin-ex/internal/httpx"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Payload struct {
	Foo string `json:"foo"`
}

func initClient(getToken func(isRetry bool) (string, error), f func(w http.ResponseWriter, r *http.Request)) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(f))
	client := NewClient(httpx.NewClient(http.Client{}), getToken)
	return client, server
}

func bearer(s string) string {
	return fmt.Sprintf("Bearer %s", s)
}

func TestGet(t *testing.T) {
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	path := "/api/foo"
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	mockResp := `{"foo":"bar"}`
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte{}, b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	err := client.Get(ctx, server.URL+path, pathParams, queryParams, &out)
	assert.Nil(t, err)
	assert.Equal(t, "bar", out.Foo)
}

func TestPost(t *testing.T) {
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	path := "/api/foo"
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	in := Payload{
		Foo: "foobar",
	}
	var out Payload
	mockResp := `{"foo":"bar"}`
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(`{"foo":"foobar"}`), b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/json", r.Header.Get("content-type"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	err := client.Post(ctx, server.URL+path, pathParams, queryParams, in, &out)
	assert.Nil(t, err)
	assert.Equal(t, "bar", out.Foo)
}

func TestPut(t *testing.T) {
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	path := "/api/foo"
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	in := Payload{
		Foo: "foobar",
	}
	var out Payload
	mockResp := `{"foo":"bar"}`
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(`{"foo":"foobar"}`), b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/json", r.Header.Get("content-type"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	err := client.Put(ctx, server.URL+path, pathParams, queryParams, in, &out)
	assert.Nil(t, err)
	assert.Equal(t, "bar", out.Foo)
}

func TestDelete(t *testing.T) {
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	path := "/api/foo"
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	in := Payload{
		Foo: "foobar",
	}
	var out Payload
	mockResp := `{"foo":"bar"}`
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(`{"foo":"foobar"}`), b)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/json", r.Header.Get("content-type"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	err := client.Delete(ctx, server.URL+path, pathParams, queryParams, in, &out)
	assert.Nil(t, err)
	assert.Equal(t, "bar", out.Foo)
}

func TestTokenErr(t *testing.T) {
	mockTokenErr := errors.New("unit-test token error")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	err := client.Get(ctx, server.URL, pathParams, queryParams, &out)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", out.Foo)
}

func TestNoReqID(t *testing.T) {
	i := 0
	getToken := func(isRetry bool) (string, error) {
		i++
		return fmt.Sprintf("test-token-%d", i), nil
	}
	ctx := context.Background()
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	mockResp := `{"foo":"bar"}`
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		assert.Equal(t, "", r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	err := client.Get(ctx, server.URL, pathParams, queryParams, &out)
	assert.Nil(t, err)
	assert.Equal(t, "bar", out.Foo)
}

func Test401Recover(t *testing.T) {
	expiredToken := "test-expired-token"
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		if isRetry {
			return token, nil
		}
		return expiredToken, nil
	}
	ctx := context.Background()
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	mockResp := `{"foo":"bar"}`
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Header.Get("authorization") == bearer(expiredToken) {
			w.WriteHeader(401)
			return
		}
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(mockResp))
	})
	err := client.Get(ctx, server.URL, pathParams, queryParams, &out)
	assert.Nil(t, err)
	assert.Equal(t, "bar", out.Foo)
}

func Test401Twice(t *testing.T) {
	getToken := func(isRetry bool) (string, error) {
		return "test-token", nil
	}
	ctx := context.Background()
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(401)
	})
	err := client.Get(ctx, server.URL, pathParams, queryParams, &out)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Equal(t, "", out.Foo)
}

func Test401ThenTokenErr(t *testing.T) {
	mockTokenErr := errors.New("unit-test token error")
	getToken := func(isRetry bool) (string, error) {
		if isRetry {
			return "", mockTokenErr
		}
		return "test-token", nil
	}
	ctx := context.Background()
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(401)
	})
	err := client.Get(ctx, server.URL, pathParams, queryParams, &out)
	assert.NotNil(t, err)
	assert.NotEqual(t, mockTokenErr, err)
	assert.Contains(t, err.Error(), "401")
	assert.Equal(t, "", out.Foo)
}

func Test401Then400(t *testing.T) {
	expiredToken := "test-expired-token"
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		if isRetry {
			return token, nil
		}
		return expiredToken, nil
	}
	ctx := context.Background()
	pathParams := map[string]string{}
	queryParams := map[string][]string{}
	var out Payload
	client, server := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Header.Get("authorization") == bearer(expiredToken) {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(400)
	})
	err := client.Get(ctx, server.URL, pathParams, queryParams, &out)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Equal(t, "", out.Foo)
}

// token error
// no reqID
// 401 error - token success - 200 success
// 401 error - token error
// 401 error - token success - 401 error (not infinite loop)
// 401 error - token success - 400 error
