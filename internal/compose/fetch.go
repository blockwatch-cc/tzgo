// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/echa/log"
	"github.com/tidwall/gjson"
)

func Fetch[T any](ctx Context, url string) (*T, error) {
	cleanUrl, jsonPath, hasPath := strings.Cut(url, "#")
	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cleanUrl, nil)
	if err != nil {
		return nil, err
	}
	if ctx.apiKey != "" {
		req.Header.Add("X-Api-Key", ctx.apiKey)
	}
	if ctx.Log.Level() == log.LevelDebug {
		ctx.Log.Debugf("%s %s %s", req.Method, req.URL, req.Proto)
	}
	if ctx.Log.Level() == log.LevelTrace {
		d, _ := httputil.DumpRequest(req, true)
		ctx.Log.Trace(string(d))
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if ctx.Log.Level() == log.LevelTrace {
		d, _ := httputil.DumpResponse(res, true)
		ctx.Log.Trace(string(d))
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s ", res.Status)
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if hasPath {
		res := gjson.GetBytes(buf, jsonPath)
		if !res.Exists() {
			return nil, fmt.Errorf("missing path %q in result from %s", jsonPath, cleanUrl)
		}
		buf = []byte(res.Raw)
	}
	var k T
	if err := json.Unmarshal(buf, &k); err != nil {
		return nil, err
	}
	return &k, nil
}
