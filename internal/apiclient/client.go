package apiclient

import (
	"context"
	"fmt"

	"github.com/RyanBard/go-service-ex/internal/ctxutil"
	"github.com/RyanBard/go-service-ex/internal/httpx"
)

type Client struct {
	hc       *httpx.Client
	getToken func(isRetry bool) (string, error)
}

func NewClient(hc *httpx.Client, getToken func(isRetry bool) (string, error)) *Client {
	return &Client{
		hc:       hc,
		getToken: getToken,
	}
}

func (ac *Client) Get(ctx context.Context, path string, pathParams map[string]string, queryParams map[string][]string, out interface{}) (err error) {
	return ac.common(ctx, "GET", path, pathParams, queryParams, nil, out)
}

func (ac *Client) Post(ctx context.Context, path string, pathParams map[string]string, queryParams map[string][]string, in interface{}, out interface{}) (err error) {
	return ac.common(ctx, "POST", path, pathParams, queryParams, in, out)
}

func (ac *Client) Put(ctx context.Context, path string, pathParams map[string]string, queryParams map[string][]string, in interface{}, out interface{}) (err error) {
	return ac.common(ctx, "PUT", path, pathParams, queryParams, in, out)
}

func (ac *Client) Delete(ctx context.Context, path string, pathParams map[string]string, queryParams map[string][]string, in interface{}, out interface{}) (err error) {
	return ac.common(ctx, "DELETE", path, pathParams, queryParams, in, out)
}

func (ac *Client) common(ctx context.Context, method string, path string, pathParams map[string]string, queryParams map[string][]string, in interface{}, out interface{}) (err error) {
	reqID, _ := ctx.Value(ctxutil.ContextKeyReqID{}).(string)

	var hb httpx.Builder
	switch method {
	case "GET":
		hb = ac.hc.Get(path, pathParams).
			WithQueryParams(queryParams).
			WithAccept("application/json")
	case "POST":
		hb = ac.hc.Post(path, pathParams).
			WithQueryParams(queryParams).
			WithAccept("application/json").
			WithContentType("application/json").
			WithBody(in)
	case "PUT":
		hb = ac.hc.Put(path, pathParams).
			WithQueryParams(queryParams).
			WithAccept("application/json").
			WithContentType("application/json").
			WithBody(in)
	case "DELETE":
		hb = ac.hc.Delete(path, pathParams).
			WithQueryParams(queryParams).
			WithAccept("application/json").
			WithContentType("application/json").
			WithBody(in)
	}

	token, err := ac.getToken(false)
	if err != nil {
		return err
	}
	headers := map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	if reqID != "" {
		headers["X-Request-Id"] = []string{reqID}
	}
	statusCode, err := hb.WithHeaders(headers).
		RetrieveWithContext(ctx, &out)
	if err != nil {
		if statusCode == 401 {
			token, tokenErr := ac.getToken(true)
			if tokenErr != nil {
				return err
			}
			headers["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}
			_, err = hb.WithHeaders(headers).
				RetrieveWithContext(ctx, &out)
		}
	}
	return err
}
