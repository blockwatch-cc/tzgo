package bind

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
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

func (b *Bigmap[K, V]) UnmarshalPrim(prim micheline.Prim) error {
	*b = NewBigmap[K, V](prim.Int.Int64())
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
