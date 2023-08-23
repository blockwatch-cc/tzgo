package bind

import (
	"context"
	"fmt"
	"strconv"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"github.com/pkg/errors"
)

// Bigmap is a handle to a Tezos bigmap.
//
// It has two type parameters, K and V, which are determined when a contract
// script is parsed.
//
// Before using a Bigmap to Get a value, its RPC client must be set with SetRPC.
type Bigmap[K, V any] struct {
	id      int64
	keyType *micheline.Type
	rpc     RPC
	m       []MapEntry[K, V]
}

// NewBigmap returns a new Bigmap that points to the given bigmap id.
//
// The type parameters must match the key and value type of the corresponding
// bigmap.
func NewBigmap[K, V any](id int64) Bigmap[K, V] {
	return Bigmap[K, V]{id: id}
}

// ID returns the id of the bigmap.
func (b *Bigmap[K, B]) ID() int64 {
	return b.id
}

func (b *Bigmap[K, V]) SetContent(elt ...MapEntry[K, V]) {
	b.m = elt
}

// SetRPC defines the client to use when getting a value from the bigmap.
func (b *Bigmap[K, B]) SetRPC(client RPC) *Bigmap[K, B] {
	b.rpc = client
	return b
}

// SetKeyType forces the key micheline type to use, when marshaling a key
// into an expression hash.
func (b *Bigmap[K, B]) SetKeyType(keyType micheline.Type) *Bigmap[K, B] {
	b.keyType = &keyType
	return b
}

// Get the value corresponding to the given key.
//
// This makes a rpc call, so SetRPC must have been called before calling this.
//
// If the key doesn't exist in the bigmap, an ErrKeyNotFound is returned.
func (b *Bigmap[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	if b.rpc == nil {
		return v, errors.New("rpc not set in bigmap")
	}

	keyVal, err := MarshalPrim(key, true)
	if err != nil {
		return v, err
	}

	if b.keyType == nil {
		b.SetKeyType(keyVal.BuildType())
	}

	k, err := micheline.NewKey(*b.keyType, keyVal)
	if err != nil {
		return v, err
	}

	keyHash := k.Hash()

	prim, err := b.rpc.GetBigmapValue(ctx, b.id, keyHash, rpc.Head)
	if err != nil {
		var httpError rpc.HTTPError
		if errors.As(err, &httpError) && httpError.StatusCode() == 404 {
			return v, &ErrKeyNotFound{Key: keyHash.String()}
		}
		return v, err
	}

	if len(prim.Args) > 2 {
		prim.Type = micheline.PrimSequence
		prim = prim.FoldPair()
	}

	if err = UnmarshalPrim(prim, &v); err != nil {
		return v, err
	}

	return v, nil
}

func (b Bigmap[K, V]) String() string {
	return "Bigmap#" + strconv.Itoa(int(b.id))
}

func (b Bigmap[K, V]) MarshalPrim(optimized bool) (micheline.Prim, error) {
	entries := make([]micheline.Prim, 0, len(b.m))
	for _, value := range b.m {
		keyPrim, err := MarshalPrim(value.Key, optimized)
		if err != nil {
			return micheline.Prim{}, errors.Wrap(err, "failed to marshal key")
		}
		valuePrim, err := MarshalPrim(value.Value, optimized)
		if err != nil {
			return micheline.Prim{}, errors.Wrap(err, "failed to marshal value")
		}
		entries = append(entries, micheline.NewMapElem(keyPrim, valuePrim))
	}
	return micheline.NewSeq(entries...), nil
}

func (b *Bigmap[K, V]) UnmarshalPrim(prim micheline.Prim) error {
	if prim.Int != nil {
		*b = NewBigmap[K, V](prim.Int.Int64())
		return nil
	}
	if prim.Type != micheline.PrimSequence {
		return fmt.Errorf("not supported type for unmarshall")
	}
	*b = NewBigmap[K, V](0)
	for _, entry := range prim.Args {
		if entry.OpCode != micheline.D_ELT {
			return errors.Errorf("bigmap entries should be ELT, got %s", prim.OpCode)
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
		b.m = append(b.m, MapEntry[K, V]{Key: key, Value: value})
	}
	return nil
}

type ErrKeyNotFound struct {
	Key string
}

func (e *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("bigmap key not found %q", e.Key)
}

func (e *ErrKeyNotFound) Is(target error) bool {
	other, ok := target.(*ErrKeyNotFound)
	if !ok {
		return false
	}
	return e.Key == other.Key
}
