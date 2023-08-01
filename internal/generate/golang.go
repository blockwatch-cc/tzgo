package generate

import (
	"fmt"
	"strings"
	"text/template"

	"blockwatch.cc/tzgo/contract/ast"
	"github.com/iancoleman/strcase"
)

var funcMap = template.FuncMap{
	"receiver":    receiver,
	"pascal":      strcase.ToCamel,
	"camel":       strcase.ToLowerCamel,
	"sub":         func(a, b int) int { return a - b },
	"type":        goType,
	"mkprim":      marshalPrimMethod,
	"pathFromIdx": pathFromIndex,
}

func receiver(typeName string) string {
	if typeName == "" {
		return "_r"
	}
	return "_" + strings.ToLower(string(typeName[0]))
}

func goType(typ *ast.Struct) string {
	switch typ.MichelineType {
	case "nat", "mutez", "int":
		return "*big.Int"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "bytes", "key_hash":
		return "[]byte"
	case "timestamp":
		return "time.Time"
	case "address":
		return "tezos.Address"
	case "key":
		return "tezos.Key"
	case "unit":
		return "struct{}"
	case "chain_id":
		return "tezos.ChainIdHash"
	case "signature":
		return "tezos.Signature"
	case "struct":
		return "*" + strcase.ToCamel(typ.Name)
	case "big_map":
		return fmt.Sprintf("bind.Bigmap[%s, %s]", goType(typ.Key), goType(typ.Value))
	case "lambda":
		return "bind.Lambda"
	case "list", "set":
		return "[]" + goType(typ.Type)
	case "map":
		return fmt.Sprintf("bind.Map[%s, %s]", goType(typ.Key), goType(typ.Value))
	case "option":
		return fmt.Sprintf("bind.Option[%s]", goType(typ.Type))
	case "union":
		return fmt.Sprintf("bind.Or[%s, %s]", goType(typ.LeftType), goType(typ.RightType))
	case "operation", "contract":
		return "any"
		// skip
	}

	return "any"
}

func marshalPrimMethod(typ *ast.Struct) string {
	switch typ.MichelineType {
	case "nat", "int", "mutez":
		return "micheline.NewBig(%s)"
	case "string":
		return "micheline.NewString(%s)"
	case "bool":
		return "tzgoext.MarshalPrimBool(%s)"
	case "bytes", "keyhash":
		return "micheline.NewBytes(%s)"
	case "timestamp":
		return "tzgoext.MarshalPrimTimestamp(%s)"
	case "address":
		return "micheline.NewString(%s.String())"
	case "signature", "key", "chain_id":
		return "micheline.NewBytes(%s.Bytes())"
	case "list":
		return "tzgoext.MarshalPrimSeq[" + goType(typ) + "](%s, tzgoext.MarshalAny)"
	case "lambda", "struct":
		return "%s.Prim"
	default:
		// case *types.Operation:
		// case *types.Contract:
		return "micheline.Prim{}/* %s */"
	}
}

// pathFromIndex returns a path to a right-comb nested Pairs, from the index of a struct's field
// and the total number of fields.
func pathFromIndex(i, n int) string {
	if n == 1 {
		panic("pathFromIndex should not be called when a struct has 1 field")
	}
	if i == n-1 {
		return strings.TrimSuffix(strings.Repeat("r/", n-1), "/")
	}
	return strings.Repeat("r/", i) + "l"
}
