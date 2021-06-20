// Copyright (c) 2021 Blockwatch Data Inc.
// Author:
//

package rpc

import (
	"context"
	"testing"

	"blockwatch.cc/tzgo/tezos"
)

func TestGetBlock(t *testing.T) {
	c, _ := NewClient("https://mainnet-tezos.giganode.io", nil)
	ctx := context.TODO()

	// GetBlock with BlockHash as parameter
	resBlock, err := c.GetBlock(ctx, tezos.MustParseBlockHash("BLxneVqo5855NK2gqdsGmMeAPRtgXxyFgncesY5sityY1dUbDcM"))
	if err != nil {
		t.Errorf("getblock error: %v", err)
	}
	if resBlock.Hash.String() != "BLxneVqo5855NK2gqdsGmMeAPRtgXxyFgncesY5sityY1dUbDcM" {
		t.Error("GetBlock error. See log for details")
		t.FailNow()
	}

	// GetBlock with block level (int64) as parameter
	var blockLevel int64 = 121212
	resBlock, err = c.GetBlock(ctx, blockLevel)
	if err != nil {
		t.Errorf("getblock error: %v", err)
	}
	if resBlock.Hash.String() != "BLxneVqo5855NK2gqdsGmMeAPRtgXxyFgncesY5sityY1dUbDcM" {
		t.Error("GetBlock error. See log for details")
		t.FailNow()
	}

	// GetBlock with block tag "genesis"
	resBlock, err = c.GetBlock(ctx, "genesis")
	if err != nil {
		t.Errorf("getblock error: %v", err)
	}
	if resBlock.Hash.String() != "BLockGenesisGenesisGenesisGenesisGenesisf79b5d1CoW2" {
		t.Error("GetBlock error. See log for details")
		t.FailNow()
	}

}
