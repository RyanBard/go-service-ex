package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RyanBard/go-service-ex/internal/apiclient"
	"github.com/RyanBard/go-service-ex/internal/httpx"
	"github.com/stretchr/testify/assert"
)

func initClient(getToken func(isRetry bool) (string, error), f func(w http.ResponseWriter, r *http.Request)) (*userClient, *httptest.Server) {
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
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	expectedUser := User{
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
		assert.Equal(t, "/api/users/"+id, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"id":"test-user-id","name":"foo"}`))
	})
	u, err := client.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, u)
}

func TestGetByID_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	u, err := client.GetByID(ctx, id)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", u.ID)
}

func TestGetByID_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	u, err := client.GetByID(ctx, id)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, "", u.ID)
}

func TestGetAll(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	expectedUser := User{
		ID:   "test-user-id",
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
		assert.Equal(t, "/api/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"test-user-id","name":"foo"}]`))
	})
	u, err := client.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []User{expectedUser}, u)
}

func TestGetAll_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	u, err := client.GetAll(ctx)
	assert.Equal(t, mockTokenErr, err)
	assert.Len(t, u, 0)
}

func TestGetAll_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	u, err := client.GetAll(ctx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Len(t, u, 0)
}

func TestGetAllByOrgID(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	orgID := "test-org-id"
	expectedUser := User{
		ID:    "test-user-id",
		OrgID: orgID,
		Name:  "foo",
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
		assert.Equal(t, "/api/orgs/test-org-id/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"test-user-id","org_id":"test-org-id","name":"foo"}]`))
	})
	u, err := client.GetAllByOrgID(ctx, orgID)
	assert.Nil(t, err)
	assert.Equal(t, []User{expectedUser}, u)
}

func TestGetAllByOrgID_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	orgID := "test-org-id"
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	u, err := client.GetAllByOrgID(ctx, orgID)
	assert.Equal(t, mockTokenErr, err)
	assert.Len(t, u, 0)
}

func TestGetAllByOrgID_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	orgID := "test-org-id"
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	u, err := client.GetAllByOrgID(ctx, orgID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Len(t, u, 0)
}

func TestCreate(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	input := User{
		OrgID: "test-org-id",
		Name:  "foo",
	}
	expectedUser := User{
		ID:    "test-user-id",
		OrgID: "test-org-id",
		Name:  "foo",
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
		assert.Equal(t, []byte(`{"org_id":"test-org-id","name":"foo","is_system":false,"is_admin":false,"is_active":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","version":0}`), b)
		assert.Equal(t, "/api/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Write([]byte(`{"id":"test-user-id","org_id":"test-org-id","name":"foo"}`))
	})
	u, err := client.Save(ctx, input)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, u)
}

func TestCreate_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	input := User{
		OrgID: "test-org-id",
		Name:  "foo",
	}
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	u, err := client.Save(ctx, input)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", u.Name)
}

func TestCreate_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	input := User{
		OrgID: "test-org-id",
		Name:  "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	u, err := client.Save(ctx, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, "", u.Name)
}

func TestUpdate(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	input := User{
		ID:    id,
		OrgID: "test-org-id",
		Name:  "foo",
	}
	expectedUser := input
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		orgJSONStr := `{"id":"test-user-id","org_id":"test-org-id","name":"foo","is_system":false,"is_admin":false,"is_active":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","version":0}`
		assert.Equal(t, []byte(orgJSONStr), b)
		assert.Equal(t, "/api/users/"+id, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, bearer(token), r.Header.Get("authorization"))
		assert.Equal(t, reqID, r.Header.Get("x-request-id"))
		w.Write([]byte(orgJSONStr))
	})
	u, err := client.Save(ctx, input)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, u)
}

func TestUpdate_TokenErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	input := User{
		ID:    id,
		OrgID: "test-org-id",
		Name:  "foo",
	}
	mockTokenErr := errors.New("test-token-err")
	getToken := func(isRetry bool) (string, error) {
		return "", mockTokenErr
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {})
	u, err := client.Save(ctx, input)
	assert.Equal(t, mockTokenErr, err)
	assert.Equal(t, "", u.Name)
}

func TestUpdate_HTTPErr(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	input := User{
		ID:    id,
		OrgID: "test-org-id",
		Name:  "foo",
	}
	token := "test-token"
	getToken := func(isRetry bool) (string, error) {
		return token, nil
	}
	client, _ := initClient(getToken, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	u, err := client.Save(ctx, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, "", u.Name)
}

func TestDelete(t *testing.T) {
	reqID := "test-req-id"
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	input := DeleteUser{
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
		assert.Equal(t, []byte(`{"id":"test-user-id","version":1}`), b)
		assert.Equal(t, "/api/users/"+id, r.URL.Path)
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
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	input := DeleteUser{
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
	ctx := context.WithValue(context.Background(), "reqID", reqID)
	id := "test-user-id"
	input := DeleteUser{
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
