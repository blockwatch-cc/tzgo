package bind

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"github.com/stretchr/testify/require"
	"math/big"
	"reflect"
	"testing"
)

var testAddress = tezos.MustParseAddress("tz1KqTpEZ7Yob7QbPE4Hy4Wo8fHG8LhKxZSx")

func TestUnmarshalPrim(t *testing.T) {
	cases := map[string]struct {
		prim micheline.Prim
		dst  any
		want any
		wErr error
	}{
		"string":        {prim: micheline.NewString("hello"), dst: "", want: "hello"},
		"bigInt":        {prim: micheline.NewInt64(42), dst: &big.Int{}, want: big.NewInt(42)},
		"bytes":         {prim: micheline.NewBytes([]byte{4, 2}), dst: []byte{}, want: []byte{4, 2}},
		"address":       {prim: micheline.NewString(testAddress.String()), dst: tezos.Address{}, want: testAddress},
		"string slice":  {prim: micheline.NewSeq(micheline.NewString("1"), micheline.NewString("2")), dst: []string{}, want: []string{"1", "2"}},
		"struct":        {prim: micheline.NewPair(micheline.NewString("aaa"), micheline.NewPair(micheline.NewInt64(42), micheline.NewBytes([]byte{1, 2, 3}))), dst: (*unmarshaler)(nil), want: &unmarshaler{"aaa", big.NewInt(42), []byte{1, 2, 3}}},
		"nested struct": {prim: micheline.NewPair(micheline.NewPair(micheline.NewString("aaa"), micheline.NewPair(micheline.NewInt64(42), micheline.NewBytes([]byte{1, 2, 3}))), micheline.NewString("uuu")), dst: (*nestedUnmarshaler)(nil), want: &nestedUnmarshaler{&unmarshaler{"aaa", big.NewInt(42), []byte{1, 2, 3}}, "uuu"}},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			val := reflect.New(reflect.TypeOf(c.dst))
			val.Elem().Set(reflect.ValueOf(c.dst))

			err := UnmarshalPrim(c.prim, val.Interface())
			if c.wErr != nil {
				require.ErrorIs(t, err, c.wErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.want, val.Elem().Interface())
		})
	}
}

type unmarshaler struct {
	A string
	B *big.Int
	C []byte
}

func (u *unmarshaler) UnmarshalPrim(prim micheline.Prim) error {
	return UnmarshalPrimPaths(prim, map[string]any{"l": &u.A, "r/l": &u.B, "r/r": &u.C})
}

type nestedUnmarshaler struct {
	U *unmarshaler
	S string
}

func (u *nestedUnmarshaler) UnmarshalPrim(prim micheline.Prim) error {
	return UnmarshalPrimPaths(prim, map[string]any{"l": &u.U, "r": &u.S})
}
