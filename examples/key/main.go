// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Key examples
//
// # Private Keys
// tz1: edsk3nM41ygNfSxVU4w1uAW3G9EnTQEB5rjojeZedLTGmiGRcierVv
// tz2: spsk1zkqrmst1yg2c4xi3crWcZPqgdc9KtPtb9SAZWYHAdiQzdHy7j
// tz3: p2sk3PM77YMR99AvD3fSSxeLChMdiQ6kkEzqoPuSwQqhPsh29irGLC
//
// # Encrypted keys (pw: foo)
// tz1: edesk1uiM6BaysskGto8pRtzKQqFqsy1sea1QRjTzaQYuBxYNhuN6eqEU78TGRXZocsVRJYcN7AaU9JDykwUd8KW
// tz2: spesk246GnDVaqGoYZvKbjrWM1g6xUXnyETXtwZgEYFnP8BQXcaS4rfQQco7C94D1yBmcL1v46Sqy8fXrhBSM7TW
// tz3: p2esk27ocLPLp1JkTWfxByXysGyB7MBDURYJAzAGJLR3XSEV9Nq8wFFdDVXVTwvCwR7Ne2dcUveamjXbvZf3on6T
//
// # Signatures
// edsigtzLBGCyadERX1QsYHKpwnxSxEYQeGLnJGsSkHEsyY8vB5GcNdnvzUZDdFevJK7YZQ2ujwVjvQZn62ahCEcy74AwtbA8HuN
// spsig1RriZtYADyRhyNoQMa6AiPuJJ7AUDcrxWZfgqexzgANqMv4nXs6qsXDoXcoChBgmCcn2t7Y3EkJaVRuAmNh2cDDxWTdmsz
// sigUdRdXYCXW14xqT8mFTMkX4wSmDMBmcW1Vuz1vanGWqYTmuBodueUHGPUsbxgn73AroNwpEBHwPdhXUswzmvCzquiqtcHC

package main

import (
    "encoding/hex"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"
    "syscall"

    "blockwatch.cc/tzgo/tezos"
    "golang.org/x/term"
)

var (
    flags    = flag.NewFlagSet("key", flag.ContinueOnError)
    verbose  bool
    password string
)

func init() {
    flags.Usage = func() {}
    flags.BoolVar(&verbose, "v", false, "be verbose")
    flags.StringVar(&password, "password", "", "password for encrypted keys (may also use env TEZOS_KEY_PASSPHRASE)")
}

func main() {
    if err := flags.Parse(os.Args[1:]); err != nil {
        if err == flag.ErrHelp {
            fmt.Println("Usage: key <cmd> [args]")
            fmt.Println("\nCommands")
            fmt.Printf("  gen <type>                           generate a new key of type\n")
            fmt.Printf("                                         normal: edsk [tz1], spsk [tz2], p2sk [tz3]\n")
            fmt.Printf("                                         encrypted: edesk, spesk, p2esk\n")
            fmt.Printf("  info <key>                           prints info about `key`\n")
            fmt.Printf("  encrypt <key>                        encrypt private `key`\n")
            fmt.Printf("  sign <sk> <msg> [generic]            sign blake2b hash of message `msg` with key\n")
            fmt.Printf("                                         outputs typed signature by default\n")
            fmt.Printf("                                         use `generic` to create a generic signature\n")
            fmt.Printf("  sign-digest <sk> <digest> [generic]  sign blake2b digest with key\n")
            fmt.Printf("                                         outputs typed signature by default\n")
            fmt.Printf("                                         use `generic` to create a generic signature\n")
            fmt.Printf("  verify <pk> <sig> <msg>              verify signature `sig` using pubkey `pk` against blake2b hash of message `msg`\n")
            flags.PrintDefaults()
            os.Exit(0)
        }
        log.Fatalln("Error:", err)
    }

    if err := run(); err != nil {
        log.Fatalln("Error:", err)
    }
}

func run() error {
    n := flags.NArg()
    if n < 1 {
        return fmt.Errorf("Command required")
    }
    switch cmd := flags.Arg(0); cmd {
    case "info":
        if n != 2 {
            return fmt.Errorf("Missing key")
        }
        return info(flags.Arg(1))
    case "gen":
        if n != 2 {
            return fmt.Errorf("Missing key type")
        }
        return gen(flags.Arg(1))
    case "encrypt":
        if n != 2 {
            return fmt.Errorf("Missing key")
        }
        return encrypt(flags.Arg(1))
    case "sign":
        if n < 3 {
            return fmt.Errorf("Missing arguments")
        }
        generic := n == 4 && flags.Arg(3) == "generic"
        return sign(flags.Arg(1), flags.Arg(2), generic)
    case "sign-digest":
        if n < 3 {
            return fmt.Errorf("Missing arguments")
        }
        generic := n == 4 && flags.Arg(3) == "generic"
        return signDigest(flags.Arg(1), flags.Arg(2), generic)
    case "verify":
        if n < 4 {
            return fmt.Errorf("Missing arguments")
        }
        return verify(flags.Arg(1), flags.Arg(2), flags.Arg(3))
    default:
        return fmt.Errorf("Unknown command %q", cmd)
    }
}

func readPassword() tezos.PassphraseFunc {
    pwd := password
    source := "command line"
    if pwd == "" {
        pwd = os.Getenv("TEZOS_KEY_PASSPHRASE")
        source = "env TEZOS_KEY_PASSPHRASE"
    }

    if pwd != "" {
        fmt.Printf("Using password from %s %s", source, strings.Repeat("*", len(pwd)))
        return func() ([]byte, error) { return []byte(pwd), nil }
    } else {
        return func() ([]byte, error) {
            fmt.Print("Enter Password: ")
            buf, err := term.ReadPassword(int(syscall.Stdin))
            fmt.Println()
            return buf, err
        }
    }
}

func info(val string) error {
    if tezos.IsPrivateKey(val) {
        var (
            sk  tezos.PrivateKey
            err error
        )
        if tezos.IsEncryptedKey(val) {
            sk, err = tezos.ParseEncryptedPrivateKey(val, readPassword())
        } else {
            sk, err = tezos.ParsePrivateKey(val)
        }
        if err != nil {
            return err
        }
        if !sk.IsValid() {
            return fmt.Errorf("Invalid private key")
        }

        pk := sk.Public()
        if !pk.IsValid() {
            return fmt.Errorf("Invalid public key")
        }

        fmt.Printf("Key Type         %s\n", sk.Type)
        fmt.Printf("Private Key      %s\n", sk)
        fmt.Printf("Public Key       %s\n", pk)
        fmt.Printf("Address          %s\n", pk.Address())
        fmt.Printf("Private Key Hex  %x (%d)\n", sk.Data, len(sk.Data))
        fmt.Printf("Public Key Hex   %x (%d)\n", pk.Data, len(pk.Data))
    } else if tezos.IsPublicKey(val) {
        pk, err := tezos.ParseKey(val)
        if err != nil {
            return err
        }
        if !pk.IsValid() {
            return fmt.Errorf("Invalid public key")
        }
        fmt.Printf("Key Type         %s\n", pk.Type)
        fmt.Printf("Public Key       %s\n", pk)
        fmt.Printf("Public Key Hex   %x (%d)\n", pk.Data, len(pk.Data))
        fmt.Printf("Address          %s\n", pk.Address())
    } else if tezos.IsSignature(val) {
        sig, err := tezos.ParseSignature(val)
        if err != nil {
            return err
        }
        if !sig.IsValid() {
            return fmt.Errorf("Invalid signature")
        }
        fmt.Printf("Sig Type  %s\n", sig.Type)
        fmt.Printf("Sig Hex   %x (%d)\n", sig.Data, len(sig.Data))

    }
    return nil
}

func gen(val string) error {
    if !tezos.IsPrivateKey(val) {
        return fmt.Errorf("unsupported private key type")
    }

    typ, doEncrypt := tezos.ParseKeyType(val)
    if !typ.IsValid() {
        return fmt.Errorf("invalid private key type")
    }

    sk, err := tezos.GenerateKey(typ)
    if err != nil {
        return err
    }
    pk := sk.Public()

    fmt.Printf("Key Type         %s\n", sk.Type.SkPrefix())
    fmt.Printf("Private Key      %s\n", sk)
    fmt.Printf("Public Key       %s\n", pk)
    fmt.Printf("Address          %s\n", pk.Address())
    fmt.Printf("Private Key Hex  %x (%d)\n", sk.Data, len(sk.Data))
    fmt.Printf("Public Key Hex   %x (%d)\n", pk.Data, len(pk.Data))

    if doEncrypt {
        enc, err := sk.Encrypt(readPassword())
        if err != nil {
            return err
        }
        fmt.Printf("Encrypted Key    %s\n", enc)
    }
    return nil
}

func encrypt(key string) error {
    if !tezos.IsPrivateKey(key) {
        return fmt.Errorf("private key required")
    }

    sk, err := tezos.ParseEncryptedPrivateKey(key, readPassword())
    if err != nil {
        return err
    }
    pk := sk.Public()

    fmt.Printf("Key Type         %s\n", sk.Type.SkPrefix())
    fmt.Printf("Private Key      %s\n", sk)
    fmt.Printf("Public Key       %s\n", pk)
    fmt.Printf("Address          %s\n", pk.Address())
    fmt.Printf("Private Key Hex  %x (%d)\n", sk.Data, len(sk.Data))
    fmt.Printf("Public Key Hex   %x (%d)\n", pk.Data, len(pk.Data))

    enc, err := sk.Encrypt(readPassword())
    if err != nil {
        return err
    }
    fmt.Printf("Encrypted Key    %s\n", enc)
    return nil
}

func sign(key, msg string, generic bool) error {
    sk, err := tezos.ParseEncryptedPrivateKey(key, readPassword())
    if err != nil {
        return err
    }
    pk := sk.Public()
    digest := tezos.Digest([]byte(msg))
    sig, err := sk.Sign(digest[:])
    if err != nil {
        return err
    }

    fmt.Printf("Key Type         %s\n", sk.Type)
    fmt.Printf("Private Key      %s\n", sk)
    fmt.Printf("Public Key       %s\n", pk)
    fmt.Printf("Address          %s\n", pk.Address())
    fmt.Printf("Msg digest       %x\n", digest[:])
    fmt.Printf("Signature        %s\n", sig)
    fmt.Printf("Signature Hex    %x\n", sig.Data)
    if generic {
        fmt.Printf("Generic Sig      %s\n", sig.Generic())
    }
    return nil
}

func signDigest(key, dgst string, generic bool) error {
    sk, err := tezos.ParseEncryptedPrivateKey(key, readPassword())
    if err != nil {
        return err
    }
    pk := sk.Public()
    digest, err := hex.DecodeString(dgst)
    if err != nil {
        return err
    }
    sig, err := sk.Sign(digest[:])
    if err != nil {
        return err
    }

    fmt.Printf("Key Type         %s\n", sk.Type)
    fmt.Printf("Private Key      %s\n", sk)
    fmt.Printf("Public Key       %s\n", pk)
    fmt.Printf("Address          %s\n", pk.Address())
    fmt.Printf("Msg digest       %x\n", digest[:])
    fmt.Printf("Signature        %s\n", sig)
    fmt.Printf("Signature Hex    %x\n", sig.Bytes())
    if generic {
        fmt.Printf("Generic Sig      %s\n", sig.Generic())
    }
    return nil
}
func verify(key, sig, msg string) error {
    pk, err := tezos.ParseKey(key)
    if err != nil {
        return err
    }
    s, err := tezos.ParseSignature(sig)
    if err != nil {
        return err
    }
    digest := tezos.Digest([]byte(msg))
    if err := pk.Verify(digest[:], s); err == nil {
        fmt.Println("Signature OK")
    } else {
        return err
    }
    return nil
}
