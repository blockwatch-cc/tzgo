## Work with smart contract and tokens

Use TzGo to work with smart contracts and tokens. This examples shows you how to

- get contract entrypoints
- execute on-chain views to read data
- fetch contract and token metadata
- read FA token info
    - `balance_of`
    - `getBalance`
    - `getTotalSupply`
    - `getAllowance`
- send token transactions (private key required)
    - `transfer`
    - `approve`
    - `revoke`
    - `addOperator`
    - `removeOperator`

### Usage

```sh
Usage: contract [flags] <cmd> [sub-args]

Flags
  -node string
        Tezos node URL (default "https://rpc.tzpro.io")
  -v    be verbose

Query Commands
  run_view       <contract> <name> <data>     run view entrypoint `name` with JSON-encoded micheline input `data`
  info           <contract>                   load contract, print entrypoints and views
  metadata       <contract>                   fetch contract metadata
  token_metadata <contract> <token_id>        fetch token metadata
  balance_of     <contract> <owner>           FA2: fetch token balance for owner
  getBalance     <contract> <owner>           FA1: fetch token balance for owner
  getTotalSupply <contract>                   FA1: fetch total token supply
  getAllowance   <contract> <owner> <spender> FA1: fetch spender permit

Transaction Commands (require private key
  transfer       <contract> <token_id> <amount> <receiver> <privkey> FA1+2: transfer tokens to receiver
  approve        <contract> <spender> <amount> <privkey>     FA1: grant spending right
  revoke         <contract> <spender> <amount> <privkey>     FA1: revoke spending right
  addOperator    <contract> <token_id> <spender> <privkey>   FA2: grant full operator permissions
  removeOperator <contract> <token_id> <spender> <privkey>   FA2: revoke full operator permissions
  ```