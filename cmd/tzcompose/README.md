# TzCompose - Tezos Automation Framework

TzCompose is a tool for defining and running complex transaction sequences on Tezos. With TzCompose, you use a YAML file to configure pipelines from different kinds of tasks and then run a single command to execute these pipelines.

TzCompose works on all Tezos networks and makes it easy to clone contracts and transaction sequences between them.

Developers can use TzCompose for

- smart contract deployment and maintenance
- traffic generation for protocol and application testing
- automating test-case setups
- cloning contract deployments and setup logic between networks

## Using TzCompose

```sh
go run blockwatch.cc/tzgo/cmd/tzcompose [cmd] [flags]

Env
  TZCOMPOSE_BASE_KEY  private key for base account
  TZCOMPOSE_API_KEY   API key for RPC and index calls (optional)

Flags
  -f file
      configuration file or path (default "tzcompose.yaml")
  -file file
      configuration file or path (default "tzcompose.yaml")
  -resume
      continue pipeline execution
  -rpc string
      Tezos node RPC url (default "https://rpc.tzpro.io")
  -h  print help and exit
  -v  be verbose (default true)
  -vv
      debug mode
  -vvv
      trace mode
```

TzCompose can execute configurations from a single file `-f file.yaml`, a single directory `-f ./examples/fa` or all subdirectories `-f ./examples/...` in which cases all yaml files will be read in filesystem order.

### Available Commands

- `clone`: clone transactions starting from the origination of a contract
- `validate`: validate compose file syntax and parameters
- `simulate`: simulate compose file execution against a blockchain node
- `run`: execute compose file(s) sending signed transactions to a blockchain node
- `version`: print version and exit

TzCompose relies on the Tezos Node RPC and (for clone) on the TzIndex API. Both are publicly available via https://tzpro.io with a free subscription. Export your API key as

```sh
export TZCOMPOSE_API_KEY=<your-api-key>
```

### Available Tasks

- [batch](#batch) - send multiple transactions as single operation
- [call](#call) - send smart contract call
- [delegate](#delegate) - delegate to baker
- [deploy](#deploy) - create smart contract
- [double_endorse](#double-endorse) - force a double endorsement slash
- [double_bake](#double-bake) - force a double bake slash
- [register_baker](#register-baker) - register as baker
- [token_approve](#token-approve) - approve token spender
- [token_revoke](#token-revoke) - revoke token spender
- [token_transfer](#token-transfer) - send token transfer(s)
- [transfer](#transfer) - send tez transfer(s)
- [undelegate](#undelegate) - remove delegation from baker
- [wait](#wait) - wait for condition
- [stake](#stake) - Oxford stake
- [unstake](#unstake) - Oxford unstake
- [finalize_unstake](#finalize_stake) - Oxford finalize unstake

### Configuration

TzCompose YAML files contain different sections to define accounts, variables and pipelines. Pipelines can be composed from different `tasks` which will generate transactions when executed. A compose file may contain multiple pipelines and each pipeline virtually unlimited tasks.

```yaml
# Available engines: alpha
version: <engine-version>

# Specify alias names for tz1 user accounts and optional key id
accounts:
  - name: alias
    id: 1

# Specify contract addresses or other data as variables and later
# reference them inside pipelines
variables:
  - <name>: KT1...

# Specify transaction sequences as pipelines. Each sequence consists of tasks
# of a specifed type and corresponding input parameters
pipelines:
  <name>:
    - task: deploy | call | transfer | register_baker | token_transfer | ...
      source: $var | address
      destination: $var | address

```

## How it works

TzCompose reads config files and runs pipeline tasks in order of appearance. Each task emits a transaction that is signed and broadcast to the network. If successful, the next task is processed. Task configurations may contain static content, reference variables and import data files which makes it easy to orchestrate complex scenarios.

TzCompose automates wallet/key management and guarantees deterministic account addresses. That way, running the same pipeline on a different network will lead to the same result. (contract addresses are one unfortunate exception).

TzCompose also stores pipeline state and can resume execution from where it stopped.

### Wallets

TzCompose requires a single funded `base` account to sign and send transactions. Configure the base account by exporting its private key as `TZCOMPOSE_BASE_KEY` environment variable.

On Flextes sandbox extract the private key with

```sh
export TZCOMPOSE_BASE_KEY=`docker exec tezos_sandbox flextesa key-of-name alice | cut -f4 -d, | cut -f2 -d:`
```

All other wallet keys are deterministically derived from this `base` key using BIP32. Child keys are identified by their numeric id. You can assign alias names to them in the `accounts` section of a compose file. All child accounts use Ed25519 keys (tz1 addresses).

> TzCompose does not allow you to specify wallet keys in configuration files. This is a deliberate design choice to prevent accidental leakage of key material into code repositories.

Keep in mind that when you reuse the same child ids in different compose files, these accounts may have state and history from executing other compose files. Usually this is not a problem, but it may be in certain test scenarios when the account is already a baker or is expected to be empty.

### Variables

TzCompose lets you define variables for common strings and addresses. You can use variables as source, destination and in task arguments. TzCompose defines a few default variables

* `$base` - base account address
* `$now` - current wall clock time in UTC (you can also add durations like `$now+5m`)
* `$zero` - tz1 zero address (binary all zeros)
* `$burn` - tz1 burn address

### Cloning Contracts

In some cases it is desirable to clone existing contracts and their setup procedure across networks. TzCompose implements a `clone` command which downloads contract code, initial storage and transaction history from an indexer and writes a deployable pipeline into a fresh compose config file.

TzCompose offers different clone modes which define how and where script, storage and params are stored:

* `file` stores Micheline data as JSON data in separate files and references them inside the config file
* `url` stores a Tezos RPC URL in inside the config file and fetches data on demand
* `bin` embeds Micheline data as hex string into the config file
* `json` embeds Micheline data as JSON string into the config file

In either case `clone` attempts to decode storage and call parameters into human readable args. If this fails because args are too complex or a construct is unsupported the fallback is to use pre-encoded Micheline data.

A single clone call is often enough to get a working copy of a contract. Some contracts may require admin permissions or other changes to initial storage. Then it is necessary to overwrite some arguments. With args available just replace any fixed admin address with `$base` (or `$var` for child accounts) and add/remove arg contents directly in the compose file as necessary.

Without args you can [patch](#patching-micheline) Micheline encoded data to edit the contents of specific nodes.


**Clone Example**

Here we clone the Tether contract and 2 transactions following its origination which generates a `tzcompose.yaml` file along with several files containing contract source and call paramaters.

```sh
# clone Tether FA2 contract from mainnet
tzcompose clone -contract KT1XnTn74bUtxHfDtBmm2bGZAQfhPbvKWR8o -name tether -n 2
```

Then open `tzcompose.yaml` and edit its admin accounts. For brevity we show relevant lines only, the original file is longer:
```yaml
pipelines:
  tether:
    - task: deploy
      script:
        storage:
          args:
            administrators:
              $base: "1"     # <- replace with $base variable here
    - task: call
      params:
        entrypoint: mint
        args:
          - owner: $base     # <- replace with $base variable here
```


### Micheline Arguments

Initial storage and contract call parameters must be Micheline encoded, but it's a pain to read and write this format. TzCompose allows you to either use pre-encoded files, binary blobs or YAML key/value arguments.

When using arguments keep in mind that all fields are required, even if their value is empty or default. If args are too complex and you have to work with pre-encoded data you can still use [patch](#patching-micheline) for replacing contents as described below.

Both YAML args and patch use variable replacement for addresses, strings and timestamps.

```yaml
args:
  # Keys must always be strings
  key: ...  # Regular keys don't require quites
  "0": ...  # Use quotes when the key is numeric
  "": "..." # Empty string (for metadata)

  # Integers up to int64 can be written without quotes, bigints require quotes
  val: 1
  val: "1"

  # Strings can be written without quotes as long as they don't contain a reserved
  # yaml character, yaml multiline strings are supported as well
  val: some_string
  val: some string
  val: "some string"
  val: >
    a string
    that spans
    multiple lines

  # Sets and lists must be written as YAML arrays
  my_set: null # empty
  my_set: []   # empty
  my_set:
    - val1
    - val2

  # Maps and initial bigmap storage must be written as YAML map
  my_map: null # empty
  my_map: {}   # empty
  my_map:
    key_1: val_1 # map keys must be unique
    key_2: val_2
    "tz1...,1": val # write Micheline pair keys as comma separated strings

  # Null (or Unit) as well as empty optionals (None) are written as null
  option: null

  # Union types must contain the nested name of the chosen branch
  # i.e. FA2 update_operators calls would look like
  - add_operator:
    owner: $account-1
    operator: $account-2
    token_id: 0

  # For contracts without named storage or paramaeters, use the numeric position
  "0": val # first arg
  "1": val # second arg
```

### Specifying Data Sources

TzCompose can read Micheline data from different sources. Compilers typically store code and storage in JSON files, but when cloning we may want to embed everything into a the config file to make it self-sufficient or even pull data from a URL when running the pipeline.

Below is a list of available options which work for scripts and call params:

```yaml
- task: deploy
  script:
    code:
      # Read from a file relative to the config file's directory
      file:
      # Read from a remote URL
      url:
      # Use the embedded value (either JSON string or hex string)
      value:
    storage:
      # Read from a file relative to the config file's directory
      file:
      # Read from a remote URL
      url:
      # Use the embedded value (either JSON string or hex string)
      value:
      # Replace specific contents read from file, url or value
      patch:
      # Use human readable arguments
      args:

- task: call
  params:
    # Entrypoint name is always required
    entrypoint: name
    # Read from a file relative to the config file's directory
    file:
    # Read from a remote URL
    url:
    # Use the embedded value (either JSON string or hex string)
    value:
    # Replace specific contents read from file, url or value
    patch:
    # Use human readable arguments
    args:
```

### Referencing Files

To increase flexibility when working with different Michelson compilers TzCompose provides a few ways to load data from files:

```yaml
- task: deploy
  script:
    code:
      # When the file contains only the code segment
      file: code.json
      # When code is nested like with Ligo we can append a json-path
      file: code.json#michelson
    storage:
      # Read the entire file
      file: storage.json
      # Extract with json-path from file
      file: script.json#storage
      # Inside args reference files as `@filename#json-path`
      # (may need quotes since @ is reserved for future use in yaml)
      args:
        func: "@lambdas.json#0"
```

### Patching Micheline

In case you only have storage or parameters in files or binary blobs you may want to patch certain fields inside. To do this add one or more `patch` sections:

```yaml
- task: deploy
  script:
    storage:
      # Specifies the source where to read the original full storage primitive tree from
      file: storage.json
      # Patch takes a list of instructions and executes them in order
      patch:
      -
        # Key defines a (nested) label from the Micheline type tree where to
        # patch contents in the value tree. this works in many cases, but if
        # it doesn't (because type is too complex like nested maps in maps etc)
        # you can use path below
        key: manager
        # Path is an alternative to key. it specifies the exact path in the
        # Micheline value tree to replace. Path segments are always numeric.
        path: 0/0/1
        # Type defines the primitive type to replace. It is used to correctly
        # parse and type-check the value and to produce the correct Micheline
        # primitive.
        type: address
        # Value contains the contents we want to replace with. It is typically
        # a string, int or variable.
        value: $base
        # Optimized defines if the output prim is in optimized form or not.
        # This applies to addresses, timestamps, keys and signatures.
        optimized: false
```

### Error Handling

Per default, tzcompose will stop executing all pipelines when the first error occurs. This may be a configuration, network communication or transaction state error. Sometimes it is useful to skip non-critical expected errors such as re-activating an already active baker. Since tzcompose does not know what is critical to your pipeline and what not, you can use the `on-error` argument to specify how an error is handled.

```yaml
- task: transfer
  on-error: ignore # allowed values: ignore, warn, fail
```

## Task Reference

Tasks define a single action which typically emits an on-chain transaction. The transaction will be signed by the account specified in the `source` field. If source is empty it defaults to `$base`. Most tasks also define a `destination` which is either target of a call or receiver of a transfer or other activity.

Continue reading for a list of available tasks in alphabetic order:

### Batch

Batch tasks combine multiple actions into a single signed transaction. They are useful to save time and gas, but also to atomically perform events that would otherwise be less secure (like granting and revoking token allowances). Almost all other tasks can be nested inside a batch.

```yaml
# Spec
task: batch
source: $var
contents:
- task: transfer
  destination: $var | string
  amount: number
- task: transfer
  destination: $var | string
  amount: number
```

### Call

Call is a generic way to send smart contract calls. Calls require a destination contract, an entrypoint and call arguments. Entrypoint and argument types must match the contract interface. TzCompose checks before signing a transaction whether the entrypoint exists and arguments are valid.

```yaml
# Spec
task: call
source: $var
destination: $var | string
params:
  entrypoint: name
  args:
  file:
  url:
  value:
```

### Delegate

Delegate delegates `source`'s full spendable balance to the selected `destination` baker. On Tezos this balance remains liquid.

```yaml
# Spec
task: delegate
source: $var
destination: $var | string
```

### Deploy

Deploy creates a new smart contract from `code` and `storage` and on success makes the contract's address available as variable (use `alias` to define the variable name).

```yaml
# Spec
task: deploy
# required, alias defines the variable name under which the new contract
# address is made available to subsequent tasks and pipelines
alias: string
# optional tez amount to send to the new contract
amount: number
# specify contract script
script:
  # you may use a single source for both code and storage if available
  # only one of url, file, value is permitted
  url:
  file:
  value:
  # or you can specify independent code and storage sections
  # only one of url, file, value is permitted
  code:
    file:
    url:
    value:
  # add an independent storage section when you use args or patch or
  # when you need to load data from a different source than code
  # only one of args, url, file, value is permitted
  # use patch for replacing content loaded via url, file or value
  storage:
    args:
    file:
    url:
    value:
    patch:
```

### Double Endorse

Produces a fake double-endorsement which slashes the baker in `destination` and awards denunciation rewards to the baker who includes this operation. The destination must be registered as baker and a private key must be available for signing. The slashed baker must have endorsements rights and this task waits until a block with such rights is baked.

To successfully execute this task you need to sufficiently fund and register the baker and then wait a few cycles for rights to activate. Note that on sandboxes the $alice key is not the sandbox baker. To lookup the actual baker key, docker exec into the sandbox and search for the `secret_keys` file in the Tezos client dir.

```yaml
# Spec
task: double_endorse
destination: $var # <- this baker is slashed
```

### Double Bake

Produces a fake double-bake which slashes the baker in `destination` and awards denunciation rewards to the baker who includes this operation. The destination must be registered as baker and a private key must be available for signing. The slashed baker must have at least one round zero baking rights. The task waits until a block with such right is baked and then sends a double baking evidence with two fake (random) payload hashes.

To successfully execute this task you need to sufficiently fund and register the baker and then wait a few cycles for rights to activate. Note that on sandboxes the $alice key is not the sandbox baker. To lookup the actual baker key, docker exec into the sandbox and search for the `secret_keys` file in the Tezos client dir.

```yaml
# Spec
task: double_bake
destination: $var # <- this baker is slashed
```

### Register Baker

Registers the source account as baker on the network by sending a self-delegation. Registration is idempotent. In case a baker becomes inactive it can be easily reactivated by registering it again.

```yaml
# Spec
task: register_baker
source: $var
```

### Token Approve

Approve specifies that a `spender` account is allowed to transfer a token amount owned by `source` in the token ledger contract `destination`. FA1.2 supports max amount, but no token id. FA2 has token ids, but you can only allow to spend everything or nothing.

Approve is a convenience task that hides details of different standards. For full control use a `call` task.

```yaml
# Spec
task: token_approve
source: $var
destination: $var | string # token ledger
args:
  standard: fa12 | fa2
  token_id: number # optional, fa2 only
  spender: $var | string
  amount: number # required for fa12, ignored for fa2
```

### Token Revoke

Revoke is the opposite of approve and withdraws a `spender`'s right to transfer tokens in ledger `destination`. Revoke is a convenience task that hides details of different standards. For full control use a `call` task.


```yaml
# Spec
task: token_revoke
source: $var
destination: $var | string # token ledger
args:
  standard: fa12 | fa2
  token_id: number # optional, fa2 only
  spender: $var | string
```

### Token Transfer

Transfers an `amount` of tokens with `token_id` owned by account `from` in ledger `destination` to another account `to`. `source` must either be the same as `from` or must be approved as spender.

```yaml
# Spec
task: token_transfer
source: $var
destination: $var | string # token ledger
args:
  standard: fa12 | fa2
  token_id: number # optional, fa2 only
  from: $var | string
  to: $var | string
  amount: number
  # fa2 only: specify multiple receivers and multiple token ids
  # this is mutually exclusive with token_id, to and amount above
  receivers:
    token_id: number
    to: $var | string
    amount: number

```

### Transfer

Transfers an `amount` of tez from `source` to `destination`.

```yaml
# Spec
task: transfer
source: $var
destination: $var | string
amount: number
```

### Undelegate

Undelegate removes the `source`'s delegation from its current baker.

```yaml
# Spec
task: undelegate
source: $var
```

### Wait

Wait stops pipeline execution for a defined amount of blocks, time or cycles. It is useful to wait for protocol events (at cycle and) or allow for some traffic getting generated by a
second tzcompose instance. Wait is not able to syncronize between pipelines though.

```yaml
# Spec
task: wait
for: cycle | block | time
value: N | +N
```

```yaml
# wait for next cycle
type: wait
for: cycle
value: +1

# wait 10 blocks
type: wait
for: block
value: +10

# wait until tomorrow
type: wait
for: time
value: +24h
```

```yaml
pipelines:
  transfer-and-wait:
  - task: transfer
    destination: $account-1
    amount: 10_000_000
  - task: wait
    for: block
    value: +10
  - task: transfer
    destination: $account-2
    amount: 10_000_000
```

### Stake

Lets the `source` stake `amount` with its current baker. Source must be delegated and staking must be enabled in the protocol for this to succeed. This operation takes no destination because the operation is sent to source.

```yaml
# Spec
task: stake
source: $var
amount: number
```

### Unstake

Lets the `source` request unstake of `amount` from the current baker. Source must be staking for this to succeed. This operation takes no destination because the operation is sent to source.

```yaml
# Spec
task: unstake
source: $var
amount: number
```

### Finalize Unstake

Lets `source` request to pay out unstaked unfrozen funds back to spendable balance. This operation takes no amount and no destination. All available funds are sent.

```yaml
# Spec
task: finalize_unstake
source: $var
```
