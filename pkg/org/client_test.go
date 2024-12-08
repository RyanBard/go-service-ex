package org

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RyanBard/go-service-ex/internal/apiclient"
	"github.com/RyanBard/go-service-ex/internal/ctxutil"
	"github.com/RyanBard/go-service-ex/internal/httpx"
	"github.com/stretchr/testify/assert"
)

func initClient(getToken func(isRetry bool) (string, error), f func(w http.ResponseWriter, r *http.Request)) (*orgClient, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(f))
	cfg := Config{
		BaseURL: server.URL,
	}
	client := NewClient(cfg, apiclient.NewClient(httpx.NewClient(http.Client{}), getToken))
	return client, server
}

func bearer(s string) string {
	return fmt.Sprintf("Bearer %s", s)
}

func TestGetByID(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	expectedOrg := Org{
		ID:   id,
		Name: "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte{}, b)
		assert.Equal(t, "/api/orgs/"+id, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"id":"test-org-id","name":"foo"}`))
	})
	o, err := client.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, o)
}

func TestGetByID_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	o, err := client.GetByID(ctx, id)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", o.ID)
}

func TestGetByID_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	o, err := client.GetByID(ctx, id)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, "", o.ID)
}

func TestGetAll(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	expectedOrg := Org{
		ID:   "test-org-id",
		Name: "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte{}, b)
		assert.Equal(t, "/api/orgs", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"test-org-id","name":"foo"}]`))
	})
	o, err := client.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []Org{expectedOrg}, o)
}

func TestGetAll_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	o, err := client.GetAll(ctx)
	assert.Equal(t, mockTokenErr, err)
	assert.Len(t, o, 0)
}

func TestGetAll_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	o, err := client.GetAll(ctx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Len(t, o, 0)
}

func TestSearchByName(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	name := "foo"
	expectedOrg := Org{
		ID:   "test-org-id",
		Name: name,
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte{}, b)
		assert.Equal(t, "/api/orgs", r.URL.Path)
		assert.Equal(t, "name=foo", r.URL.RawQuery)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"test-org-id","name":"foo"}]`))
	})
	o, err := client.SearchByName(ctx, name)
	assert.Nil(t, err)
	assert.Equal(t, []Org{expectedOrg}, o)
}

func TestSearchByName_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	name := "foo"
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	o, err := client.SearchByName(ctx, name)
	assert.Equal(t, mockTokenErr, err)
	assert.Len(t, o, 0)
}

func TestSearchByName_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	name := "foo"
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	o, err := client.SearchByName(ctx, name)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Len(t, o, 0)
}

func TestCreate(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	input := Org{
		Name: "foo",
	}
	expectedOrg := Org{
		ID:   "test-org-id",
		Name: "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(`{"name":"foo","is_system":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","version":0}`), b)
		assert.Equal(t, "/api/orgs", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Write([]byte(`{"id":"test-org-id","name":"foo"}`))
	})
	o, err := client.Save(ctx, input)
	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, o)
}

func TestCreate_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	input := Org{
		Name: "foo",
	}
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	o, err := client.Save(ctx, input)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", o.Name)
}

func TestCreate_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	input := Org{
		Name: "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	o, err := client.Save(ctx, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, "", o.Name)
}

func TestUpdate(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	input := Org{
		ID:   id,
		Name: "foo",
	}
	expectedOrg := input
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		orgJSONStr := `{"id":"test-org-id","name":"foo","is_system":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","version":0}`
		assert.Equal(t, []byte(orgJSONStr), b)
		assert.Equal(t, "/api/orgs/"+id, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Write([]byte(orgJSONStr))
	})
	o, err := client.Save(ctx, input)
	assert.Nil(t, err)
	assert.Equal(t, expectedOrg, o)
}

func TestUpdate_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	input := Org{
		ID:   id,
		Name: "foo",
	}
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	o, err := client.Save(ctx, input)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", o.Name)
}

func TestUpdate_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	input := Org{
		ID:   id,
		Name: "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	o, err := client.Save(ctx, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, "", o.Name)
}

func TestDelete(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	input := DeleteOrg{
		ID:      id,
		Version: 1,
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, []byte(`{"id":"test-org-id","version":1}`), b)
		assert.Equal(t, "/api/orgs/"+id, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.WriteHeader(204)
	})
	err := client.Delete(ctx, input)
	assert.Nil(t, err)
}

func TestDelete_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	input := DeleteOrg{
		ID:      id,
		Version: 1,
	}
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	err := client.Delete(ctx, input)
	assert.Equal(t, mockTokenErr, err)
}

func TestDelete_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, reqID)
	id := "test-org-id"
	input := DeleteOrg{
		ID:      id,
		Version: 1,
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	err := client.Delete(ctx, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
}
