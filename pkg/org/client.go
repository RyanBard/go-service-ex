package org

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

type orgClient struct {
	cfg        Config
	httpClient http.Client
	getToken   func(isRetry bool) (string, error)
}

func NewClient(cfg Config, httpClient http.Client, getToken func(isRetry bool) (string, error)) *orgClient {
	return &orgClient{
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

func (oc *orgClient) GetByID(ctx context.Context, id string) (o Org, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/orgs/%s", oc.cfg.BaseURL, id), nil)
	// TODO - escape id properly
	if err != nil {
		return o, err
	}
	req.Header.Set("Accept", "application/json")
	token, err := oc.getToken(false)
	if err != nil {
		return o, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return o, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = oc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = oc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return o, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return o, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return o, err
	}
	err = json.Unmarshal(b, &o)
	return o, err
}

func (oc *orgClient) GetAll(ctx context.Context) (o []Org, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/orgs", oc.cfg.BaseURL), nil)
	if err != nil {
		return o, err
	}
	req.Header.Set("Accept", "application/json")
	token, err := oc.getToken(false)
	if err != nil {
		return o, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return o, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = oc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = oc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return o, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return o, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return o, err
	}
	err = json.Unmarshal(b, &o)
	return o, err
}

func (oc *orgClient) SearchByName(ctx context.Context, name string) (o []Org, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/orgs", oc.cfg.BaseURL), nil)
	if err != nil {
		return o, err
	}
	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")
	token, err := oc.getToken(false)
	if err != nil {
		return o, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return o, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = oc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = oc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return o, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return o, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return o, err
	}
	err = json.Unmarshal(b, &o)
	return o, err
}

func (oc *orgClient) Save(ctx context.Context, input Org) (o Org, err error) {
	reqID, _ := ctx.Value("reqID").(string)
	jsonData, err := json.Marshal(input)
	if err != nil {
		return o, err
	}
	var req *http.Request
	if input.ID == "" {
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/orgs", oc.cfg.BaseURL), bytes.NewBuffer(jsonData))
	} else {
		req, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/orgs/%s", oc.cfg.BaseURL, input.ID), bytes.NewBuffer(jsonData))
		// TODO - escape input.ID properly
	}
	if err != nil {
		return o, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	token, err := oc.getToken(false)
	if err != nil {
		return o, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return o, err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = oc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = oc.httpClient.Do(req)
		defer resp.Body.Close()
		statusCode = resp.StatusCode
	}
	if err != nil {
		return o, err
	}
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return o, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return o, err
	}
	err = json.Unmarshal(b, &o)
	return o, err
}

func (oc *orgClient) Delete(ctx context.Context, input DeleteOrg) (err error) {
	reqID, _ := ctx.Value("reqID").(string)
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/orgs/%s", oc.cfg.BaseURL, input.ID), bytes.NewBuffer(jsonData))
	// TODO - escape input.ID properly
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	token, err := oc.getToken(false)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode == 401 {
		token, err = oc.getToken(true)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp, err = oc.httpClient.Do(req)
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
