// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ApiError struct {
	Code      int    `json:"code"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Scope     string `json:"scope"`
	Detail    string `json:"detail"`
	RequestId string `json:"requestId"`
	Reason    string `json:"reason"`
}

func (e ApiError) Error() string {
	s := make([]string, 0)
	if e.Status != 0 {
		s = append(s, "status="+strconv.Itoa(e.Status))
	}
	if e.Code != 0 {
		s = append(s, "code="+strconv.Itoa(e.Code))
	}
	if e.Scope != "" {
		s = append(s, "scope="+e.Scope)
	}
	s = append(s, "message=\""+e.Message+"\"")
	if e.Detail != "" {
		s = append(s, "detail=\""+e.Detail+"\"")
	}
	if e.RequestId != "" {
		s = append(s, "request-id="+e.RequestId)
	}
	if e.Reason != "" {
		s = append(s, "reason=\""+e.Reason+"\"")
	}
	return strings.Join(s, " ")
}

type ApiErrors struct {
	Errors []ApiError `json:"errors"`
}

func (e *ApiErrors) UnmarshalJSON(buf []byte) error {
	if len(buf) < 2 {
		return nil
	}
	var t map[string]json.RawMessage
	if err := json.Unmarshal(buf, &t); err != nil {
		return err
	}
	// check if we have an embedded array and decode
	if v, ok := t["errors"]; ok {
		e.Errors = make([]ApiError, 0)
		return json.Unmarshal(v, &e.Errors)
	}
	// if not, decode as single error
	var v ApiError
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	e.Errors = []ApiError{v}
	return nil
}

func (e ApiErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	return e.Errors[0].Error()
}

func IsApiError(err error) (ApiErrors, bool) {
	e, ok := err.(ApiErrors)
	return e, ok
}

type HttpError struct {
	Status  int
	Data    string
	Request string
	Header  http.Header
}

func newHttpError(resp *http.Response, buf []byte, req string) error {
	if len(buf) > 0 && (buf[0] == '[' || buf[0] == '{') {
		var val interface{}
		json.Unmarshal(buf, &val)
		buf, _ = json.Marshal(val)
	} else {
		buf = bytes.Replace(bytes.TrimRight(buf[:min(len(buf), 512)], "\x00"), []byte{'\n'}, []byte{}, -1)
	}
	return HttpError{
		Status:  resp.StatusCode,
		Data:    string(buf),
		Request: req,
		Header:  resp.Header,
	}
}

func (e HttpError) Error() string {
	return fmt.Sprintf("%d %s: %s %s", e.Status, http.StatusText(e.Status), e.Data, e.Request)
}

func IsHttpError(err error) (HttpError, bool) {
	e, ok := err.(HttpError)
	return e, ok
}

type ErrRateLimited struct {
	Status          int
	IsResponseError bool
	deadline        time.Time
	done            chan struct{}
	Header          http.Header
}

func newRateLimitError(d time.Duration, resp *http.Response) ErrRateLimited {
	e := ErrRateLimited{
		Status:          resp.StatusCode,
		Header:          mergeHeaders(make(http.Header), resp.Header, resp.Trailer),
		IsResponseError: true,
		deadline:        time.Now().UTC().Add(d),
		done:            make(chan struct{}),
	}
	go e.timeout(d)
	return e
}

func (e ErrRateLimited) timeout(d time.Duration) {
	select {
	case <-time.After(d):
		close(e.done)
	}
}

func NewErrRateLimited(d time.Duration, isResponse bool) ErrRateLimited {
	e := ErrRateLimited{
		Status:          429,
		IsResponseError: isResponse,
		deadline:        time.Now().UTC().Add(d),
		done:            make(chan struct{}),
	}
	go e.timeout(d)
	return e
}

func (e ErrRateLimited) Error() string {
	return fmt.Sprintf("rate limited for %s", time.Until(e.deadline))
}

func (e ErrRateLimited) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.done:
		return nil
	}
}

func (e ErrRateLimited) Done() <-chan struct{} {
	return e.done
}

func (e ErrRateLimited) Deadline() time.Duration {
	return e.deadline.Sub(time.Now().UTC())
}

func IsErrRateLimited(err error) (ErrRateLimited, bool) {
	e, ok := err.(ErrRateLimited)
	return e, ok
}

func ErrorStatus(err error) int {
	switch e := err.(type) {
	case ErrRateLimited:
		return 427
	case HttpError:
		return e.Status
	case ApiError:
		return e.Status
	case ApiErrors:
		if len(e.Errors) > 0 {
			return e.Errors[0].Status
		}
		return 0
	default:
		return 0
	}
}
