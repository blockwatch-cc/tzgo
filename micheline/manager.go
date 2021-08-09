// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// see
// https://gitlab.com/nomadic-labs/mi-cho-coq/merge_requests/29/diffs
// https://gitlab.com/tezos/tezos/blob/510ae152082334b79d3364079cd466e07172dc3a/specs/migration_004_to_005.md#origination-script-transformation

package micheline

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// manager.tz
// https://blog.nomadic-labs.com/babylon-update-instructions-for-delegation-wallet-developers.html
const m_tz_script = "000000c602000000c105000764085e036c055f036d0000000325646f046c000000082564656661756c740501035d050202000000950200000012020000000d03210316051f02000000020317072e020000006a0743036a00000313020000001e020000000403190325072c020000000002000000090200000004034f0327020000000b051f02000000020321034c031e03540348020000001e020000000403190325072c020000000002000000090200000004034f0327034f0326034202000000080320053d036d0342"

// empty store (used as placeholder and to satisfy script decoding which expects storage)
const m_tz_store = "0000001a0a00000015000000000000000000000000000000000000000000"

var (
	manager_tz_code_len    int
	manager_tz_storage_len int
	manager_tz             []byte
)

func init() {
	// unpack manager.tz from binary blob
	buf, err := hex.DecodeString(m_tz_script + m_tz_store)
	if err != nil {
		panic(fmt.Errorf("micheline: decoding manager script: %w", err))
	}
	manager_tz = buf
	manager_tz_code_len = len(m_tz_script) / 2
	manager_tz_storage_len = len(m_tz_store) / 2
}

func MakeManagerScript(managerHash []byte) (*Script, error) {
	script := NewScript()
	if err := script.UnmarshalBinary(manager_tz); err != nil {
		return nil, fmt.Errorf("micheline: unmarshal manager script: %w", err)
	}

	// patch storage
	copy(script.Storage.Bytes, managerHash)

	return script, nil
}

func IsManagerTz(buf []byte) bool {
	return len(buf) >= manager_tz_code_len && bytes.Compare(buf[:manager_tz_code_len], manager_tz[:manager_tz_code_len]) == 0
}

func (p Prim) MigrateToBabylonStorage(managerHash []byte) Prim {
	return NewCode(D_PAIR, NewBytes(managerHash), p)
}

// Patch params, storage and code
func (s *Script) MigrateToBabylonAddDo(managerHash []byte) {
	// add default entrypoint annotation
	s.Code.Param.Args[0].Anno = append([]string{"%default"}, s.Code.Param.Args[0].Anno...)

	// adjust prim type to annotation type
	switch s.Code.Param.Args[0].Type {
	case PrimNullary, PrimUnary, PrimBinary:
		s.Code.Param.Args[0].Type++
	}

	// wrap params
	s.Code.Param.Args[0] = NewCode(
		T_OR,
		NewCodeAnno(T_LAMBDA, "%do", NewCode(T_UNIT), NewCode(T_LIST, NewCode(T_OPERATION))),
		s.Code.Param.Args[0],
	)

	// wrap storage
	s.Code.Storage.Args[0] = NewCode(T_PAIR, NewCode(T_KEY_HASH), s.Code.Storage.Args[0])

	// wrap code
	s.Code.Code.Args[0] = NewSeq(
		NewCode(I_DUP),
		NewCode(I_CAR),
		NewCode(I_IF_LEFT,
			DO_ENTRY(),
			NewSeq(
				// # Transform the inputs to the original script types
				NewCode(I_DIP, NewSeq(NewCode(I_CDR), NewCode(I_DUP), NewCode(I_CDR))),
				NewCode(I_PAIR),
				// # 'default' entrypoint - original code
				s.Code.Code.Args[0],
				// # Transform the outputs to the new script types
				NewCode(I_SWAP),
				NewCode(I_CAR),
				NewCode(I_SWAP),
				UNPAIR(),
				NewCode(I_DIP, NewSeq(NewCode(I_SWAP), NewCode(I_PAIR))),
				NewCode(I_PAIR),
			),
		),
	)

	// migrate storage
	s.Storage = s.Storage.MigrateToBabylonStorage(managerHash)
}

func (s *Script) MigrateToBabylonSetDelegate(managerHash []byte) {
	// add default entrypoint annotation
	s.Code.Param.Args[0].Anno = append([]string{"%default"}, s.Code.Param.Args[0].Anno...)

	// adjust prim type to annotation type
	switch s.Code.Param.Args[0].Type {
	case PrimNullary, PrimUnary, PrimBinary:
		s.Code.Param.Args[0].Type++
	}

	// wrap params
	s.Code.Param.Args[0] = NewCode(
		T_OR,
		NewCode(T_OR,
			NewCodeAnno(T_KEY_HASH, "%set_delegate"),
			NewCodeAnno(T_UNIT, "%remove_delegate"),
		),
		s.Code.Param.Args[0],
	)

	// wrap storage
	s.Code.Storage.Args[0] = NewCode(T_PAIR, NewCode(T_KEY_HASH), s.Code.Storage.Args[0])

	// wrap code
	s.Code.Code.Args[0] = NewSeq(
		NewCode(I_DUP),
		NewCode(I_CAR),
		NewCode(I_IF_LEFT,
			DELEGATE_ENTRY(),
			NewSeq(
				// # Transform the inputs to the original script types
				NewCode(I_DIP, NewSeq(NewCode(I_CDR), NewCode(I_DUP), NewCode(I_CDR))),
				NewCode(I_PAIR),
				// # 'default' entrypoint - original code
				s.Code.Code.Args[0],
				// # Transform the outputs to the new script types
				NewCode(I_SWAP),
				NewCode(I_CAR),
				NewCode(I_SWAP),
				UNPAIR(),
				NewCode(I_DIP, NewSeq(NewCode(I_SWAP), NewCode(I_PAIR))),
				NewCode(I_PAIR),
			),
		),
	)

	// migrate storage
	s.Storage = s.Storage.MigrateToBabylonStorage(managerHash)
}

// Macros
func DO_ENTRY() Prim {
	return NewSeq(
		// # Assert no token was sent:
		NewCode(I_PUSH, NewCode(T_MUTEZ), NewInt64(0)), // PUSH mutez 0 ;
		NewCode(I_AMOUNT), // AMOUNT ;
		ASSERT_CMPEQ(),    // ASSERT_CMPEQ ;
		// # Assert that the sender is the manager
		DUUP(),                      // DUUP ;
		NewCode(I_CDR),              // CDR ;
		NewCode(I_CAR),              // CAR ;
		NewCode(I_IMPLICIT_ACCOUNT), // IMPLICIT_ACCOUNT ;
		NewCode(I_ADDRESS),          // ADDRESS ;
		NewCode(I_SENDER),           // SENDER ;
		IFCMPNEQ( // IFCMPNEQ
			NewSeq(
				NewCode(I_SENDER), //   { SENDER ;
				NewCode(I_PUSH, NewCode(T_STRING), NewString("Only the owner can operate.")), // PUSH string "" ;
				NewCode(I_PAIR),     //     PAIR ;
				NewCode(I_FAILWITH), //     FAILWITH ;
			),
			NewSeq( // # Execute the lambda argument
				NewCode(I_UNIT),                        //     UNIT ;
				NewCode(I_EXEC),                        //     EXEC ;
				NewCode(I_DIP, NewSeq(NewCode(I_CDR))), //     DIP { CDR } ;
				NewCode(I_PAIR),                        //     PAIR ;
			),
		),
	)
}

// 'set_delegate'/'remove_delegate' entrypoints
func DELEGATE_ENTRY() Prim {
	return NewSeq(
		// # Assert no token was sent:
		NewCode(I_PUSH, NewCode(T_MUTEZ), NewInt64(0)), // PUSH mutez 0 ;
		NewCode(I_AMOUNT), // AMOUNT ;
		ASSERT_CMPEQ(),    // ASSERT_CMPEQ ;
		// # Assert that the sender is the manager
		DUUP(),                      // DUUP ;
		NewCode(I_CDR),              // CDR ;
		NewCode(I_CAR),              // CAR ;
		NewCode(I_IMPLICIT_ACCOUNT), // IMPLICIT_ACCOUNT ;
		NewCode(I_ADDRESS),          // ADDRESS ;
		NewCode(I_SENDER),           // SENDER ;
		IFCMPNEQ( // IFCMPNEQ
			NewSeq(
				NewCode(I_SENDER), // SENDER ;
				NewCode(I_PUSH, NewCode(T_STRING), NewString("Only the owner can operate.")), // PUSH string "" ;
				NewCode(I_PAIR),     // PAIR ;
				NewCode(I_FAILWITH), // FAILWITH ;
			),
			NewSeq( // # entrypoints
				NewCode(I_DIP, NewSeq(NewCode(I_CDR), NewCode(I_NIL, NewCode(T_OPERATION)))), // DIP { CDR ; NIL operation } ;
				NewCode(I_IF_LEFT,
					// # 'set_delegate' entrypoint
					NewSeq(
						NewCode(I_SOME),         // SOME ;
						NewCode(I_SET_DELEGATE), // SET_DELEGATE ;
						NewCode(I_CONS),         // CONS ;
						NewCode(I_PAIR),         // PAIR ;
					),
					// # 'remove_delegate' entrypoint
					NewSeq(
						NewCode(I_DROP),                      // DROP ;
						NewCode(I_NONE, NewCode(T_KEY_HASH)), // NONE key_hash ;
						NewCode(I_SET_DELEGATE),              // SET_DELEGATE ;
						NewCode(I_CONS),                      // CONS ;
						NewCode(I_PAIR),                      // PAIR ;
					),
				),
			),
		),
	)
}
