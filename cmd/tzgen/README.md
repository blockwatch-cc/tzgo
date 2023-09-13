# TzGen

TzGen is a code generator that enables creating Golang struct types and smart contract interface bindings based on a contract's entrypoint specification. The generated code can be used to deploy, bootstrap, and call any chosen smart contract and interact with all of its entrypoints including views. The bindings read/decode data found in transaction parameters, transaction receipts (storage and bigmap updates), and contract storage as well as write/encode type-conform parameters for sending transactions.

## Installation

```bash
go install blockwatch.cc/tzgo/cmd/tzgen
```

## Using TzGen

When using TzGen you don't have to worry about details of Micheline or TzGo. TzGen produces all interfaces and types you'll need to call any of your contract's entrypoints, read its storage and bigmap entries.

You can run tzgen manually which will write an auto-generated Go source file that you can build together with your project and check in to your repository.

### From a deployed contract

```bash
tzgen -name <name> -pkg <pkg> -address <addr> -out <file.go>
```

The endpoint is `https://rpc.tzstats.com` by default, but can be overridden with the `-endpoint` flag.

### From a Micheline file

```bash
tzgen -name <name> -pkg <pkg> -src <file.json> -out <file.go>
```

### Go Generate

You can also use tzgen in combination with the go generate tool if you want to create fresh interface definitions at build time. To use go generate you need to do two things:

1. Add a go generate comment into a Go source code file:
```
//go:generate go run -mod=mod blockwatch.cc/tzgo/cmd/tzgen -name <name> -pkg <pkg> -address <addr> -out <file.go>
```

2. Call go generate `./path-to-your-pkg-or-cmd` to make Go call whatever script you have defined in 1:
```
go generate ./cmd/myprog
```

## Renaming Structs

Some structs don't have annotations in the contract's script.
In this case, an auto-generated name is given.

It is possible to give a configuration map to tzgen, to map these auto-generated names to the one you want.

To do so, pass a yaml to tzgen with the `-fixup` flag.

Example of a fixup file:

```yaml
FA2NFTRecord3:
  name: OperatorForAll
  fields:
    Field0: Addr
    Field1: Owner

FA2NFTRecord5:
  name: BalanceOfRequest

FA2NFTRequest:
  equals: FA2NFTRecord5
```
