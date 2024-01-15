// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"encoding/json"
)

func Json(v any, indent ...string) string {
	var buf []byte
	switch len(indent) {
	case 0:
		buf, _ = json.Marshal(v)
	case 1:
		buf, _ = json.MarshalIndent(v, "", indent[0])
	default:
		buf, _ = json.MarshalIndent(v, indent[0], indent[1])
	}
	return string(buf)
}
