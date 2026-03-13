# AnchorChain Devnet Guide

AnchorChain devnet is a single-node local development network exposed by `anchorchaind devnet`.

Use it for development and testing only. The daemon prints `DEVNET MODE - NON-PRODUCTION ONLY`, and this environment should be treated as disposable. Devnet state may be reset, so do not rely on it for durable data.

## Start devnet

Build the binaries first:

```bash
make build
```

Start the local devnet:

```bash
./bin/anchorchaind devnet
```

By default this starts:

- HTTP API on `http://127.0.0.1:8081`
- Legacy Factom RPC on `localhost:8088`
- A devnet Entry Credit key inside the daemon, so writes can work without passing `--ec-key`

Leave that process running while you use the CLI.

## Connect `anchor-cli`

For the local devnet, `anchor-cli` already defaults to:

```bash
./bin/anchor-cli --api http://127.0.0.1:8081 node health
```

You can also rely on the default with no `--api` flag:

```bash
./bin/anchor-cli node health
```

To target another API endpoint, pass `--api` or set `ANCHORCHAIN_API`:

```bash
./bin/anchor-cli --api http://203.0.113.10:8081 node health
```

```bash
ANCHORCHAIN_API=http://203.0.113.10:8081 ./bin/anchor-cli node health
```

If the API was started with a token, the CLI can send it with `--token` or `ANCHORCHAIN_API_TOKEN`:

```bash
./bin/anchor-cli --api http://203.0.113.10:8081 --token <TOKEN> node health
```

## Verify health

Check that the devnet API is reachable:

```bash
./bin/anchor-cli node health
```

You can also call the HTTP API directly:

```bash
curl -s http://127.0.0.1:8081/health
```

## Basic write flow

Create a chain:

```bash
./bin/anchor-cli chain create \
  --extid demo \
  --schema json \
  --payload '{"hello":"anchorchain"}'
```

Copy the returned `Chain ID`, then append an entry:

```bash
./bin/anchor-cli entry write \
  --chain <CHAIN_ID> \
  --schema json \
  --payload '{"step":2}'
```

Copy the returned `Entry Hash`, then verify its receipt:

```bash
./bin/anchor-cli receipt verify --entry <ENTRY_HASH>
```

Helpful follow-up commands:

```bash
./bin/anchor-cli chain inspect --chain <CHAIN_ID>
./bin/anchor-cli chain tail --chain <CHAIN_ID> --limit 10
./bin/anchor-cli entry show --entry <ENTRY_HASH>
```

Pass `--json` to any `anchor-cli` command if you want raw API output.
