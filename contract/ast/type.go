package ast

// Struct is an aggregation of named fields.
// It corresponds to non-top-level `pair` prims, in Michelson.
// It can represent either a parameter's type of entrypoint, or a record in a storage type.
type Struct struct {
	Name          string
	MichelineType string
	Fields        []*Struct
	Type          *Struct
	OriginalType  string
	Key           *Struct
	Value         *Struct
	ParamType     *Struct
	ReturnType    *Struct
	LeftType      *Struct
	RightType     *Struct
	Path          [][]int
	// If true, the expected prim matching to this struct has a flat structure,
	// instead of a tree of pairs.
	Flat bool
}
