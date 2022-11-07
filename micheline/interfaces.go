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

func (s *Script) ImplementsStrict(i Interface) bool {
	eps, _ := s.Entrypoints(true)
	if len(eps) == 0 {
		return false
	}
	return i.MatchesStrict(eps)
}

func (s *Script) Interfaces() Interfaces {
	eps, _ := s.Entrypoints(true)
	if len(eps) == 0 {
		return nil
	}
	iv := make(Interfaces, 0)
	for _, i := range WellKnownInterfaces {
		if !i.Matches(eps) {
			continue
		}
		iv = append(iv, i)
	}
	return iv
}

func (s *Script) InterfacesStrict() Interfaces {
	eps, _ := s.Entrypoints(true)
	if len(eps) == 0 {
		return nil
	}
	iv := make(Interfaces, 0)
	for _, i := range WellKnownInterfaces {
		if !i.MatchesStrict(eps) {
			continue
		}
		iv = append(iv, i)
	}
	return iv
}

type Interface string

func (m Interface) String() string {
	return string(m)
}

// Checks if a contract implements all entrypoints required by a standard
// interface without requiring argument labels to match. This is a looser
// definition of interface compliance, but in line with the Michelson
// type system which ignores annotation labels for type equality.
//
// This check uses extracted Typedefs to avoid issues when the Micheline
// primitive structure diverges from the defined interface (e.g. due to
// comb type unfolding).
func (m Interface) Matches(e Entrypoints) bool {
	for _, spec := range InterfaceSpecs[m] {
		// use Typedef to avoid differences in data encoding with comb pairs
		specType := NewType(spec).Typedef("")

		var matched bool
		for _, ep := range e {
			// check entrypoint name
			if ep.Name != spec.GetVarAnnoAny() {
				continue
			}

			// check entrypoint type
			epType := NewType(*ep.Prim).Typedef("")
			if specType.Equal(epType) {
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

// Checks if a contract strictly implements all standard interface
// entrypoints including argument types and argument names (annotations).
//
// This check uses extracted Typedefs to avoid issues when the Micheline
// primitive structure diverges from the defined interface (e.g. due to
// comb type unfolding).
func (m Interface) MatchesStrict(e Entrypoints) bool {
	for _, spec := range InterfaceSpecs[m] {
		// use Typedef to avoid differences in data encoding with comb pairs
		specType := NewType(spec).Typedef("")

		var matched bool
		for _, ep := range e {
			// check entrypoint name
			if ep.Name != spec.GetVarAnnoAny() {
				continue
			}

			// check entrypoint type
			epType := NewType(*ep.Prim).Typedef("")
			if specType.StrictEqual(epType) {
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

func (m Interface) Contains(e Entrypoint) bool {
	epType := NewType(*e.Prim).Typedef("")
	for _, spec := range InterfaceSpecs[m] {
		// check entrypoint name
		if e.Name != spec.GetVarAnnoAny() {
			continue
		}
		// check entrypoint type
		specType := NewType(spec).Typedef("")
		if specType.Equal(epType) {
			return true
		}
	}
	return false
}

func (m Interface) ContainsStrict(e Entrypoint) bool {
	epType := NewType(*e.Prim).Typedef("")
	for _, spec := range InterfaceSpecs[m] {
		// check entrypoint name
		if e.Name != spec.GetVarAnnoAny() {
			continue
		}
		// check entrypoint type
		specType := NewType(spec).Typedef("")
		if specType.StrictEqual(epType) {
			return true
		}
	}
	return false
}

func (m Interface) TypeOf(name string) Type {
	for _, v := range InterfaceSpecs[m] {
		if v.GetVarAnnoAny() == name {
			return NewType(v)
		}
	}
	return Type{}
}

func (m Interface) PrimOf(name string) Prim {
	for _, v := range InterfaceSpecs[m] {
		if v.GetVarAnnoAny() == name {
			return v
		}
	}
	return Prim{}
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

	WellKnownInterfaces = []Interface{
		IManager,
		ISetDelegate,
		ITzip5,
		ITzip7,
		ITzip12,
	}
)

// WellKnownInterfaces contains entrypoint types for standard call interfaces and other
// known contracts.
var InterfaceSpecs = map[Interface][]Prim{
	// manager.tz
	IManager: {
		// 1 (lambda %do unit (list operation))
		NewCodeAnno(T_LAMBDA, "%do", NewCode(T_UNIT), NewCode(T_LIST, NewCode(T_OPERATION))),
		// 2 (unit %default)
		NewCodeAnno(T_UNIT, "%default"),
	},
	// generic set delegate interface
	ISetDelegate: {
		// option %setDelegate key_hash
		NewCodeAnno(T_OPTION, "%setDelegate", NewCode(T_KEY_HASH)),
	},
	// Tzip 5 a.k.a FA1
	// https://gitlab.com/tzip/tzip/-/blob/master/proposals/tzip-5/tzip-5.md
	ITzip5: {
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
	ITzip7: {
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
	ITzip12: {
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
				NewCodeAnno(T_ADDRESS, "%from_"),
				NewCodeAnno(T_LIST, "%txs",
					NewPairType(
						NewCodeAnno(T_ADDRESS, "%to_"),
						NewPairType(
							NewCodeAnno(T_NAT, "%token_id"),
							NewCodeAnno(T_NAT, "%amount"),
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
		// DISABLED because of some broken tokens which get mis-classified
		// NewPairType(
		// 	NewCodeAnno(T_LIST, "%requests",
		// 		NewPairType(
		// 			NewCodeAnno(T_ADDRESS, "%owner"),
		// 			NewCodeAnno(T_NAT, "%token_id"),
		// 		),
		// 	),
		// 	NewCodeAnno(T_CONTRACT, "%callback",
		// 		NewCode(T_LIST,
		// 			NewPairType(
		// 				NewPairType(
		// 					NewCodeAnno(T_ADDRESS, "%owner"),
		// 					NewCodeAnno(T_NAT, "%token_id"),
		// 					"%request",
		// 				),
		// 				NewCodeAnno(T_NAT, "%balance"),
		// 			),
		// 		),
		// 	),
		// 	"%balance_of",
		// ),
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
					NewCodeAnno(T_ADDRESS, "%owner"),
					NewPairType(
						NewCodeAnno(T_ADDRESS, "%operator"),
						NewCodeAnno(T_NAT, "%token_id"),
					),
					"%add_operator",
				),
				NewPairType(
					NewCodeAnno(T_ADDRESS, "%owner"),
					NewPairType(
						NewCodeAnno(T_ADDRESS, "%operator"),
						NewCodeAnno(T_NAT, "%token_id"),
					),
					"%remove_operator",
				),
			),
		),
	},
}
