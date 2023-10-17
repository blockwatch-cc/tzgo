package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"blockwatch.cc/tzgo/micheline"
	m "blockwatch.cc/tzgo/micheline"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("Expecting packed bytes as hex string")
	}
	buf, err := hex.DecodeString(os.Args[1])
	if err != nil {
		return err
	}
	if buf[0] == 0x5 {
		p := m.NewBytes(buf)
		if p, err = p.UnpackAll(); err != nil {
			return err
		}
		fmt.Println(p.Dump())
	} else {
		var p micheline.Prim
		if err := p.UnmarshalBinary(buf); err != nil {
			return err
		}
		fmt.Println(p.Dump())
	}
	return nil
}
