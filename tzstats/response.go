// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	headerRuntime  = "X-Runtime"
	trailerError   = "X-Streaming-Error"
	trailerCursor  = "X-Streaming-Cursor"
	trailerCount   = "X-Streaming-Count"
	trailerRuntime = "X-Streaming-Runtime"
)

type StreamResponse struct {
	Runtime time.Duration
	Cursor  string
	Count   int
}

func NewStreamResponse(header http.Header) (StreamResponse, error) {
	r := StreamResponse{}
	if header == nil {
		return r, nil
	}
	if cur, ok := header[trailerCursor]; ok && len(cur) > 0 {
		r.Cursor = cur[0]
	}
	if cnt, ok := header[trailerCount]; ok && len(cnt) > 0 {
		r.Count, _ = strconv.Atoi(cnt[0])
	}
	if rt, ok := header[trailerRuntime]; ok && len(rt) > 0 {
		d, _ := strconv.ParseInt(rt[0], 10, 64)
		r.Runtime = time.Duration(d) * time.Millisecond
	}
	if errStr, ok := header[trailerError]; ok && len(errStr) > 0 {
		e := &ApiErrors{}
		if err := e.UnmarshalJSON([]byte(errStr[0])); err != nil {
			return r, err
		}
		return r, e
	}
	return r, nil
}

// request holds information about a request that is used to properly
// detect, interpret, and deliver a reply to it.
type request struct {
	id              uint64
	httpRequest     *http.Request
	responseVal     interface{}
	responseHeaders http.Header
	responseChan    chan *response
}

func (r *request) String() string {
	return strings.Join([]string{
		r.httpRequest.Method,
		r.httpRequest.Proto,
		r.httpRequest.URL.String(),
	}, " ")
}

// response is the raw bytes of a JSON result, or the error if the
// HTTP call failed
type response struct {
	status  int
	headers http.Header
	result  []byte
	request string
	err     error
}

type FutureResult chan *response

func (r FutureResult) Receive(ctx context.Context) error {
	resp, err := receiveFuture(ctx, r)
	if err != nil {
		if herr, ok := IsHttpError(err); ok {
			return herr
		} else if rerr, ok := IsErrRateLimited(err); ok {
			return rerr
		} else {
			buf := resp.result
			if resp != nil && buf != nil && resp.status > 299 {
				errs := ApiErrors{}
				if err := json.Unmarshal(buf, &errs); err != nil {
					buf = bytes.Replace(bytes.TrimRight(resp.result[:min(len(buf), 512)], "\x00"), []byte{'\n'}, []byte{}, -1)
					errs.Errors = append(errs.Errors, ApiError{
						Status:  resp.status,
						Message: string(buf),
						Detail:  err.Error(),
					})
				}
				return errs
			}
		}
		return err
	}
	return nil
}

func (r FutureResult) Done() bool {
	return len(r) > 0
}

// newFutureError returns a new future result channel that already has the
// passed error waitin on the channel with the reply set to nil.  This is useful
// to easily return errors from the various Async functions.
func newFutureError(err error) chan *response {
	responseChan := make(chan *response, 1)
	responseChan <- &response{err: err}
	return responseChan
}

// receiveFuture receives from the passed futureResult channel to extract a
// reply or any errors.  The examined errors include an error in the
// futureResult and the error in the reply from the server.  This will block
// until the result is available on the passed channel.
func receiveFuture(ctx context.Context, f chan *response) (*response, error) {
	// Wait for a response on the returned channel.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-f:
		return r, r.err
	}
}

func mergeHeaders(merged, header, trailer http.Header) http.Header {
	if merged == nil {
		return merged
	}
	for n, _ := range merged {
		merged.Del(n)
	}
	for n, v := range header {
		for _, vv := range v {
			merged.Add(n, vv)
		}
	}
	for n, v := range trailer {
		for _, vv := range v {
			merged.Add(n, vv)
		}
	}
	return merged
}

var textResponseTypes = []string{
	"application/json",
	"text/csv",
	"text/html",
}

func isTextResponse(resp *http.Response) bool {
	h := resp.Header.Get("Content-Type")
	for _, v := range textResponseTypes {
		if strings.HasPrefix(h, v) {
			return true
		}
	}
	return false
}
