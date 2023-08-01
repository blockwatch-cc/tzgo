package ast

type Contract struct {
	// Name of the Contract.
	// Can be inferred from its metadata, if it implements TZIP-16 (TODO).
	Name string
	// Micheline script of the contract.
	Micheline string
	// Callable Entrypoints of the Contract.
	Entrypoints []*Entrypoint
	// Getters are TZIP-4 views (which should not be confused with Hangzhou views).
	// Although they are entrypoints, they require to be handled differently from
	// regular entrypoints.
	Getters []*Getter
	// Type of the Contract's Storage.
	Storage *Struct
	// Bigmaps referenced in the Contract's Storage.
	Bigmaps []any
}
