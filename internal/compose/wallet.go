// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"crypto/ed25519"

	"blockwatch.cc/tzgo/tezos"
	"github.com/tyler-smith/go-bip32"
)

func (c *Context) MakeAccount(id int, alias string) (Account, error) {
	if alias == "" {
		return Account{}, ErrNoAccountName
	}
	for _, acc := range c.Accounts {
		if acc.Id == id {
			c.AddVariable(alias, acc.Address.String())
			return acc, nil
		}
	}
	if id < 0 {
		id = c.MaxId + 1
	}
	sk, err := deriveChildKey(c.BaseAccount.PrivateKey, id)
	if err != nil {
		return Account{}, nil
	}
	acc := Account{
		Id:         id,
		Address:    sk.Address(),
		PrivateKey: sk,
	}
	c.Log.Debugf("Creating account %d %s %s", id, acc.Address, alias)
	c.AddVariable(alias, acc.Address.String())
	c.AddAccount(acc)
	c.MaxId = acc.Id
	return acc, nil
}

func deriveChildKey(seed tezos.PrivateKey, id int) (tezos.PrivateKey, error) {
	var sk tezos.PrivateKey
	masterKey, err := bip32.NewMasterKey(seed.Data)
	if err != nil {
		return sk, err
	}
	bip32ChildKey, err := masterKey.NewChildKey(uint32(id))
	if err != nil {
		return sk, err
	}
	sk = tezos.PrivateKey{
		Type: tezos.KeyTypeEd25519,
		Data: ed25519.NewKeyFromSeed(bip32ChildKey.Key),
	}
	return sk, nil
}
