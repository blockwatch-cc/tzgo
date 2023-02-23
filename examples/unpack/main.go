package main

import (
    m "blockwatch.cc/tzgo/micheline"
    "encoding/hex"
    "fmt"
    "os"
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
    p := m.NewBytes(buf)
    if p, err = p.UnpackAll(); err != nil {
        return err
    }
    fmt.Println(p.Dump())
    return nil
}
