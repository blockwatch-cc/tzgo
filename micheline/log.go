// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	logpkg "github.com/echa/log"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log logpkg.Logger = logpkg.Log

// The default amount of logging is none.
func init() {
	DisableLog()
}

// DisableLog disables all library log output.  Logging output is disabled
// by default until either UseLogger or SetLogWriter are called.
func DisableLog() {
	log = logpkg.Disabled
}

// UseLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using logpkg.
func UseLogger(logger logpkg.Logger) {
	log = logger
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

// LogFn is a shot alias for a log function of type func(string, interface...)
type LogFn logpkg.LogfFn

// trace is a private trace logging function
var trace LogFn = nil

// UseTrace sets fn to be used as trace function
func UseTrace(fn LogFn) {
	trace = fn
}

// Trace is a function closure wrapper that forwards trace calls to an
// output function if set. Call UseTrace() to set a function of type LogFn
func Trace(fn func(log LogFn)) {
	if trace != nil {
		fn(trace)
	}
}
