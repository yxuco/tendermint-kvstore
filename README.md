# tendermint-kvstore

[Tendermint](https://github.com/tendermint/tendermint) tutorial

## Get latest Tendermint source

```bash
mkdir -p ~/work/tendermint
cd ~/work/tendermint
git clone https://github.com/tendermint/tendermint.git
```

Build Tendermint and install executable in `$GOPATH/bin/tendermint`

```bash
cd tendermint
go mod tidy
make tools
make install
```

## Create Tendermint default configurations

Create default node config in a sample folder. Note that if `TMHOME` is not specified, the config will be created in `$HOME/.tendermint`.

```bash
mkdir -p ~/work/tendermint/node
cd ~/work/tendermint/node
TMHOME="." tendermint init
```

## Build sample kvstore app

Assume that this repo is cloned in `~/work/tendermint/kvstore`.

```bash
cd ~/work/tendermint/kvstore
# go mod init github.com/yxuco/tendermint-kvstore
go mod tidy
go build
```

## Start kvstore app

start the kvstore app, which will wait for tendermint node to connect.

```bash
cd ~/work/tendermint/node
rm example.sock
../kvstore/kvstore
```

## Start full Tendermint node

In another terminal, start a full tendermint node, and connect it to kvstore app.

```bash
cd ~/work/tendermint/node
TMHOME="." tendermint node --proxy_app=unix://example.sock
```

## Send test transaction and query

From another terminal, send a transaction

```bash
curl -s 'localhost:26657/broadcast_tx_commit?tx="tendermint=rocks"'
```

Send a query

```bash
curl -s 'localhost:26657/abci_query?data="tendermint"'
```

## Check base64 encoded key-values in query result

```bash
echo -n "dGVuZGVybWludA==" | base64 --decode
echo -n "cm9ja3M=" | base64 --decode
```

## Test kvstore app as built-in app

Shutdown Tendermint node and kvstore app from the previous testing. Restart the `kvstore` app as a `built-in` app.

```bash
../kvstore/kvstore -config "./config/config.toml" -built-in true
```

The built-in app itself is a full Tendermint node. From another terminal, send new transaction and query similar the previous tests.

## Send transaction and query using JSON-RPC

When sending transaction and queries using JSON-RPC, instead of OpenAPI in the previous steps, you need to encode the data carefully as required by the app.

Transaction parameters must use base64 encoding, e.g.,

```bash
tx=$(echo -n "tendermint=jsonrpc" | base64)
curl --header "Content-Type: application/json" --request POST --data '{"method": "broadcast_tx_commit", "params": {"tx": "'${tx}'"}, "id": 1}' localhost:26657
```

Query parameters must use HEX dump, e.g.,

```bash
data=$(echo -n "tendermint" | xxd -pu)
curl --header "Content-Type: application/json" --request POST --data '{"method": "abci_query", "params": {"data": "'${data}'"}, "id": 2}' localhost:26657
```
