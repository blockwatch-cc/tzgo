// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"fmt"
	"time"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

func ParseValue(typ micheline.OpCode, value string) (any, error) {
	switch typ {
	case micheline.T_STRING:
		return value, nil
	case micheline.T_ADDRESS:
		return tezos.ParseAddress(value)
	case micheline.T_NAT, micheline.T_MUTEZ, micheline.T_INT:
		return tezos.ParseZ(value)
	case micheline.T_TIMESTAMP:
		return time.Parse(time.RFC3339, value)
	case micheline.T_BYTES:
		var h tezos.HexBytes
		if err := h.UnmarshalText([]byte(value)); err != nil {
			return nil, err
		}
		return h.Bytes(), nil
	case micheline.T_KEY:
		return tezos.DecodeKey([]byte(value))
	case micheline.T_SIGNATURE:
		return tezos.ParseSignature(value)
	case micheline.T_CHAIN_ID:
		return tezos.ParseChainIdHash(value)
	default:
		return nil, fmt.Errorf("cannot parsed typ %q is ", typ)
	}
}
