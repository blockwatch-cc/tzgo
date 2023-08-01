package bind

import "blockwatch.cc/tzgo/micheline"

// Lambda is raw Michelson code represented as a Prim tree.
type Lambda struct {
	micheline.Prim
}

func (l *Lambda) MarshalPrim(_ bool) (micheline.Prim, error) {
	return l.Prim, nil
}

func (l *Lambda) UnmarshalPrim(prim micheline.Prim) error {
	l.Prim = prim
	return nil
}
