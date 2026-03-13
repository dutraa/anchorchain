# AnchorChain

> AnchorChain v0.1.0 (Developer Preview)

AnchorChain is a developer-friendly interface for the **Factom protocol**, designed to make it easy to create chains, write entries, and verify cryptographic receipts through a small HTTP API and CLI.

Instead of interacting directly with the lower-level Factom RPC interface, AnchorChain provides:

- a simple HTTP API
- a CLI for quick interaction
- structured payload support (JSON schemas)
- receipt verification
- a local devnet environment for experimentation

AnchorChain is intended for developers building systems that require **tamper-evident logs, proofs, or immutable records**.

Example use cases include:

- document proof and verification
- dataset provenance
- audit logs
- compliance records
- AI training dataset integrity
- timestamped research records

---

## Prerequisites

- Go 1.22+
- `make`

---

## Build

```bash
make build
```

This creates:

- `./bin/anchorchaind`
- `./bin/anchor-cli`

---

## Run a Local Devnet

```bash
./bin/anchorchaind devnet
```

Devnet starts a single local node with:

- HTTP API at `http://127.0.0.1:8081`
- Legacy Factom RPC at `localhost:8088`
- A preloaded devnet Entry Credit key so writes work without `--ec-key`

Leave that process running in one terminal.

---

## First Commands

In a second terminal:

Check node health:

```bash
./bin/anchor-cli node health
```

Create a chain:

```bash
./bin/anchor-cli chain create   --extid demo   --schema json   --payload '{"hello":"anchorchain"}'
```

Copy the returned `Chain ID`, then write a second entry:

```bash
./bin/anchor-cli entry write   --chain <CHAIN_ID>   --schema json   --payload '{"step":2}'
```

---

## Inspect Data

Inspect the chain and entries:

```bash
./bin/anchor-cli chain inspect --chain <CHAIN_ID>
./bin/anchor-cli chain tail --chain <CHAIN_ID> --limit 10
./bin/anchor-cli entry show --entry <ENTRY_HASH>
./bin/anchor-cli receipt verify --entry <ENTRY_HASH>
```

You can pass `--json` to any `anchor-cli` command for raw API output.

---

## Project Components

AnchorChain includes several components intended for developers:

| Component | Description |
|-----------|-------------|
| `anchorchaind` | Node + HTTP API |
| `anchor-cli` | Command line interface |
| `explorer/` | Web UI for browsing chains and entries |
| `sdk/js` | JavaScript SDK |
| `examples/` | Example applications |
| `docs/` | Documentation |

---

## Docs

- Architecture: `docs/architecture.md`
- Demo: `docs/demo.md`
- Devnet guide: `docs/devnet.md`
- Devnet deployment: `docs/deploy-devnet.md`
- API reference: `docs/api.md`

---

## Troubleshooting

**Missing Go**

Run:

```bash
go version
```

If Go is not installed or is older than 1.22, install or update Go first.

---

**Missing `make`**

Install `make`, or run the build commands manually:

```bash
mkdir -p bin
go build -o bin/anchorchaind ./cmd/anchorchaind
go build -o bin/anchor-cli ./cmd/anchor-cli
```

---

**Build artifacts not committed**

`bin/` is generated locally by `make build`.  
If the binaries are missing, rebuild them locally.

---

## Attribution

This software is derived from the open-source Factom protocol and remains licensed under the MIT License.

```
Copyright (c) Factom Foundation 2017.
Licensed under the MIT License.
```

All protocol-level code under `node/` is preserved to maintain upstream compatibility.
