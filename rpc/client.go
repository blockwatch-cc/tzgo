// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

const (
	libraryVersion = "1.17.0"
	userAgent      = "tzgo/v" + libraryVersion
	mediaType      = "application/json"
	ipfsUrl        = "https://ipfs.io"
)

// Client manages communication with a Tezos RPC server.
type Client struct {
	// HTTP client used to communicate with the Tezos node API.
	client *http.Client
	// Base URL for API requests.
	BaseURL *url.URL
	// Base URL for IPFS requests.
	IpfsURL *url.URL
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
	// MetadataMode defines the metadata reconstruction mode used for fetching
	// block and operation receipts. Set this mode to `always` if an RPC node prunes
	// metadata (i.e. you see metadata too large in certain operations)
	MetadataMode MetadataMode
	// Close connections. This may help with EOF errors from unexpected
	// connection close by Tezos RPC.
	CloseConns bool
	// Log is the logger implementation used by this client
	Log log.Logger
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
	q := u.Query()
	key := q.Get("api_key")
	if key != "" {
		q.Del("api_key")
		u.RawQuery = q.Encode()
	} else {
		key = os.Getenv("TZGO_API_KEY")
	}
	ipfs, _ := url.Parse(ipfsUrl)
	c := &Client{
		client:          httpClient,
		BaseURL:         u,
		IpfsURL:         ipfs,
		UserAgent:       userAgent,
		ApiKey:          key,
		BlockObserver:   NewObserver(),
		MempoolObserver: NewObserver(),
		MetadataMode:    MetadataModeAlways,
		Log:             logger,
	}
	return c, nil
}

func (c *Client) Init(ctx context.Context) error {
	return c.ResolveChainConfig(ctx)
}

func (c *Client) UseIpfsUrl(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	c.IpfsURL = u
	return nil
}

func (c *Client) Client() *http.Client {
	return c.client
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
	id, err := c.GetChainId(ctx)
	if err != nil {
		return err
	}
	c.ChainId = id
	p, err := c.GetParams(ctx, Head)
	if err != nil {
		return err
	}
	c.Params = p
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
	req.Close = c.CloseConns

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.UserAgent)
	if c.ApiKey != "" {
		req.Header.Add("X-Api-Key", c.ApiKey)
	}

	c.logDebugOnly(func() {
		c.Log.Debugf("%s %s %s", req.Method, req.URL, req.Proto)
	})
	c.logTraceOnly(func() {
		d, _ := httputil.DumpRequest(req, true)
		c.Log.Trace(string(d))
	})

	return req, nil
}

func (c *Client) handleResponse(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) handleResponseMonitor(ctx context.Context, resp *http.Response, mon Monitor) {
	// decode stream
	dec := json.NewDecoder(resp.Body)

	// close body when stream stopped
	defer func() {
		io.Copy(io.Discard, resp.Body)
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
			mon.Err(fmt.Errorf("rpc: %v", err))
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
		if e, ok := err.(*url.Error); ok {
			return e.Err
		}
		return err
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	c.logTraceOnly((func() {
		d, _ := httputil.DumpResponse(resp, true)
		c.Log.Trace(string(d))
	}))

	statusClass := resp.StatusCode / 100
	if statusClass == 2 {
		if v == nil {
			return nil
		}
		return c.handleResponse(resp, v)
	}

	return c.handleError(resp)
}

// DoAsync retrieves values from the API and sends responses using the provided monitor.
func (c *Client) DoAsync(req *http.Request, mon Monitor) error {
	//nolint:bodyclose
	resp, err := c.client.Do(req)
	if err != nil {
		if e, ok := err.(*url.Error); ok {
			return e.Err
		}
		return err
	}

	if resp.StatusCode == http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
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
		return c.handleError(resp)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
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
		c.Log.Errorf("rpc: error decoding RPC error response: %v", err)
		return &httpErr
	}

	return &rpcError{
		httpError: &httpErr,
		errors:    errs,
	}
}
