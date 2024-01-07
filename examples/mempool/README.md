## Observe the Tezos Mempool

Use TzGo to watch the mempool for new operations.

### Usage

```sh
Usage: mempool [args] <cmd> [sub-args]

Commands
  stream [<filter>]   stream new ops entering the mempool, optional filter
  wait <ophash>       wait for operation to be visible in mempool
  info                print info about mempool

Arguments
  -node string
      Tezos node URL (default "https://rpc.tzpro.io")
  -ttl int
      Operation TTL (default 120)
  -v  be verbose
```
