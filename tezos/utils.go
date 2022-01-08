// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
    "encoding/hex"
)

// HexBytes represents bytes as a JSON string of hexadecimal digits
type HexBytes []byte

// UnmarshalText umarshalls a hex string to bytes
func (h *HexBytes) UnmarshalText(data []byte) error {
    dst := make([]byte, hex.DecodedLen(len(data)))
    if _, err := hex.Decode(dst, data); err != nil {
        return err
    }
    *h = dst
    return nil
}

func (h HexBytes) MarshalText() ([]byte, error) {
    return []byte(hex.EncodeToString(h)), nil
}

func (h HexBytes) String() string {
    return hex.EncodeToString(h)
}

func (h HexBytes) Bytes() []byte {
    return []byte(h)
}
