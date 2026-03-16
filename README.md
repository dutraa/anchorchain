# AnchorChain

![License](https://img.shields.io/badge/license-MIT-blue)
![Status](https://img.shields.io/badge/status-developer%20preview-orange)
![Go](https://img.shields.io/badge/go-1.22%2B-blue)

## AnchorChain

**Immutable records. Simple developer tools. Built on Factom.**

AnchorChain is a developer-friendly interface for the Factom protocol designed to make it easy to:

- Create chains
- Write entries
- Verify cryptographic receipts
- Anchor application data permanently to a blockchain

Instead of interacting with low-level Factom RPC calls, AnchorChain provides a modern developer stack:

- HTTP API
- CLI tooling
- Structured payload support
- Cryptographic receipt verification
- Local devnet for testing
- Web explorer for browsing anchored data

AnchorChain makes **blockchain anchoring accessible to application developers**, auditors, and data integrity platforms.

---

# Why AnchorChain?

Factom is one of the most efficient protocols for anchoring data to blockchain systems, but its tooling can be difficult for new developers.

AnchorChain solves that problem by providing:

- a clean API
- simple CLI commands
- structured data schemas
- developer-first workflows
- easy local testing environments

The goal is simple:

**Make anchoring data to blockchain as easy as writing to an API.**

---

# Architecture

AnchorChain consists of several components:

```
Applications
     │
     ▼
AnchorChain API
     │
     ▼
Factom Network
     │
     ▼
Bitcoin Anchoring
```

Components:

| Component | Purpose |
|--------|--------|
| API Server | REST interface for writing and verifying entries |
| CLI | Developer tool for interacting with chains |
| Explorer | Web interface to browse anchored data |
| Devnet | Local Factom test network |
| Receipt Engine | Cryptographic proof verification |

---

# Quick Start

## Clone the Repository

```bash
git clone https://github.com/anchorchain/anchorchain.git
cd anchorchain
```

---

## Install Dependencies

AnchorChain requires:

- Go 1.22+
- Docker (optional for devnet)

Install Go dependencies:

```bash
go mod download
```

---

# Running AnchorChain

## Start the API Server

```bash
go run cmd/server/main.go
```

Server will start on:

```
http://localhost:8080
```

---

# Using the CLI

AnchorChain includes a CLI tool for interacting with the network.

Build the CLI:

```bash
go build -o anchorchain cmd/cli/main.go
```

Example usage:

### Create a Chain

```bash
./anchorchain chain create
```

### Write an Entry

```bash
./anchorchain entry write \
  --chain <CHAIN_ID> \
  --data '{"message":"Hello AnchorChain"}'
```

### Get Chain Entries

```bash
./anchorchain chain entries <CHAIN_ID>
```

---

# Running a Local Devnet

AnchorChain includes a **local Factom devnet** for testing.

Start the devnet:

```bash
docker compose up
```

This launches:

- factomd
- factom-walletd
- AnchorChain API

Once running, you can test anchoring locally without interacting with the public network.

---

# Docker Deployment

Build the container:

```bash
docker build -t anchorchain .
```

Run it:

```bash
docker run -p 8080:8080 anchorchain
```

---

# API Overview

AnchorChain exposes a simple REST API.

### Create Chain

```
POST /chains
```

Response:

```
{
  "chain_id": "..."
}
```

---

### Write Entry

```
POST /chains/{chain_id}/entries
```

Payload:

```
{
  "data": {...}
}
```

---

### Verify Receipt

```
GET /receipts/{entry_hash}
```

Returns cryptographic proof that the data was anchored.

---

# Web Explorer

AnchorChain includes a lightweight explorer UI that allows users to:

- browse chains
- inspect entries
- verify receipts
- visualize anchored data

Explorer runs at:

```
http://localhost:3000
```

---

# Project Structure

```
anchorchain/
│
├── cmd/
│   ├── server/
│   └── cli/
│
├── internal/
│   ├── api/
│   ├── chains/
│   ├── receipts/
│   └── factom/
│
├── explorer/
│
├── devnet/
│
├── docker/
│
└── README.md
```

---

# Example Use Cases

AnchorChain can be used for:

- document timestamping
- supply chain proofs
- audit trails
- compliance logging
- data integrity verification
- legal evidence anchoring

Industries that benefit from immutable proofs include:

- finance
- healthcare
- government
- legal
- digital identity
- research

---

# Roadmap

### v0.1

- basic API
- CLI tooling
- local devnet
- receipt verification

### v0.2

- schema validation
- batch anchoring
- improved explorer

### v0.3

- SDKs (Go / JS / Python)
- hosted node support
- authentication layer

### v1.0

- production release
- scaling optimizations
- enterprise integrations

---

# Contributing

Contributions are welcome.

You can help by:

- submitting pull requests
- opening issues
- suggesting features
- improving documentation

---

# License

MIT License

---

# Learn More

Factom Protocol  
https://www.factomprotocol.org

---

# AnchorChain

**Immutable records for the modern internet.**
