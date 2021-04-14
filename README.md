## TzGo – Go SDK for Tezos by Blockwatch Data

TzGo is the officially supported Tezos Go client library by [Blockwatch](https://blockwatch.cc). This SDK is free to use in commercial and non-commercial projects with a permissive license. Blockwatch is committed to keep interfaces stable, provide long-term support and update tzgo on a regular basis to stay compliant with the most recent Tezos network protocol.

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

- an RPC library `tzgo/rpc` for accessing the Tezos Node RPC
- a low-level Tezos types library `tzgo/tezos` to handle all sorts of hashes, addresses and more
- a powerful smart contract data library `tzgo/micheline` to decode and translate contract calls, storage and bigmap contents
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

### Examples

Here's how to use the TzGo SDK to easily access Tezos data in your application


#### Configuring RPC client


#### Monitoring for new blocks
MonitorBlockHeader


#### Parsing an address


#### Fetching and decoding storage content


#### Fetching and decoding bigmap content




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