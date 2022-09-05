# Changelog

## v1.14.0

* Simplify protocol constants (drop constants not required for indexing or to send txs)
* Add Kathmandu EMIT opcode
* Add parsing for Kathmandu events
* Add parsing for Kathmandu ops `vdf_revelation` and `increase_paid_storage`
* Add op type enums and tags for Kathmandu TORU, SCORU and DAL ops
* Fix block offset calculations (regression from 1.13.2)
* Improve Micheline type handling (nested lists)
* Fix fee estimation
* Add Zarith helpers
* Improve bigmap detection from Micheline value
* Add simple Micheline type checker

## v1.13.2

* Don't use zero address for contract deployments
* Fix transaction burn accounting
* Fix origination burn accounting
* Remove unnecessary start block offsets from recent testnet configs
* Fix RPC execution receipt error handling
* Add simple Zarith arithmetic funcs
* Add log to contract package
* Fix negative block offset in RPC
* Add Deku contract address support
* Harden address set against hash collisions
* Read more token metdata
* Improve decoding Go structs from Micheline primitives
* Improve bigmap detection
* Fix ghostnet start cycle

## v1.13.1

* Add Ghostnet support
* Clear old code after constants have been expanded
* Handle top-level constant in script code
* Update value render testdata
* Skip hashing invalid prims
* Add jakarta to list of mainnet protocols


## v1.13.0

* Add method to expose token uri
* Support gas/milligas selection
* Rollup data decoding
* Rename nft ledger types
* Add IPFS url helper
* Improve typedefs
* Calculate block in cycle position
* Fix marshal empty primitive
* Fix Jakarta voting power change
* Add noview token type
* Add fa2 balance helper
* Make call args chainable
* Add mutez prim helper
* Support address add/list for remote signer
* Fix merging params
* Fix min fee, add min-fee helper
* Client function to read contract balance
* Expose contract errors
* Fix RPC run call
* Add call option to select BlockID when simulating operations
* Fix writing binary block header
* Simple rollup RPC support
* Add cycles per vote period constant
* Add ticket value wrapper
* Add LB toggle vote
* Fix bootstrap protocol identification
* Add Jakarta support (params, hashes, opcodes, op types)
* Fetch chain id if not yet set
* Refactor lazy bigmap updates
* Fix initializing chain params in RPC client
* Preserve API key privacy
* Fix BSON protocol data decode and encode

## v1.12.3

* Add quipuswap pool example
* Add remote signer example
* Fix remote signer package name
* Improve path extraction from raw Micheline
* Fix getting integers from Value
* Add decimal string export to Zarith numbers
* Improve contract and token handling
* Fix generic signature bytes output
* Decode contract and token metadata
* Support run_code execution
* Add IPFS Url and allow access to http client

## v1.12.2

* Add method to create TxArgs
* Add default testnet params
* Allow custom contract calls via TxArgs
* Read contract address after origination
* Avoid double fee estimation during Send()
* Add a remote signer client
* Refactor signer interface for compatibility with remote signers

## v1.12.1

* Add simple transfer example
* Refactor tx and contract call sending
* Add FA examples and convenience methods
* Add in-memory signer
* Support more tenderbake op type changes, add tests
* Refactor params loading to avoid uninitialized chain id
* Add base tests
* Optimize primitive JSON encoding
* Add prim size counter
* Add optimized script decoder
* Default to Ithaca
* Improve encoding performance
* Fix vesting contract test case
* Fix translating embedded Elts in bigmap origination data
* Fix comparing zero length annots
* Cleanup old protocol settings
* Update entrypoint testcases
* Update testcases to new bigmap ptr type

## v1.12.0

* Rename entrypoint call to name
* Use uint64 for script and storage hashes
* Use int64 as bigmap pointers
* Add schema decode helper
* Update rights RPC call
* Add known burn address
* Fix imports
* Add initial NFT ledger unmarshalling support
* Add type checked prim path getters
* Remove deprecated hashes and flags
* Support Ithaca snapshot RPC, cleanup params
* Support pre-Carthage contract flags
* Add helper functions to hashes
* Remove v002 bug handling, support Ithaca snapshot offset change
* Remove indexer-centric operation types
* Fix set deposits limit encoding

## v1.12-rc1

- Refactored wallet functionality into rpc package
- Refactored op construction
- Refactored cost and limit types
- Initial contract and token support
- Added Ithaca constants, hashes, types and RPC updates

## v1.11-rc0

This is the first release of TzGo that allows to send transaction. All types and interfaces are engineered to be easily composable. We start with essential low level types/functions for public and private key handling, operation encoding/signing/broadcast and mempool/block monitoring. This feature set is already fully sufficient to build end-user applications including the possibility to send smart contract calls, but may at times be less convenient. To simplify complex use cases we will introduce higher order functionality in the next release candidate.

**Package `tezos`**

- New: parse, generate, sign, verify operations Ed25519, Secp256k1 and P256 private and public keys
- New: reading and writing of encrypted keys
- New: explicit EndorsementWithSlot support
- Refactored Zarith encoder, added unsigned Zarith type

**Package `micheline`**

- Refactored transaction parameter encoding

**Package `rpc`**

- New: POST requests to forge, simulate and broadcast operations
- New: calls accept interface type `BlockID` which can be
    - `BlockAlias` (genesis or head)
    - `BlockLevel` an int64
    - `tezos.BlockHash` for named blocks
    - `BlockOffset` for offsets from a BlockID
- New: `MempoolMonitor` to monitor new mempool transactions
- Refactored `Mempool` type to return the same Operation type like block calls
- Refactored contract, rights and vote calls for consistent naming and parameters
- Refactored operations
  - renamed `OpKind()` into `Kind()`
  - renamed `RevelationOp` into `Reveal` and removed `..Op` suffix from all types
  - renamed `Origination.Manager()` into `Origination.ManagerAddress()`
  - unified operation metadata and results
  - added helpers to extract metadata, result and costs from typed interface

**Package `encoding`**

- New package for operation construction and serialization

**Package `wallet`**

- New package for account and operation management
- New types `Result`, `FutureResult ` and `Cost` to work with forge/simulate/broadcast results
- New `Monitor` to observe transaction completion

### Breaking changes

- RPC functions now use `BlockID` (BlockLevel, BlockHash, BlockOffset) to reference a block, all related functions that used to take a block height or hash now take a BlockID

## v0.11

Micheline
- renamed `SearchEntrypointName()` to `ResolveEntrypointPath()`
- support on-chain views in script
- support global constant detection and injection into script
- support timelock detection and types
- add new type `View` and related helper functions
