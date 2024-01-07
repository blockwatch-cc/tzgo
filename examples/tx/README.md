## Build and sign Tezos operations

Use TzGo to produce any type of Tezos operation. One by one, this example shows all the basic steps to make, encode, simulate, sign, broadacst and wait for operation to be confirmed .

### Usage

Operation content is expected as JSON string. Make sure you use proper shell quotation like `'{"source":"tz1...","fee":"1000",..}'`

```sh
Usage: tx [args] <cmd> [sub-args]

Arguments
  -node string
      Tezos node URL (default "https://rpc.tzpro.io")
  -v  be verbose

Commands
  encode <type> <data>       generate operation `type` from JSON `data`
  validate <type> <data>     compare local encoding against remote encoding
  decode <msg>               decode binary operation
  digest <msg>               generate operation digest for signing
  sign <key> <msg>           sign message digest
  sign-remoate <key> <msg>   sign message digest using remote signer
  simulate <msg>             simulate executing operation using invalid signature
  broadcast <msg> <sig>      broadcast signed operation
  wait <ophash> [<n>]        waits for operation to be included after n confirmations (optional)

Operation types & required JSON keys
  endorsement                    level:int slot:int round:int payload_hash:hash
  preendorsement                 level:int slot:int round:int payload_hash:hash
  double_baking_evidence         <complex>
  double_endorsement_evidence    <complex>
  double_preendorsement_evidence <complex>
  seed_nonce_revelation          level:str(int) nonce:hash
  activate_account               pkh:addr secret:hex32
  reveal                         source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) public_key:key
  transaction                    source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) amount:str(int) destination:addr
  origination                    source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) balance:str(int) delegate?:addr script:prim
  delegation                     source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) delegate?:addr
  proposals                      source:addr period:str(int) proposal:[hash]
  ballot                         source:addr period:str(int) proposal:hash ballot:(yay,nay,pass)
  register_global_constant       source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) value:prim
  set_deposits_limit             source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) limit:int
  failing_noop                   arbitrary:str
```

### Examples

We use a `reveal` operation as simple example, but others work with the same schema. Note that the binary encoding used as input to simulation and signing already contains a recent block hash. You can't jost copy paste below examples 1:1 to walk through the steps. Instead, start at the first command and use each command's output as the input to the following.

```sh
# encode to binary (also adds a recent block hash for TTL control)
go run ./examples/tx -v encode reveal '{"source":"tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S", "public_key":"edpkuQqN9HB3jY1FvDzt15WQDVSHR4vQGd1wv6iqJ73wkrKecRtnXh","fee": "1000", "counter": "2886593", "gas_limit": "1000", "storage_limit": "0"}'

# decode converts a binary encoded transaction back to Go
go run ./examples/tx -v decode "09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6b005c7886828ec2a24f1814484de7dd53e559831c3fe807c197b001e8070100654b5b22880736d33865b4f30367e90feb81b17cc0ceb7ac951a0066142d5847"

# validate, compares locally created result with a version created by a Tezos node
go run ./examples/tx -v validate reveal '{"source":"tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S", "public_key":"edpkuQqN9HB3jY1FvDzt15WQDVSHR4vQGd1wv6iqJ73wkrKecRtnXh","fee": "1000", "counter": "2886593", "gas_limit": "1000", "storage_limit": "0"}'

# simulate, dry-runs a transaction to measure its effects on gas and storage consumption
# you may use this to adjust gas/storage limit so that the operation will not fail.
go run ./examples/tx -v simulate "5f2a8d51254e06fcbd276228e84d8c9dbd9c8fc89a5986cd28b5d71b66ee57466b005c7886828ec2a24f1814484de7dd53e559831c3fe807c197b001e8070000654b5b22880736d33865b4f30367e90feb81b17cc0ceb7ac951a0066142d5847"

# sign produces a local signature from given private key
go run ./examples/tx sign "edskS3QdzK2YuceeLEaQrejebTy1fzy3VakDXdWMxaYanvu1F8WL6MetyYkJGVCAmSxFjgLt4ZjfwKcqETSNCkuPjrGhap24rS" "5f2a8d51254e06fcbd276228e84d8c9dbd9c8fc89a5986cd28b5d71b66ee57466b005c7886828ec2a24f1814484de7dd53e559831c3fe807c197b001e8070000654b5b22880736d33865b4f30367e90feb81b17cc0ceb7ac951a0066142d5847"

# broadcast takes encoded operation and signature and sends it for inclusion to a tezos node
go run ./examples/tx -v broadcast "5f2a8d51254e06fcbd276228e84d8c9dbd9c8fc89a5986cd28b5d71b66ee57466b005c7886828ec2a24f1814484de7dd53e559831c3fe807c197b001e8070000654b5b22880736d33865b4f30367e90feb81b17cc0ceb7ac951a0066142d5847" "edsigtrSrvNckjtxciM4iuoZjytrfQzmHyb7nv6ZLerrH5w19aMgmUkz1HDy3KMpF2S3jnZBW8ZTNqAtXYJ1ZKYtcj5NwcLmdfG"
```
```

