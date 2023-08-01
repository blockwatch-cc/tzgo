package bind

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"time"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"github.com/pkg/errors"
)

// Map is a map type used to interact with Tezos smart contracts.
//
// Go's map type cannot be used in this context, because the "comparable"
// types are different in Go and in Tezos specification.
type Map[K, V any] struct {
	m        map[hashType]MapEntry[K, V]
	hashFunc func(any) hashType
}

type MapEntry[K, V any] struct {
	Key   K
	Value V
}

func MakeMap[K, V any](size ...int) Map[K, V] {
	h := hashFunc(zero[K]())

	if len(size) > 0 {
		return Map[K, V]{
			m:        make(map[hashType]MapEntry[K, V], size[0]),
			hashFunc: h,
		}
	}
	return Map[K, V]{
		m:        make(map[hashType]MapEntry[K, V]),
		hashFunc: h,
	}
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	value, ok := m.m[m.hashFunc(key)]
	return value.Value, ok
}

func (m *Map[K, V]) Set(key K, value V) {
	m.m[m.hashFunc(key)] = MapEntry[K, V]{Key: key, Value: value}
}

func (m *Map[K, V]) Entries() []MapEntry[K, V] {
	entries := make([]MapEntry[K, V], 0, len(m.m))
	for _, e := range m.m {
		entries = append(entries, e)
	}
	return entries
}

func (m Map[K, V]) MarshalPrim(optimized bool) (micheline.Prim, error) {
	entries := make([]micheline.Prim, 0, len(m.m))
	for key, value := range m.m {
		keyPrim, err := MarshalPrim(key, optimized)
		if err != nil {
			return micheline.Prim{}, errors.Wrap(err, "failed to marshal key")
		}
		valuePrim, err := MarshalPrim(value, optimized)
		if err != nil {
			return micheline.Prim{}, errors.Wrap(err, "failed to marshal value")
		}
		entries = append(entries, micheline.NewCode(micheline.D_ELT, keyPrim, valuePrim))
	}

	return micheline.NewSeq(entries...), nil
}

func (m *Map[K, V]) UnmarshalPrim(prim micheline.Prim) error {
	if prim.Type != micheline.PrimSequence {
		return errors.Errorf("invalid micheline type for Map: %s", prim.Type)
	}

	*m = MakeMap[K, V](len(prim.Args))
	for _, entry := range prim.Args {
		if entry.OpCode != micheline.D_ELT {
			return errors.Errorf("map entries should be ELT, got %s", prim.OpCode)
		}
		if len(entry.Args) != 2 {
			return errors.New("prim ELT should have 2 args")
		}
		var key K
		var value V
		if err := UnmarshalPrim(entry.Args[0], &key); err != nil {
			return errors.Wrap(err, "failed to unmarshal key")
		}
		if err := UnmarshalPrim(entry.Args[1], &value); err != nil {
			return errors.Wrap(err, "failed to unmarshal value")
		}
		m.Set(key, value)
	}

	return nil
}

func (m Map[K, V]) Format(f fmt.State, verb rune) {
	vSharp := verb == 'v' && f.Flag('#')

	if vSharp {
		// Write the entire Map type with %#v
		_, _ = fmt.Fprintf(f, "bind.Map[%T]%T", zero[K](), zero[V]())

		// Special case if m is nil: we print (nil) instead of {nil}
		if m.m == nil {
			_, _ = io.WriteString(f, "(nil)")
			return // we can return here
		} else {
			_, _ = f.Write([]byte{'{'})
		}
	} else {
		_, _ = io.WriteString(f, "bind.Map[")
	}

	// Begin of content

	first := true
	for _, e := range m.m {
		if first {
			first = false
		} else {
			_, _ = f.Write([]byte{' '})
		}

		_, _ = fmt.Fprintf(f, "%v:%v", e.Key, e.Value)
	}

	// End of content

	if vSharp {
		_, _ = f.Write([]byte{'}'})
	} else {
		_, _ = f.Write([]byte{']'})
	}
}

type hashType [sha256.Size]byte

// hashFunc
//
// Allowed types:
//   - string
//   - []byte
//   - bool
//   - *big.Int
//   - time.Time
//   - tezos.Address
//   - tezos.Key
//   - tezos.Signature
//   - tezos.ChainIdHash
//   - bind.Or
//   - bind.Option
//   - pair
func hashFunc(keyType any) func(any) hashType {
	switch keyType.(type) {
	case string:
		return func(k any) hashType { return hashBytes([]byte(k.(string))) }
	case []byte:
		return func(k any) hashType { return hashBytes(k.([]byte)) }
	case bool:
		return func(k any) hashType {
			if k.(bool) {
				return hashType{1}
			}
			return hashType{0}
		}
	case *big.Int:
		return func(k any) hashType { return hashBytes(k.(*big.Int).Bytes()) }
	case time.Time:
		return func(k any) hashType { return hashBytes(must(k.(time.Time).MarshalBinary())) }
	case tezos.Address:
		return func(k any) hashType { return hashBytes(k.(tezos.Address).Hash()) }
	case tezos.Key:
		return func(k any) hashType { return hashBytes(k.(tezos.Key).Bytes()) }
	case tezos.ChainIdHash:
		return func(k any) hashType { return hashBytes(k.(tezos.ChainIdHash).Bytes()) }
	case keyHasher:
		return func(k any) hashType { return k.(keyHasher).keyHash() }
	}

	panic(fmt.Sprintf("%T keys is not supported", keyType))
}

type keyHasher interface {
	keyHash() hashType
}

func hashBytes(b []byte) hashType {
	h := sha256.New()
	h.Write(b)
	var hash hashType
	copy(hash[:], h.Sum(nil))
	return hash
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func zero[T any]() T {
	var t T
	return t
}
