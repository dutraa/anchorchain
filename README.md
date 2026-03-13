# AnchorChain

> AnchorChain v0.1.0 (Developer Preview)

AnchorChain is a developer-friendly wrapper around the Factom protocol for creating chains, writing entries, and verifying receipts through a small HTTP API and CLI.

## Prerequisites

- Go 1.22+
- `make`

## Build

```bash
make build
```

This creates:

- `./bin/anchorchaind`
- `./bin/anchor-cli`

## Run a Local Devnet

```bash
./bin/anchorchaind devnet
```

Devnet starts a single local node with:

- HTTP API at `http://127.0.0.1:8081`
- Legacy Factom RPC at `localhost:8088`
- A preloaded devnet Entry Credit key, so writes work without `--ec-key`

Leave that process running in one terminal.

## First Commands

In a second terminal:

```bash
./bin/anchor-cli node health
```

```bash
./bin/anchor-cli chain create \
  --extid demo \
  --schema json \
  --payload '{"hello":"anchorchain"}'
```

Copy the returned `Chain ID`, then write a second entry:

```bash
./bin/anchor-cli entry write \
  --chain <CHAIN_ID> \
  --schema json \
  --payload '{"step":2}'
```

To inspect data:

```bash
./bin/anchor-cli chain inspect --chain <CHAIN_ID>
./bin/anchor-cli chain tail --chain <CHAIN_ID> --limit 10
./bin/anchor-cli entry show --entry <ENTRY_HASH>
./bin/anchor-cli receipt verify --entry <ENTRY_HASH>
```

Pass `--json` to any `anchor-cli` command for raw API output.

## Docs

- Architecture: [docs/architecture.md](/C:/Users/tonyd/OneDrive/Desktop/projects/anchorchain/docs/architecture.md)
- Demo: [docs/demo.md](/C:/Users/tonyd/OneDrive/Desktop/projects/anchorchain/docs/demo.md)
- Devnet guide: [docs/devnet.md](/C:/Users/tonyd/OneDrive/Desktop/projects/anchorchain/docs/devnet.md)
- Devnet deployment: [docs/deploy-devnet.md](/C:/Users/tonyd/OneDrive/Desktop/projects/anchorchain/docs/deploy-devnet.md)
- API reference: [docs/api.md](/C:/Users/tonyd/OneDrive/Desktop/projects/anchorchain/docs/api.md)

## Troubleshooting

- Missing Go: run `go version`. If Go is not installed or is older than 1.22, install or update Go first.
- Missing `make`: install `make`, or run:
  ```bash
  mkdir -p bin
  go build -o bin/anchorchaind ./cmd/anchorchaind
  go build -o bin/anchor-cli ./cmd/anchor-cli
  ```
- Build artifacts not committed: `bin/` is generated locally by `make build`; if the binaries are missing, rebuild them.

## Attribution

This software is derived from the open-source Factom Protocol and remains licensed under the MIT License.

```
Copyright (c) Factom Foundation 2017.
Licensed under the MIT License.
```

All protocol-level code under `node/` is preserved to maintain upstream compatibility.

