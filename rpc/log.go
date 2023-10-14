// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

//nolint:unused,deadcode
package rpc

import "github.com/echa/log"

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var logger log.Logger = log.Log

// The default amount of logging is none.
func init() {
	DisableLog()
}

// DisableLog disables all library log output.  Logging output is disabled
// by default until either UseLogger or SetLogWriter are called.
func DisableLog() {
	logger = log.Disabled
}

// UseLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using logpkg.
func UseLogger(l log.Logger) {
	logger = l
}

// LogClosure is a closure that can be printed with %v to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type logClosure func() string

// String invokes the log closure and returns the results string.
func (c logClosure) String() string {
	return c()
}

// newLogClosure returns a new closure over the passed function which allows
// it to be used as a parameter in a logging function that is only invoked when
// the logging level is such that the message will actually be logged.
func newLogClosure(c func() string) logClosure {
	return logClosure(c)
}

func (c *Client) logDebug(fn func()) {
	if c.Log.Level() <= log.LevelDebug {
		fn()
	}
}

func (c *Client) logDebugOnly(fn func()) {
	if c.Log.Level() == log.LevelDebug {
		fn()
	}
}

func (c *Client) logTrace(fn func()) {
	if c.Log.Level() <= log.LevelTrace {
		fn()
	}
}

func (c *Client) logTraceOnly(fn func()) {
	if c.Log.Level() == log.LevelTrace {
		fn()
	}
}
