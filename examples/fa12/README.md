## Decode FA 1.2 Transfers

Use TzGo to extract FA1.2 token transfer parameters from an operation receipt. 

### Usage

```sh
Usage: fa12 [args] <block> <pos>

  Decodes FA1.2 transfer info from transaction.

Arguments
  -node string
      Tezos node URL (default "https://rpc.tzpro.io")
  -v  be verbose
```

### Example

```sh
go run . BKuRFhvhsc3Bwdxedu6t25RmrkqpERVqEk867GAQu43muvi7j4d 17
```

### Implementation details

There are different ways to accomplish this for your own contracts. Here we just use the FA1.2 spec to illustrate how the proces works.

```go
  // you need the contract's script for type info
  script, err := c.GetContractScript(ctx, tx.Destination)
  if err != nil {
    return err
  }

  // unwind params for nested entrypoints
  ep, prim, err := tx.Parameters.MapEntrypoint(script.ParamType())
  if err != nil {
    return err
  }

  // convert Micheline params into human-readable form
  val := micheline.NewValue(ep.Type(), prim)

  // use Value interface to access data, you have multiple options
  // 1/ get a decoded `map[string]interface{}`
  m, err := val.Map()
  if err != nil {
    return err
  }

  buf, err := json.MarshalIndent(m, "", "  ")
  if err != nil {
    return err
  }
  fmt.Println("Map=", string(buf))
  fmt.Printf("Value=%s %[1]T\n", m.(map[string]interface{})["transfer"].(map[string]interface{})["value"])

  // 2/ access individual fields (ok is true when the field exists and
  //    has the correct type)
  from, ok := val.GetAddress("transfer.from")
  if !ok {
    return fmt.Errorf("No from param")
  }
  fmt.Println("Sent from", from)

  // 3/ unmarshal the decoded Micheline parameters into a Go struct
  type FA12Transfer struct {
    From  tezos.Address `json:"from"`
    To    tezos.Address `json:"to"`
    Value tezos.Z       `json:"value"`
  }
  type FA12TransferWrapper struct {
    Transfer FA12Transfer `json:"transfer"`
  }

  var transfer FA12TransferWrapper
  err = val.Unmarshal(&transfer)
  if err != nil {
    return err
  }
  buf, _ = json.MarshalIndent(transfer, "", "  ")
  fmt.Printf("FA transfer %s\n", string(buf))
```