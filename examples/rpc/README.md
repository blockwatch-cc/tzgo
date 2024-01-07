## Pull data from the Tezos RPC

Use TzGo to flexibly fetch all kinds of data from the Tezos RPC. A few basic and advanced examples are shown here.

### Usage

```sh
Usage: rpc [args] <cmd> [sub-args]

Commands
  block <hash>|head        show block info
  op <hash>:<list>:<pos>   show operation info
  contract <hash>          show contract info
  search <ops> <lvl>       output blocks containing operations in list
  bootstrap                wait until node is bootstrapped
  monitor                  wait and show new heads as they are baked

Global arguments
  -d  Enable debug mode
  -node string
      Tezos node url (default "https://rpc.tzpro.io")
  -v  Be verbose
```

