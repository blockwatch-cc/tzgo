## Working wth baking rights

Use TzGo to fetch baking rights and related data from the Tezos RPC. This example covers a few basic cases.

### Usage

```sh
Usage: rights [args] <cmd> [sub-args]

Commands
  snap [<cycle>]              get snapshot info for all or selected cycle
  bake <cycle> [<address>]    get cycle baking rights for baker
  endorse <cycle> [<address>] get cycle endorsing rights for baker

Arguments
  -node string
      Tezos node URL (default "https://rpc.tzstats.com")
  -ttl int
      Operation TTL (default 120)
  -v  be verbose
```
