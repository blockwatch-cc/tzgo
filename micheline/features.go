// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/json"
	"strings"
)

type Features uint16

const (
	FeatureAccountFactory Features = 1 << iota
	FeatureContractFactory
	FeatureSetDelegate
	FeatureLambda
	FeatureTransferTokens
	FeatureChainId
	FeatureTicket
	FeatureSapling
	FeatureView
	FeatureGlobalConstant
	FeatureTimelock
)

func (f Features) Contains(x Features) bool {
	return f&x > 0
}

func (f Features) String() string {
	return strings.Join(f.Array(), ",")
}

func (f Features) Array() []string {
	s := make([]string, 0)
	var i Features = 1
	for f > 0 {
		switch f & i {
		case FeatureAccountFactory:
			s = append(s, "account_factory")
		case FeatureContractFactory:
			s = append(s, "contract_factory")
		case FeatureSetDelegate:
			s = append(s, "set_delegate")
		case FeatureLambda:
			s = append(s, "lambda")
		case FeatureTransferTokens:
			s = append(s, "transfer_tokens")
		case FeatureChainId:
			s = append(s, "chain_id")
		case FeatureTicket:
			s = append(s, "ticket")
		case FeatureSapling:
			s = append(s, "sapling")
		case FeatureView:
			s = append(s, "view")
		case FeatureGlobalConstant:
			s = append(s, "global_constant")
		case FeatureTimelock:
			s = append(s, "timelock")
		}
		f &= ^i
		i <<= 1
	}
	return s
}

func (f Features) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Array())
}

func (s *Script) Features() Features {
	return s.Code.Param.Features() |
		s.Code.Storage.Features() |
		s.Code.Code.Features() |
		s.Code.View.Features()
}

func (p Prim) Features() Features {
	var f Features
	_ = p.Walk(func(p Prim) error {
		switch p.OpCode {
		case I_CREATE_ACCOUNT:
			f |= FeatureAccountFactory
		case I_CREATE_CONTRACT:
			f |= FeatureContractFactory
		case I_SET_DELEGATE:
			f |= FeatureSetDelegate
		case I_LAMBDA, I_EXEC, I_APPLY:
			f |= FeatureLambda
		case I_TRANSFER_TOKENS:
			f |= FeatureTransferTokens
		case I_CHAIN_ID:
			f |= FeatureChainId
		case I_TICKET, I_READ_TICKET, I_SPLIT_TICKET, I_JOIN_TICKETS:
			f |= FeatureTicket
		case I_SAPLING_VERIFY_UPDATE:
			f |= FeatureSapling
		case H_CONSTANT:
			f |= FeatureGlobalConstant
		case K_VIEW:
			f |= FeatureView
		case T_CHEST_KEY, T_CHEST, I_OPEN_CHEST:
			f |= FeatureTimelock
		}
		return nil
	})
	return f
}
