// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package main

import (
	"blockwatch.cc/tzgo/rpc"
	logpkg "github.com/echa/log"
)

var (
	log        = logpkg.NewLogger("MAIN")
	rpcLog     = logpkg.NewLogger("RPC ")
	taskLog    = logpkg.NewLogger("TASK")
	LevelTrace = logpkg.LevelTrace
)

// loggers maps each subsystem identifier to its associated logger.
var loggers = map[string]logpkg.Logger{
	"MAIN": log,
	"RPC":  rpcLog,
	"TASK": taskLog,
}

func initLogging() {
	// assign default loggers
	rpc.UseLogger(rpcLog)

	// handle cli flags
	var lvl logpkg.Level
	switch {
	case vtrace:
		lvl = logpkg.LevelTrace
	case vdebug:
		lvl = logpkg.LevelDebug
	case verbose:
		lvl = logpkg.LevelInfo
	default:
		lvl = logpkg.LevelWarn
	}
	setLogLevels(lvl)
}

// setLogLevel sets the logging level for provided subsystem.  Invalid
// subsystems are ignored.
func setLogLevel(id string, level logpkg.Level) {
	// Ignore invalid subsystems.
	logger, ok := loggers[id]
	if !ok {
		return
	}

	logger.SetLevel(level)
}

// setLogLevels sets the log level for all subsystem loggers to the passed
// level.
func setLogLevels(level logpkg.Level) {
	for id := range loggers {
		setLogLevel(id, level)
	}
}
