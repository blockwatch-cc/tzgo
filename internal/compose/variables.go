// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	VAR_PREFIX  = "$"
	FILE_PREFIX = "@"
	TIME_PREFIX = "$now"
)

var NESTED_VAR = regexp.MustCompile(`\$[a-zA-Z0-9_-]+`)

func (c Context) ResolveNestedVars(src string) (string, bool) {
	if NESTED_VAR.MatchString(src) {
		return NESTED_VAR.ReplaceAllStringFunc(src, func(v string) string {
			if val, ok := c.Variables[v]; ok {
				return val
			}
			return v
		}), true
	}
	return src, false
}

func CreateVariable(v string) string {
	if IsVariable(v) {
		return v
	}
	return VAR_PREFIX + v
}

func IsVariable(v string) bool {
	return strings.HasPrefix(v, VAR_PREFIX)
}

func IsTimeExpression(v string) bool {
	return strings.HasPrefix(v, TIME_PREFIX)
}

func IsFile(v string) bool {
	return strings.HasPrefix(v, FILE_PREFIX)
}

func ConvertTime(v string) (string, error) {
	if IsTimeExpression(v) {
		v = strings.TrimPrefix(v, "$now")
		t := time.Now().UTC()
		if len(v) == 0 {
			return t.Format(time.RFC3339), nil
		}
		switch v[0] {
		case '-':
			d, err := time.ParseDuration(v[1:])
			if err != nil {
				return "", err
			}
			return t.Add(-d).Format(time.RFC3339), nil
		case '+':
			d, err := time.ParseDuration(v[1:])
			if err != nil {
				return "", err
			}
			return t.Add(d).Format(time.RFC3339), nil
		default:
			return "", fmt.Errorf("invalid time offset")
		}
	}
	t, err := ParseTime(v)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC3339), nil
}

func ParseTime(v string) (time.Time, error) {
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Unix(i, 0), nil
	}
	for _, f := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(f, v); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format")
}
