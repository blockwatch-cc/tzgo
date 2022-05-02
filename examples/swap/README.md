## Interact with Quipuswap

Use TzGo to work with Quipuswap AMM pools. Right now this example is read-only and we may extend it to run complete swaps in the future.

### Usage

```sh
Usage: swap [flags] <cmd> [sub-args]

Flags
  -node string
      Tezos node URL (default "https://rpc.tzstats.com")
  -v  be verbose

Commands
  info  <KT1> <pool_id>             show Quipuswap pool info
  sim   <KT1> <pool_id> <in>        dry run a swap on `in` tokens
  swap  <KT1> <pool_id> <in> <pk>   execute swap of `in` tokens
```

### Examples

```sh
# get info from tez/ctez pool on Quipuswap v1
go run . info KT1FbYwEWU8BTfrvNoL5xDEC5owsDxv9nqKT 0

# estimate costs of a swapping 1000 tez for ctez
go run . sim KT1FbYwEWU8BTfrvNoL5xDEC5owsDxv9nqKT 0 1000
```

