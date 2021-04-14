## TzGo – Go SDK for Tezos by Blockwatch Data

TzGo is the officially supported Tezos Go client library by [Blockwatch](https://blockwatch.cc). This SDK is free to use in commercial and non-commercial projects with a permissive license. Blockwatch is committed to keep interfaces stable, provide long-term support and update TzGo on a regular basis to stay compliant with the most recent Tezos network protocol.

Our main focus is on **correctness**, **stability** and **compliance** with the Tezos protocol.

Current TzGo protocol support

- Florence v009
- Edo v008
- Delphi v007
- Carthage v006
- Babylon v005
- Athens v004
- Alpha v001-v003

### SDK features

TzGo contains a set of features that allow developers to read, monitor, decode, translate, analyze and debug data from the Tezos blockchain, in particular from Tezos smart contracts:

- a low-level **Tezos types library** `tzgo/tezos` to handle all sorts of hashes, addresses and more
- a powerful **Micheline library** `tzgo/micheline` to decode and translate Tezos smart contract data found in calls, storage and bigmaps
- an **RPC library** `tzgo/rpc` for accessing the Tezos Node RPC
- helpers like an efficient base58 en/decoder, hash map

### TzGo Roadmap

When new Tezos protocols are proposed and later deployed we will upgrade TzGo to support new features as soon as practically feasible and as demand for such features exists. For example, we don't fully support Sapling and Lazy Storage updates yet, but will add support in the future as usage of these features becomes more widespread.

V1 of TzGo is focused on read-only data access. We're planning to add support for transaction creation, signing, simulation and injection in the next major release.

### Usage

```sh
go get -u blockwatch.cc/tzgo
```

Then import, using

```go
import (
	"blockwatch.cc/tzgo/tezos"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
)
```

### Micheline Support

Tezos uses [Micheline](https://tezos.gitlab.io/shell/micheline.html) for encoding smart contract data and code. The positive is that Micheline is strongly typed, the downside is its complex and has a few ambiguities that make it hard to use. TzGo contains a library that lets you decode, analyze and construct compliant Micheline data structures from Go.

Micheline uses basic **primitives** for encoding types and values. These primitives can be expressed in JSON and binary format and TzGo can translate between them efficiently. Micheline also supports type **annotations** which are used by high-level languages to express complex data types like records and their field names.

```go
type Prim struct {
	Type      PrimType // primitive type
	OpCode    OpCode   // primitive opcode (invalid on sequences, strings, bytes, int)
	Args      []Prim   // optional nested arguments
	Anno      []string // optional type annotations
	Int       *big.Int // decoded value when Prim is an int
	String    string   // decoded value when Prim is a string
	Bytes     []byte   // decoded value when Prim is a byte sequence
	WasPacked bool     // true when content has been unpacked
}
```

Since Micheline value encoding is quite verbose and can be ambiguous, TzGo supports **unfolding** of raw Micheline into key/value maps using the following Go types and a few access helpers like the `Map()`, `GetInt64()`, `GetAddress()` functions.

- `Type` a simple or complex primitive representing annotated type info
- `Value` a simple or complex primitive representing a Micheline value in combination with its Type
- `Key` a special comparable value that is used as key in maps and bigmaps

Sometimes values are packed into byte sequences using the Michelson PACK instruction and it is desirable to unpack them before processing (e.g. to retrieve UFT8 strings or nested records). TzGo supports `Unpack()` and `UnpackAll()` functions on primitives and values and can detect data types of packed data necessary for unfolding.


### Examples

Below are a few examples showing how to use TzGo to easily access Tezos data in your application.

#### Parsing an address

TzGo comes with a low-level library for basic Tezos data types like protocol enums, addresses, keys, hashes, signatures, protocol parameters and more. All low-level types support encoding/decoding binary and text formats including the necessary validation.

To parse/decode an address and output its components you can do the following:

```go
import "blockwatch.cc/tzgo/tezos"

// parse and panic if invalid
addr := tezos.MustParseAddress("tz3RDC3Jdn4j15J7bBHZd29EUee9gVB1CxD9")

// parse and return error if invalid
addr, err := tezos.ParseAddress("tz3RDC3Jdn4j15J7bBHZd29EUee9gVB1CxD9")
if err != nil {
	fmt.Printf("Invalid address: %v\n", err)
}

// Do smth with the address
fmt.Printf("Address type = %s\n", addr.Type)
fmt.Printf("Address bytes = %x\n", addr.Hash)

```

See [examples/addr.go](https://github.com/blockwatch-cc/tzgo/blob/master/examples/addr.go) for more.

#### Monitoring for new blocks

A Tezos node can notify applications when new blocks are attached to the chain. The Tezos RPC calls this monitor and technically its a long-poll implementation. Here's how to use this feature in TzGo:

```go
import "blockwatch.cc/tzgo/rpc"

// init SDK client
c, _ := rpc.NewClient(nil, "https://mainnet-tezos.giganode.io")

// create block header monitor
mon := rpc.NewBlockHeaderMonitor()
defer mon.Close()

// all SDK functions take a context, here we just use a dummy
ctx := context.TODO()

// register the block monitor with our client
if err := c.MonitorBlockHeader(ctx, mon); err != nil {
	log.Fatalln(err)
}

// wait for new block headers
for {
	head, err := mon.Recv(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	// do smth with the block header
	fmt.Printf("New block %s\n", head.Hash)
}

```

#### Fetch and decode contract storage

```go
import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

// we use the Baker Registry on mainnet as example
addr := tezos.MustParseAddress("KT1ChNsEFxwyCbJyWGSL3KdjeXE28AY1Kaog")

// init RPC client
c, _ := rpc.NewClient(nil, "https://mainnet-tezos.giganode.io")

// fetch the contract's script and most recent storage
script, _ := c.GetContractScript(ctx, addr)

// unfold Micheline storage into human-readable form
val := micheline.NewValue(script.StorageType(), script.Storage)
m, _ := val.Map()
buf, _ := json.MarshalIndent(m, "", "  ")
fmt.Println(string(buf))
```

#### List a contract's bigmaps

```go
import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

// we use the hic et nunc NFT market on mainnet as example
addr := tezos.MustParseAddress("KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9")

// init RPC client
c, _ := rpc.NewClient(nil, "https://mainnet-tezos.giganode.io")

// fetch the contract's script and most recent storage
script, _ := c.GetContractScript(ctx, addr)

// bigmap pointers as []int64
ids := script.BigmapById()

// bigmap pointers as named map[string]int64 (names from type annotations)
named := script.BigmapByName()

```

#### Fetch and decode bigmap values

```go
// init RPC client
c, _ := rpc.NewClient(nil, "https://mainnet-tezos.giganode.io")

// load bigmap type info (use the Baker Registry on mainnet as example)
biginfo, _ := c.GetBigmapInfo(ctx, 17)

// list all bigmap keys
bigkeys, _ := c.GetBigmapKeys(ctx, 17)

// visit each value
for _, key := range bigkeys {
	bigval, _ := c.GetBigmapValue(ctx, 17, key)

	// unfold Micheline type into human readable form
	val := micheline.NewValue(micheline.NewType(biginfo.ValueType), bigval)
	m, _ := val.Map()
	buf, _ := json.MarshalIndent(m, "", "  ")
	fmt.Println(string(buf))
}
```

#### Custom RPC client configuration

TzGo's `rpc.NewClient()` function takes a Go `http.Client` as parameter which you can configure before or after passing it to the library. The example below shows how to set custom timeouts and disable TLS certificate checks (not recommended in production, but useful if you use self-signed certificates during testing).


```
import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"blockwatch.cc/tzgo/rpc"
)


func main() {
	hc := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 180 * time.Second,
			}).Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			}
		}
	}

	c, err := rpc.NewClient(hc, "https://my-private-node.local:8732")
	if err != nil {
		log.Fatalln(err)
	}
}
```


## License

The MIT License (MIT) Copyright (c) 2021 Blockwatch Data Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is furnished
to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.