package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	httpClient http.Client
}

type Builder struct {
	httpClient  http.Client
	method      string
	uri         string
	pathParams  map[string]string
	queryParams map[string][]string
	accept      string
	contentType string
	body        interface{}
	headers     map[string][]string
}

func NewClient(httpClient http.Client) *Client {
	return &Client{
		httpClient: httpClient,
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
	// TODO - think up something more generic than this
	errMessage := errorBody["message"]
	if errMessage == "" {
		errMessage = errorBody["error"]
	}
	return HTTPError{
		StatusCode: statusCode,
		ErrMessage: errMessage,
	}
}

func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

func (hc *Client) Head(uri string, pathParams map[string]string) Builder {
	return hc.newBuilder("HEAD", uri, pathParams)
}

func (hc *Client) Get(uri string, pathParams map[string]string) Builder {
	return hc.newBuilder("GET", uri, pathParams)
}

func (hc *Client) Post(uri string, pathParams map[string]string) Builder {
	return hc.newBuilder("POST", uri, pathParams)
}

func (hc *Client) Put(uri string, pathParams map[string]string) Builder {
	return hc.newBuilder("PUT", uri, pathParams)
}

func (hc *Client) Delete(uri string, pathParams map[string]string) Builder {
	return hc.newBuilder("DELETE", uri, pathParams)
}

func (hc *Client) newBuilder(method string, uri string, pathParams map[string]string) Builder {
	return Builder{
		httpClient:  hc.httpClient,
		method:      method,
		uri:         uri,
		pathParams:  pathParams,
		queryParams: map[string][]string{},
		headers:     map[string][]string{},
	}
}

func (hb Builder) WithQueryParams(queryParams map[string][]string) Builder {
	return Builder{
		httpClient:  hb.httpClient,
		method:      hb.method,
		uri:         hb.uri,
		pathParams:  hb.pathParams,
		queryParams: queryParams,
		accept:      hb.accept,
		contentType: hb.contentType,
		body:        hb.body,
		headers:     hb.headers,
	}
}

func (hb Builder) WithHeaders(headers map[string][]string) Builder {
	return Builder{
		httpClient:  hb.httpClient,
		method:      hb.method,
		uri:         hb.uri,
		pathParams:  hb.pathParams,
		queryParams: hb.queryParams,
		accept:      hb.accept,
		contentType: hb.contentType,
		body:        hb.body,
		headers:     headers,
	}
}

func (hb Builder) WithAccept(accept string) Builder {
	return Builder{
		httpClient:  hb.httpClient,
		method:      hb.method,
		uri:         hb.uri,
		pathParams:  hb.pathParams,
		queryParams: hb.queryParams,
		accept:      accept,
		contentType: hb.contentType,
		body:        hb.body,
		headers:     hb.headers,
	}
}

func (hb Builder) WithContentType(contentType string) Builder {
	return Builder{
		httpClient:  hb.httpClient,
		method:      hb.method,
		uri:         hb.uri,
		pathParams:  hb.pathParams,
		queryParams: hb.queryParams,
		accept:      hb.accept,
		contentType: contentType,
		body:        hb.body,
		headers:     hb.headers,
	}
}

func (hb Builder) WithBody(body interface{}) Builder {
	return Builder{
		httpClient:  hb.httpClient,
		method:      hb.method,
		uri:         hb.uri,
		pathParams:  hb.pathParams,
		queryParams: hb.queryParams,
		accept:      hb.accept,
		contentType: hb.contentType,
		body:        body,
		headers:     hb.headers,
	}
}

func (hb Builder) RetrieveStr(out *string) (int, error) {
	return hb.RetrieveStrWithContext(context.Background(), out)
}

func (hb Builder) RetrieveStrWithContext(ctx context.Context, out *string) (statusCode int, err error) {
	statusCode, b, err := hb.common(ctx)
	if err != nil {
		return statusCode, err
	}
	if hb.method == "HEAD" || statusCode == 204 || out == nil {
		// noop
	} else {
		*out = string(b)
	}
	return statusCode, err
}

func (hb Builder) Retrieve(out interface{}) (int, error) {
	return hb.RetrieveWithContext(context.Background(), out)
}

func (hb Builder) RetrieveWithContext(ctx context.Context, out interface{}) (statusCode int, err error) {
	statusCode, b, err := hb.common(ctx)
	if err != nil {
		return statusCode, err
	}
	if hb.method == "HEAD" || statusCode == 204 || out == nil {
		// noop
	} else {
		err = json.Unmarshal(b, &out)
	}
	return statusCode, err
}

func (hb Builder) common(ctx context.Context) (statusCode int, b []byte, err error) {
	u, err := url.Parse(hb.uri)
	if err != nil {
		return statusCode, b, err
	}
	path, escapedPath, err := renderPath(u.Path, hb.pathParams)
	if err != nil {
		return statusCode, b, err
	}
	u.RawPath = escapedPath
	u.Path = path
	if len(hb.queryParams) > 0 {
		values := u.Query()
		for k, vals := range hb.queryParams {
			for _, v := range vals {
				values.Add(k, v)
			}
		}
		u.RawQuery = values.Encode()
	}
	var body io.Reader
	if hb.method != "GET" && hb.method != "HEAD" && hb.body != nil && hb.body != "" {
		var data []byte
		if hb.contentType == "application/json" {
			if data, err = json.Marshal(hb.body); err != nil {
				return statusCode, b, err
			}
		} else {
			if bodyStr, ok := hb.body.(string); ok {
				data = []byte(bodyStr)
			} else {
				return statusCode, b, errors.New(fmt.Sprintf("contentType was '%s' and input body was not a string: body=%v", hb.contentType, hb.body))
			}
		}
		body = bytes.NewBuffer(data)
	}
	req, err := http.NewRequestWithContext(ctx, hb.method, u.String(), body)
	headers := hb.headers
	if hb.contentType != "" {
		headers["Content-Type"] = []string{hb.contentType}
	}
	if hb.accept != "" {
		headers["Accept"] = []string{hb.accept}
	}
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}
	resp, err := hb.httpClient.Do(req)
	if err != nil {
		return statusCode, b, err
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	if !isSuccess(statusCode) {
		err = newHTTPError(statusCode, resp.Body)
		return statusCode, b, err
	}
	b, err = io.ReadAll(resp.Body)
	return statusCode, b, err
}

func renderPath(inputPath string, pathParams map[string]string) (path string, escapedPath string, err error) {
	pathParts := strings.Split(inputPath, "/")
	renderedPathParts := make([]string, len(pathParts))
	escapedRenderedPathParts := make([]string, len(pathParts))
	for i, p := range pathParts {
		var val string
		if strings.HasPrefix(p, ":") {
			val = pathParams[p[1:]]
			if val == "" {
				err = errors.New(fmt.Sprintf("path param reference of '%s' not found in the path params", p))
				return path, escapedPath, err
			}
		} else {
			val = p
		}
		renderedPathParts[i] = val
		escapedRenderedPathParts[i] = url.PathEscape(val)
	}
	path = strings.Join(renderedPathParts, "/")
	escapedPath = strings.Join(escapedRenderedPathParts, "/")
	return path, escapedPath, err
}
