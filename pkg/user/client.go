package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Config struct {
	BaseURL string
}

type userClient struct {
	cfg        Config
	httpClient http.Client
	getToken   func(isRetry bool) (string, error)
}

func NewClient(cfg Config, httpClient http.Client, getToken func(isRetry bool) (string, error)) *userClient {
	return &userClient{
		cfg:        cfg,
		httpClient: httpClient,
		getToken:   getToken,
	}
}

type HTTPError struct {
	StatusCode int
	ErrMessage string
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("http call failed with %d status: %s", h.StatusCode, h.ErrMessage)
}

func newHTTPError(statusCode int, body io.ReadCloser) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return HTTPError{
			StatusCode: statusCode,
		}
	}
	var errorBody map[string]string
	err = json.Unmarshal(b, &errorBody)
	if err != nil {
		return HTTPError{
			StatusCode: statusCode,
			ErrMessage: string(b),
		}
	}
	errMessage := errorBody["message"]
	return HTTPError{
		StatusCode: statusCode,
		ErrMessage: errMessage,
	}
}

func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func (uc *userClient) GetByID(ctx context.Context, id string) (u User, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/users/%s", uc.cfg.BaseURL, id), nil)
	// TODO - escape id properly
	if err != nil {
		return u, err
	}
	req.Header.Set("Accept", "application/json")
	token, err := uc.getToken(false)
	if err != nil {
		return u, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = uc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = uc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return u, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return u, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return u, err
	}
	err = json.Unmarshal(b, &u)
	return u, err
}

func (uc *userClient) GetAll(ctx context.Context) (u []User, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/users", uc.cfg.BaseURL), nil)
	if err != nil {
		return u, err
	}
	req.Header.Set("Accept", "application/json")
	token, err := uc.getToken(false)
	if err != nil {
		return u, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = uc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = uc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return u, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return u, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return u, err
	}
	err = json.Unmarshal(b, &u)
	return u, err
}

func (uc *userClient) GetAllByOrgID(ctx context.Context, orgID string) (u []User, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/orgs/%s/users", uc.cfg.BaseURL, orgID), nil)
	if err != nil {
		return u, err
	}
	req.Header.Set("Accept", "application/json")
	token, err := uc.getToken(false)
	if err != nil {
		return u, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = uc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = uc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return u, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return u, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return u, err
	}
	err = json.Unmarshal(b, &u)
	return u, err
}

func (uc *userClient) Save(ctx context.Context, input User) (u User, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	jsonData, err := json.Marshal(input)
	if err != nil {
		return u, err
	}
	var req *http.Request
	if input.ID == "" {
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/users", uc.cfg.BaseURL), bytes.NewBuffer(jsonData))
	} else {
		req, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/users/%s", uc.cfg.BaseURL, input.ID), bytes.NewBuffer(jsonData))
		// TODO - escape input.ID properly
	}
	if err != nil {
		return u, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	token, err := uc.getToken(false)
	if err != nil {
		return u, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = uc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = uc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return u, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return u, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return u, err
	}
	err = json.Unmarshal(b, &u)
	return u, err
}

func (uc *userClient) Delete(ctx context.Context, input DeleteUser) (err error) {
	reqID, _ := ctx.Value("reqID").(string)
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/users/%s", uc.cfg.BaseURL, input.ID), bytes.NewBuffer(jsonData))
	// TODO - escape input.ID properly
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	token, err := uc.getToken(false)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = uc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = uc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return err
	}
	return nil
}
