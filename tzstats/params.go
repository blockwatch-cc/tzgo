// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"fmt"
	"net/url"
	// "strconv"
	"strings"
)

type Params struct {
	Server string
	Prefix string
	Query  url.Values
}

func NewParams() Params {
	return Params{
		Query: url.Values{},
	}
}

func (p Params) Check() error {
	if p.Server == "" {
		return fmt.Errorf("empty server URL")
	}
	return nil
}

func (p Params) Copy() Params {
	np := NewParams()
	np.Server = p.Server
	np.Prefix = p.Prefix
	for n, v := range p.Query {
		np.Query[n] = v
	}
	return np
}

func (p Params) AppendQuery(path string) string {
	if len(p.Query) > 0 {
		return path + "?" + p.Query.Encode()
	}
	return path
}

func (p Params) Url(actions ...string) string {
	fields := make([]string, 0, 4)
	fields = append(fields, p.Server)
	if p.Prefix != "" {
		fields = append(fields, p.Prefix)
	}
	if len(actions) > 0 {
		for _, v := range actions {
			fields = append(fields, strings.TrimSuffix(strings.TrimPrefix(v, "/"), "/"))
		}
	}
	if len(p.Query) == 0 {
		return strings.Join(fields, "/")
	}
	return strings.Join([]string{
		strings.Join(fields, "/"),
		p.Query.Encode(),
	}, "?")
}

// parse from
// http://server:port/prefix
// server:port/prefix
// server/prefix
// /prefix
func ParseParams(urlString string) (Params, error) {
	p := NewParams()
	if !strings.HasPrefix(urlString, "http") {
		urlString = "https://" + urlString
	}
	u, err := url.Parse(urlString)
	if err != nil {
		return p, err
	}
	p.Query = u.Query()
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	if u.Path != "" {
		p.Prefix = u.Path
	}
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = ""
	u.RawPath = ""
	p.Server = u.String()
	return p, nil
}
