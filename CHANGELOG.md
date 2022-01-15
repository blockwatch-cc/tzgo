# Changelog

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
