// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"errors"
)

var (
	ErrNoVersion      = errors.New("missing engine version")
	ErrInvalidVersion = errors.New("unsupported engine version")
	ErrNoPipeline     = errors.New("missing pipeline definition")
	ErrNoBaseKey      = errors.New("missing base account key, set with TZCOMPOSE_BASE_KEY")
	ErrNoAccount      = errors.New("missing account")
	ErrNoAccountName  = errors.New("emoty account name")
	ErrNoPipelineName = errors.New("empty pipeline name")
	ErrNoTaskType     = errors.New("missing task type")
	ErrSkip           = errors.New("skip task")
)
