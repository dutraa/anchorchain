# AnchorChain

![License](https://img.shields.io/badge/license-MIT-blue)
![Status](https://img.shields.io/badge/status-developer%20preview-orange)
![Go](https://img.shields.io/badge/go-1.22%2B-blue)

> Immutable records. Simple developer tools. Built on Factom.

AnchorChain is a **developer-friendly interface for the Factom protocol** designed to make it easy to create chains, write entries, and verify cryptographic receipts using a clean HTTP API, CLI, and web explorer.

Instead of interacting directly with low-level Factom RPC calls, AnchorChain provides:

- a simple HTTP API
- a CLI for rapid interaction
- structured payload support (JSON schemas)
- cryptographic receipt verification
- a local devnet for experimentation
- a web explorer for browsing anchored data

AnchorChain is intended for developers building systems that require **tamper‑evident logs, proofs, or immutable records**.

Example use cases include:

- document proof & verification
- dataset provenance
- audit logs
- compliance records
- AI training dataset integrity
- timestamped research records
- supply chain records

---

# Quick Start

## Prerequisites

- Go 1.22+
- `make`

Check Go installation:

```
go version
```

---

# Build

```
make build
```

This creates:

```
./bin/anchorchaind
./bin/anchor-cli
```

---

# Run a Local Devnet

Start the local node:

```
./bin/anchorchaind devnet
```

Devnet launches:

| Service | Address |
|-------|--------|
| AnchorChain API | http://127.0.0.1:8081 |
| Factom RPC | localhost:8088 |

The devnet automatically loads a **funded Entry Credit key** so writes work without specifying `--ec-key`.

Leave this terminal running.

---

# First Commands

Open a second terminal.

Check node health:

```
./bin/anchor-cli node health
```

Create a chain:

```
./bin/anchor-cli chain create   --extid demo   --schema json   --payload '{"hello":"anchorchain"}'
```

Example output:

```
Chain ID   : <CHAIN_ID>
Entry Hash : <ENTRY_HASH>
```

Write another entry:

```
./bin/anchor-cli entry write   --chain <CHAIN_ID>   --schema json   --payload '{"step":2}'
```

---

# Inspect Data

```
./bin/anchor-cli chain inspect --chain <CHAIN_ID>

./bin/anchor-cli chain tail   --chain <CHAIN_ID>   --limit 10

./bin/anchor-cli entry show   --entry <ENTRY_HASH>

./bin/anchor-cli receipt verify   --entry <ENTRY_HASH>
```

Add `--json` to any command to return raw API output.

---

# Using the Web Explorer

AnchorChain includes a lightweight **web explorer** that lets you browse chains and entries visually.

Start the explorer from the project root:

```
cd explorer
npm install
npm run dev
```

The explorer will start at:

```
http://localhost:3000
```

From the explorer you can:

- search for chains
- view entries in a chain
- inspect entry payloads
- verify receipts
- explore anchored data visually

The explorer connects to your local devnet API at:

```
http://localhost:8081
```

---

# Project Components

| Component | Description |
|-----------|-------------|
| `anchorchaind` | Node + HTTP API |
| `anchor-cli` | Command line interface |
| `explorer/` | Web UI for browsing chains and entries |
| `sdk/js` | JavaScript SDK |
| `examples/` | Example applications |
| `docs/` | Project documentation |

---

# Documentation

- `docs/architecture.md` — system architecture
- `docs/demo.md` — project demo walkthrough
- `docs/devnet.md` — devnet details
- `docs/deploy-devnet.md` — deploying devnet nodes
- `docs/api.md` — HTTP API reference

---

# Contributing

We welcome contributions from the community.

Open issues are available for:

- Python SDK
- Docker devnet
- Public devnet deployment
- AI dataset provenance examples

If you'd like to contribute:

1. Fork the repo
2. Create a feature branch
3. Submit a pull request

---

# Troubleshooting

### Missing Go

```
go version
```

If Go is missing or older than 1.22, install the latest Go release.

---

### Missing `make`

Build manually:

```
mkdir -p bin

go build -o bin/anchorchaind ./cmd/anchorchaind
go build -o bin/anchor-cli ./cmd/anchor-cli
```

---

### Missing binaries

`bin/` artifacts are generated locally by `make build`.

Rebuild them if necessary.

---

# Attribution

AnchorChain builds on the open‑source **Factom Protocol** and retains the MIT License.

```
Copyright (c) Factom Foundation 2017
Licensed under the MIT License
```

All protocol‑level code under `node/` remains preserved to maintain upstream compatibility.
