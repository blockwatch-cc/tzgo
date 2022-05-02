## Create and use private keys in Tezos

Use TzGo to generate and use private keys for joy and pleasure.

### Usage

```sh
Usage: key <cmd> [args]

Commands
  gen <type>                           generate a new key of type
                                         normal: edsk [tz1], spsk [tz2], p2sk [tz3]
                                         encrypted: edesk, spesk, p2esk
  info <key>                           prints info about `key`
  encrypt <key>                        encrypt private `key`
  sign <sk> <msg> [generic]            sign blake2b hash of message `msg` with key
                                         outputs typed signature by default
                                         use `generic` to create a generic signature
  sign-digest <sk> <digest> [generic]  sign blake2b digest with key
                                         outputs typed signature by default
                                         use `generic` to create a generic signature
  verify <pk> <sig> <msg>              verify signature `sig` using pubkey `pk` against blake2b hash of message `msg`
  -password string
      password for encrypted keys (may also use env TEZOS_KEY_PASSPHRASE)
  -v  be verbose
  ```

### Examples

```sh
# create new private key
go run . gen edsk

# show address and pubkey
go run . info edsk4KVDTs6J69y5EYELh1WRPBsuj5fRRJdRonyE2P5uaeEi7hqSNc

# encyrpt the private key to pretoct ot
go run . encrypt edsk4KVDTs6J69y5EYELh1WRPBsuj5fRRJdRonyE2P5uaeEi7hqSNc

# sign a message (actually, this signs the blake2b hash of the message)
go run . sign edsk4KVDTs6J69y5EYELh1WRPBsuj5fRRJdRonyE2P5uaeEi7hqSNc "Cool stuff"

# verify a signature
go run . verify edpkuHNzxQnjV5jNAKWBDjmb2fyEZH2KrsFywwBHELMKozTGU9RDEU edsigtjPe9hmQvYcn9P24zEHT4neheDQ39jCwAoKF4pLxqYyq9cVpkLBHUvivmEVdXKDYE2s1cqQd8YEpD2irG13W6d7LdeYbSs "Cool stuff"
```

