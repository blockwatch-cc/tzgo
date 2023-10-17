## Transfer tez with ease

Use TzGo to construct and broadcast valid Tezos transactions. There are many convenience wrappers that help with the construction side and a simple one-stop-shop `Send()` call that does all the heavy lifting of fee estimation, coordinating signing, broadcasting and waitinh for confirmations. Checkout the source code in the `codec` package and `rpc/run.go` to get a sense of what's possible.

The most simple chain of function calls to send a single transfer with default options is

```go
rcpt, err := c.Send(
  ctx,
  codec.NewOp().
    WithSource(sender).
    WithTransfer(receiver, amount)
  nil,
)
```

### Usage

```sh
Usage: transfer [flags] <cmd> [sub-args]

Flags
  -key string
      private key
  -node string
      Tezos node URL (default "https://rpc.tzpro.io")
  -v  be verbose

Transaction Commands
  transfer   {<receiver> <amount>}+  transfer tez to single or multiple receiver(s)
```

If you provide multiple `<receiver> <amount>` pairs to the command, it produces a single batch transaction.