// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var (
	ClientVersion = "1.0.0"
	userAgent     = "tzgo/v" + ClientVersion
	DefaultLimit  = 50000
)

type Client struct {
	httpClient *http.Client
	params     Params
	UserAgent  string
}

func NewClient(httpClient *http.Client, url string) (*Client, error) {
	params, err := ParseParams(url)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		params:     params,
		UserAgent:  userAgent,
	}, nil
}

func (c *Client) get(ctx context.Context, path string, headers http.Header, result interface{}) error {
	return c.call(ctx, http.MethodGet, path, headers, nil, result)
}

func (c *Client) post(ctx context.Context, path string, headers http.Header, data, result interface{}) error {
	return c.call(ctx, http.MethodPost, path, headers, data, result)
}

func (c *Client) put(ctx context.Context, path string, headers http.Header, data, result interface{}) error {
	return c.call(ctx, http.MethodPut, path, headers, data, result)
}

func (c *Client) delete(ctx context.Context, path string, headers http.Header) error {
	return c.call(ctx, http.MethodDelete, path, headers, nil, nil)
}

func (c *Client) getAsync(ctx context.Context, path string, headers http.Header, result interface{}) FutureResult {
	return c.callAsync(ctx, http.MethodGet, path, headers, nil, result)
}

func (c *Client) call(ctx context.Context, method, path string, headers http.Header, data, result interface{}) error {
	return c.callAsync(ctx, method, path, headers, data, result).Receive(ctx)
}

func (c *Client) callAsync(ctx context.Context, method, path string, headers http.Header, data, result interface{}) FutureResult {
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("User-Agent", c.UserAgent)
	if !strings.HasPrefix(path, "http") {
		path = c.params.Url(path)
	}

	req, err := c.newRequest(ctx, method, path, headers, data, result)
	if err != nil {
		return newFutureError(err)
	}

	responseChan := make(chan *response, 1)
	c.handleRequest(&request{
		httpRequest:     req,
		responseVal:     result,
		responseHeaders: headers,
		responseChan:    responseChan,
	})

	return responseChan
}

func (c *Client) newRequest(ctx context.Context, method, path string, headers http.Header, data, result interface{}) (*http.Request, error) {
	// prepare headers
	if headers == nil {
		headers = make(http.Header)
	}

	// prepare POST/PUT/PATCH payload
	var body io.Reader
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(b)
		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", "application/json")
		}
	}

	if result != nil && headers.Get("Accept") == "" {
		headers.Set("Accept", "application/json")
	}

	// create http request
	log.Debugf("%s %s", method, path)
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)

	// add content-type header to POST, PUT, PATCH
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		headers.Del("Content-Type")
	}

	// add all passed in headers
	for n, v := range headers {
		for _, vv := range v {
			req.Header.Add(n, vv)
		}
	}

	return req, nil
}

// handleRequest executes the passed HTTP request, reading the
// result, unmarshalling it, and delivering the unmarshalled result to the
// provided response channel.
func (c *Client) handleRequest(req *request) {
	// only dump content-type application/json
	log.Tracef("%v", newLogClosure(func() string {
		r, _ := httputil.DumpRequestOut(req.httpRequest, req.httpRequest.Header.Get("Content-Type") == "application/json")
		return string(r)
	}))

	resp, err := c.httpClient.Do(req.httpRequest)
	if err != nil {
		req.responseChan <- &response{err: err, request: req.String()}
		return
	}
	defer resp.Body.Close()

	log.Tracef("response: %v", newLogClosure(func() string {
		s, _ := httputil.DumpResponse(resp, isTextResponse(resp))
		return string(s)
	}))

	// process as stream when response interface is an io.Writer
	if resp.StatusCode == http.StatusOK && req.responseVal != nil {
		if stream, ok := req.responseVal.(io.Writer); ok {
			// log.Tracef("start streaming response")
			// forward stream
			_, err := io.Copy(stream, resp.Body)
			// close consumer if possible
			if closer, ok := req.responseVal.(io.WriteCloser); ok {
				// log.Tracef("closing stream after %d bytes", n)
				closer.Close()
			}
			// log.Tracef("response headers: %#v", resp.Header)
			// log.Tracef("response trailer: %#v", resp.Trailer)
			req.responseChan <- &response{
				status:  resp.StatusCode,
				request: req.String(),
				headers: mergeHeaders(req.responseHeaders, resp.Header, resp.Trailer),
				err:     err,
			}
			return
		}
	}

	// non-stream handling below

	// Read the raw bytes
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		req.responseChan <- &response{
			status:  resp.StatusCode,
			request: req.String(),
			headers: mergeHeaders(req.responseHeaders, resp.Header, resp.Trailer),
			err:     fmt.Errorf("reading reply: %v", err),
		}
		return
	}

	// on failure, return error and response (some API's send specific
	// error codes as details which we cannot parse here; some other APIs
	// even send 5xx error codes to signal non-error situations)
	if resp.StatusCode >= 400 {
		if resp.StatusCode == 429 {
			// TODO: read rate limit header
			wait := time.Second
			err = newRateLimitError(wait, resp)
		} else {
			err = newHttpError(resp, respBytes, req.String())
		}
		req.responseChan <- &response{
			status:  resp.StatusCode,
			request: req.String(),
			headers: mergeHeaders(req.responseHeaders, resp.Header, resp.Trailer),
			result:  respBytes,
			err:     err,
		}
		return
	}

	// unmarshal any JSON response
	isJson := strings.Contains(resp.Header.Get("Content-Type"), "application/json")

	// do this even if the response looks like JSON
	isJson = isJson || bytes.HasPrefix(respBytes, []byte("{")) || bytes.HasPrefix(respBytes, []byte("["))

	if isJson && req.responseVal != nil && (resp.ContentLength > 0 || resp.ContentLength == -1) {
		if err = json.Unmarshal(respBytes, req.responseVal); err == nil {
			req.responseChan <- &response{
				status:  resp.StatusCode,
				request: req.String(),
				headers: mergeHeaders(req.responseHeaders, resp.Header, resp.Trailer),
				err:     nil,
			}
			return
		}
	}
	req.responseChan <- &response{
		status:  resp.StatusCode,
		request: req.String(),
		headers: mergeHeaders(req.responseHeaders, resp.Header, resp.Trailer),
		result:  respBytes,
		err:     fmt.Errorf("unmarshalling reply: %v", err),
	}
}
