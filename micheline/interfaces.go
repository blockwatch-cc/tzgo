// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/json"
	"strings"
)

func (s *Script) Implements(i Interface) bool {
	eps, _ := s.Entrypoints(true)
	if len(eps) == 0 {
		return false
	}
	return i.Matches(eps)
}

func (s *Script) Interfaces() Interfaces {
	eps, _ := s.Entrypoints(true)
	if len(eps) == 0 {
		return nil
	}
	iv := make(Interfaces, 0)
	for _, i := range knownInterfaces {
		if !i.Matches(eps) {
			continue
		}
		iv = append(iv, i)
	}
	return iv
}

type Interface string

// search all interfaces in the list of entrypoints
func (m Interface) Matches(e Entrypoints) bool {
	for _, spec := range michelsonInterfaces[m] {
		var matched bool
		for _, ep := range e {
			if IsEqualPrim(spec, *ep.Prim, false) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

type Interfaces []Interface

func (i Interfaces) Contains(x Interface) bool {
	for _, v := range i {
		if v == x {
			return true
		}
	}
	return false
}

func (i Interfaces) String() string {
	if len(i) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range i {
		if k > 0 {
			b.WriteRune(',')
		}
		b.WriteString(string(v))
	}
	return b.String()
}

func (i *Interfaces) Parse(s string) error {
	if len(s) == 0 {
		return nil
	}
	split := strings.Split(s, ",")
	if cap(*i) < len(split) {
		*i = make([]Interface, len(split))
	}
	(*i) = (*i)[:len(split)]
	for k := range split {
		(*i)[k] = Interface(split[k])
	}
	return nil
}

func (i Interfaces) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Interfaces) UnmarshalText(b []byte) error {
	return i.Parse(string(b))
}

func (i Interfaces) MarshalJSON() ([]byte, error) {
	x := make([]string, len(i))
	for k, v := range i {
		x[k] = string(v)
	}
	return json.Marshal(x)
}

var (
	IManager     = Interface("MANAGER")
	ISetDelegate = Interface("SET_DELEGATE")
	ITzip5       = Interface("TZIP-005")
	ITzip7       = Interface("TZIP-007")
	ITzip12      = Interface("TZIP-012")
	IDexter      = Interface("DEXTER")
	// IKolibriVault = Interface("KOLIBRI_VAULT")
	// IWXTZVault    = Interface("WXTZ_VAULT")

	knownInterfaces = []Interface{
		IManager,
		ISetDelegate,
		ITzip5,
		ITzip7,
		ITzip12,
		// ITzip13,
		IDexter,
		// IKolibriVault,
		// IWXTZVault,
	}
)

// lists of required entrypoints (note: annotations are optional)
var michelsonInterfaces = map[Interface][]Prim{
	// manager.tz
	IManager: []Prim{
		// 1 (lambda %do unit (list operation))
		NewCodeAnno(T_LAMBDA, "%do", NewCode(T_UNIT), NewCode(T_LIST, NewCode(T_OPERATION))),
		// 2 (unit %default)
		NewCodeAnno(T_UNIT, "%default"),
	},
	// generic set delegate interface
	ISetDelegate: []Prim{
		// option %setDelegate key_hash
		NewCodeAnno(T_OPTION, "%setDelegate", NewCode(T_KEY_HASH)),
	},
	// Tzip 5 a.k.a FA1
	// https://gitlab.com/tzip/tzip/-/blob/master/proposals/tzip-5/tzip-5.md
	ITzip5: []Prim{
		// (address :from, (address :to, nat :value)) %transfer
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":from"),
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":to"),
				NewCodeAnno(T_NAT, ":value"),
			),
			"%transfer",
		),
		// view (address :owner) nat %getBalance
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":owner"),
			NewCode(T_CONTRACT, NewCode(T_NAT)),
			"%getBalance",
		),
		// view unit nat %getTotalSupply
		NewPairType(
			NewCode(T_UNIT),
			NewCode(T_CONTRACT, NewCode(T_NAT)),
			"%getTotalSupply",
		),
	},
	// Tzip 7 a.k.a FA1.2
	// https://gitlab.com/tzip/tzip/-/blob/master/proposals/tzip-7/tzip-7.md
	ITzip7: []Prim{
		// (address :from, (address :to, nat :value)) %transfer
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":from"),
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":to"),
				NewCodeAnno(T_NAT, ":value"),
			),
			"%transfer",
		),
		// (address :spender, nat :value) %approve
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":spender"),
			NewCodeAnno(T_NAT, ":value"),
			"%approve",
		),
		// (view (address :owner, address :spender) nat) %getAllowance
		NewPairType(
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":owner"),
				NewCodeAnno(T_ADDRESS, ":spender"),
			),
			NewCode(T_CONTRACT, NewCode(T_NAT)),
			"%getAllowance",
		),
		// (view (address :owner) nat) %getBalance
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":owner"),
			NewCode(T_CONTRACT, NewCode(T_NAT)),
			"%getBalance",
		),
		// (view unit nat) %getTotalSupply
		NewPairType(
			NewCode(T_UNIT),
			NewCode(T_CONTRACT, NewCode(T_NAT)),
			"%getTotalSupply",
		),
	},
	// Tzip 12 a.k.a. FA2
	// https://gitlab.com/tzip/tzip/-/blob/master/proposals/tzip-12/tzip-12.md
	ITzip12: []Prim{
		// (list %transfer
		//   (pair
		//     (address %from_)
		//     (list %txs
		//       (pair
		//         (address %to_)
		//         (pair
		//           (nat %token_id)
		//           (nat %amount)
		//         )
		//       )
		//     )
		//   )
		// )
		NewCodeAnno(T_LIST, "%transfer",
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":from_"),
				NewCodeAnno(T_LIST, ":txs",
					NewPairType(
						NewCodeAnno(T_ADDRESS, ":to_"),
						NewPairType(
							NewCodeAnno(T_NAT, ":token_id"),
							NewCodeAnno(T_NAT, ":amount"),
						),
					),
				),
			),
		),
		// (pair %balance_of
		//   (list %requests
		//     (pair
		//       (address %owner)
		//       (nat %token_id)
		//     )
		//   )
		//   (contract %callback
		//     (list
		//       (pair
		//         (pair %request
		//           (address %owner)
		//           (nat %token_id)
		//         )
		//         (nat %balance)
		//       )
		//     )
		//   )
		// )
		NewPairType(
			NewCodeAnno(T_LIST, ":requests",
				NewPairType(
					NewCodeAnno(T_ADDRESS, ":owner"),
					NewCodeAnno(T_NAT, ":token_id"),
				),
			),
			NewCodeAnno(T_CONTRACT, ":callback",
				NewCode(T_LIST,
					NewPairType(
						NewPairType(
							NewCodeAnno(T_ADDRESS, ":owner"),
							NewCodeAnno(T_NAT, ":token_id"),
							":request",
						),
						NewCodeAnno(T_NAT, ":balance"),
					),
				),
			),
			"%balance_of",
		),
		// (list %update_operators
		//   (or
		//     (pair %add_operator
		//       (address %owner)
		//       (pair
		//         (address %operator)
		//         (nat %token_id)
		//       )
		//     )
		//     (pair %remove_operator
		//       (address %owner)
		//       (pair
		//         (address %operator)
		//         (nat %token_id)
		//       )
		//     )
		//   )
		// )
		NewCodeAnno(T_LIST, "%update_operators",
			NewCode(T_OR,
				NewPairType(
					NewCodeAnno(T_ADDRESS, ":owner"),
					NewPairType(
						NewCodeAnno(T_ADDRESS, ":operator"),
						NewCodeAnno(T_NAT, ":token_id"),
					),
					":add_operator",
				),
				NewPairType(
					NewCodeAnno(T_ADDRESS, ":owner"),
					NewPairType(
						NewCodeAnno(T_ADDRESS, ":operator"),
						NewCodeAnno(T_NAT, ":token_id"),
					),
					":remove_operator",
				),
			),
		),
	},
	IDexter: []Prim{
		// 1 ( pair %approve
		//     ( address :spender )
		//     ( pair ( nat :allowance ) ( nat :currentAllowance ) ) )
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":spender"),
			NewPairType(
				NewCodeAnno(T_NAT, ":allowance"),
				NewCodeAnno(T_NAT, ":currentAllowance"),
			),
			"%approve",
		),

		// 2 ( pair %addLiquidity
		//     ( pair ( address :owner ) ( nat :minLqtMinted ) )
		//     ( pair ( nat :maxTokensDeposited ) ( timestamp :deadline ) ) ) )
		NewPairType(
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":owner"),
				NewCodeAnno(T_NAT, ":minLqtMinted"),
			),
			NewPairType(
				NewCodeAnno(T_NAT, ":maxTokensDeposited"),
				NewCodeAnno(T_TIMESTAMP, ":deadline"),
			),
			"%addLiquidity",
		),

		// 3 ( pair %removeLiquidity
		//     ( pair ( address :owner ) ( pair ( address :to ) ( nat :lqtBurned ) ) )
		//     ( pair ( mutez :minXtzWithdrawn ) ( pair ( nat :minTokensWithdrawn ) ( timestamp :deadline ) ) ) )
		NewPairType(
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":owner"),
				NewPairType(
					NewCodeAnno(T_ADDRESS, ":to"),
					NewCodeAnno(T_NAT, ":lqtBurned"),
				),
			),
			NewPairType(
				NewCodeAnno(T_MUTEZ, ":minXtzWithdrawn"),
				NewPairType(
					NewCodeAnno(T_NAT, ":minTokensWithdrawn"),
					NewCodeAnno(T_TIMESTAMP, ":deadline"),
				),
			),
			"%removeLiquidity",
		),

		// 4 ( pair %xtzToToken
		//     ( address :to )
		//     ( pair ( nat :minTokensBought ) ( timestamp :deadline ) ) )
		NewPairType(
			NewCodeAnno(T_ADDRESS, ":to"),
			NewPairType(
				NewCodeAnno(T_NAT, ":minTokensBought"),
				NewCodeAnno(T_TIMESTAMP, ":deadline"),
			),
			"%xtzToToken",
		),

		// 5 pair %tokenToXtz
		//     ( pair ( address :owner ) ( address :to ) )
		//     ( pair ( nat :tokensSold ) ( pair ( mutez :minXtzBought ) ( timestamp :deadline ) ) ) ) ) ) )
		NewPairType(
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":owner"),
				NewCodeAnno(T_ADDRESS, ":to"),
			),
			NewPairType(
				NewCodeAnno(T_NAT, ":tokensSold"),
				NewPairType(
					NewCodeAnno(T_MUTEZ, ":minXtzBought"),
					NewCodeAnno(T_TIMESTAMP, ":deadline"),
				),
			),
			"%tokenToXtz",
		),

		// 6 pair %tokenToToken
		//     ( pair ( address :outputDexterContract ) ( pair ( nat :minTokensBought ) ( address :owner ) ) )
		//     ( pair ( address :to ) ( pair ( nat :tokensSold ) ( timestamp :deadline ) ) ) )
		NewPairType(
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":outputDexterContract"),
				NewPairType(
					NewCodeAnno(T_NAT, ":minTokensBought"),
					NewCodeAnno(T_ADDRESS, ":owner"),
				),
			),
			NewPairType(
				NewCodeAnno(T_ADDRESS, ":to"),
				NewPairType(
					NewCodeAnno(T_NAT, ":tokensSold"),
					NewCodeAnno(T_TIMESTAMP, ":deadline"),
				),
			),
			"%tokenToToken",
		),

		// 7 ( key_hash %updateTokenPool )
		NewCodeAnno(T_KEY_HASH, "%updateTokenPool"),

		// 8 ( nat %updateTokenPoolInternal ) )
		NewCodeAnno(T_NAT, "%updateTokenPoolInternal"),

		// 9 pair %setBaker ( option key_hash ) bool
		NewPairType(
			NewCode(T_OPTION, NewCode(T_KEY_HASH)),
			NewCode(T_BOOL),
			"%setBaker",
		),

		// 10 address %setManager
		NewCodeAnno(T_ADDRESS, "%setManager"),
	},

	// Note: Kolibri and wXTZ vault interfaces are ambiguous, i.e. wXTZ is a
	// superset of Kolibri and there's little use to distinguish them, so
	// detection is disabled until we know any better
	// IKolibriVault: []Prim{
	// 	// 0 borrow
	// 	NewCodeAnno(T_NAT, "%borrow"),
	// 	// 1 default
	// 	NewCodeAnno(T_UNIT, "%default"),
	// 	// 2 liquidate
	// 	NewCodeAnno(T_UNIT, "%liquidate"),
	// 	// 3 repay
	// 	NewCodeAnno(T_NAT, "%repay"),
	// 	// 4 setDelegate
	// 	NewCodeAnno(T_OPTION, "%setDelegate", NewCode(T_KEY_HASH)),
	// 	// 5 updateState
	// 	NewPairType(
	// 		NewCode(T_ADDRESS),
	// 		NewPairType(
	// 			NewCode(T_NAT),
	// 			NewPairType(
	// 				NewCode(T_INT),
	// 				NewPairType(
	// 					NewCode(T_INT),
	// 					NewCode(T_BOOL),
	// 				),
	// 			),
	// 		),
	// 		"%updateState",
	// 	),
	// 	// 6 withdraw
	// 	NewCodeAnno(T_MUTEZ, "%withdraw"),
	// },
	// // https://medium.com/stakerdao/the-wrapped-tezos-wxtz-beta-guide-6917fa70116e
	// IWXTZVault: []*Prim{
	// 	// 1 default
	// 	NewCodeAnno(T_UNIT, "%default"),
	// 	// 2 setDelegate
	// 	NewCodeAnno(T_OPTION, "%setDelegate", NewCode(T_KEY_HASH)),
	// 	// 3 withdraw
	// 	NewCodeAnno(T_MUTEZ, "%withdraw"),
	// },
}
