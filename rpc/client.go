// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
)

const (
	libraryVersion = "1.12-rc1"
	userAgent      = "tzgo/v" + libraryVersion
	mediaType      = "application/json"
)

// Client manages communication with a Tezos RPC server.
type Client struct {
	// HTTP client used to communicate with the Tezos node API.
	client *http.Client
	// Base URL for API requests.
	BaseURL *url.URL
	// User agent name for client.
	UserAgent string
	// Optional API key for protected endpoints
	ApiKey string
	// The chain the client will query.
	ChainId tezos.ChainIdHash
	// The current chain configuration.
	Params *tezos.Params
	// An active event observer to watch for operation inclusion
	BlockObserver *Observer
	// An active event observer to watch for operation posting to the mempool
	MempoolObserver *Observer
	// A default signer used for transaction sending
	Signer signer.Signer
}

// NewClient returns a new Tezos RPC client.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		client:          httpClient,
		BaseURL:         u,
		UserAgent:       userAgent,
		BlockObserver:   NewObserver(),
		MempoolObserver: NewObserver(),
	}
	return c, nil
}

func (c *Client) Init(ctx context.Context) error {
	return c.ResolveChainConfig(ctx)
}

func (c *Client) Listen() {
	// start observers
	c.BlockObserver.Listen(c)
	c.MempoolObserver.ListenMempool(c)
}

func (c *Client) Close() {
	c.BlockObserver.Close()
	c.MempoolObserver.Close()
}

func (c *Client) ResolveChainConfig(ctx context.Context) error {
	p, err := c.GetParams(ctx, Head)
	if err != nil {
		return err
	}
	c.Params = p
	c.ChainId = p.ChainId
	return nil
}

func (c *Client) Get(ctx context.Context, urlpath string, result interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodGet, urlpath, nil)
	if err != nil {
		return err
	}
	return c.Do(req, result)
}

func (c *Client) GetAsync(ctx context.Context, urlpath string, mon Monitor) error {
	req, err := c.NewRequest(ctx, http.MethodGet, urlpath, nil)
	if err != nil {
		return err
	}
	return c.DoAsync(req, mon)
}

func (c *Client) Put(ctx context.Context, urlpath string, body, result interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodPut, urlpath, body)
	if err != nil {
		return err
	}
	return c.Do(req, result)
}

func (c *Client) Post(ctx context.Context, urlpath string, body, result interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodPost, urlpath, body)
	if err != nil {
		return err
	}
	return c.Do(req, result)
}

// NewRequest creates a Tezos RPC request.
func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.UserAgent)
	if c.ApiKey != "" {
		req.Header.Add("X-Api-Key", c.ApiKey)
	}

	log.Debug(newLogClosure(func() string {
		d, _ := httputil.DumpRequest(req, true)
		return string(d)
	}))

	return req, nil
}

func (c *Client) handleResponse(ctx context.Context, resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) handleResponseMonitor(ctx context.Context, resp *http.Response, mon Monitor) {
	// decode stream
	dec := json.NewDecoder(resp.Body)

	// close body when stream stopped
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	for {
		chunkVal := mon.New()
		if err := dec.Decode(chunkVal); err != nil {
			select {
			case <-mon.Closed():
				return
			case <-ctx.Done():
				return
			default:
			}
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				mon.Err(io.EOF)
				return
			}
			mon.Err(fmt.Errorf("rpc: %w", err))
			return
		}
		select {
		case <-mon.Closed():
			return
		case <-ctx.Done():
			return
		default:
			mon.Send(ctx, chunkVal)
		}
	}
}

// Do retrieves values from the API and marshals them into the provided interface.
func (c *Client) Do(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	log.Trace(newLogClosure(func() string {
		d, _ := httputil.DumpResponse(resp, true)
		return string(d)
	}))

	statusClass := resp.StatusCode / 100
	if statusClass == 2 {
		if v == nil {
			return nil
		}
		return c.handleResponse(req.Context(), resp, v)
	}

	return handleError(resp)
}

// DoAsync retrieves values from the API and sends responses using the provided monitor.
func (c *Client) DoAsync(req *http.Request, mon Monitor) error {
	resp, err := c.client.Do(req)
	if err != nil {
		if e, ok := err.(*url.Error); ok {
			return e.Err
		}
		return err
	}

	if resp.StatusCode == http.StatusNoContent {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return nil
	}

	statusClass := resp.StatusCode / 100
	if statusClass == 2 {
		if mon != nil {
			go func() {
				c.handleResponseMonitor(req.Context(), resp, mon)
			}()
			return nil
		}
	} else {
		return handleError(resp)
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

func handleError(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	httpErr := httpError{
		request:    resp.Request.Method + " " + resp.Request.URL.RequestURI(),
		status:     resp.Status,
		statusCode: resp.StatusCode,
		body:       bytes.ReplaceAll(body, []byte("\n"), []byte{}),
	}

	if resp.StatusCode < 500 || !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		// Other errors with unknown body format (usually human readable string)
		return &httpErr
	}

	var errs Errors
	if err := json.Unmarshal(body, &errs); err != nil {
		return &plainError{&httpErr, fmt.Sprintf("rpc: error decoding RPC error: %v", err)}
	}

	if len(errs) == 0 {
		log.Errorf("rpc: error decoding RPC error response: %w", err)
		return &httpErr
	}

	return &rpcError{
		httpError: &httpErr,
		errors:    errs,
	}
}
