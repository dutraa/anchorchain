# AnchorChain Quickstart Guide
**March 2026**

## Overview
This guide helps developers install and run AnchorChain locally, create chains,
append entries, and verify receipts.

## Prerequisites

Ensure the following are installed:

- Git
- Local terminal access
- AnchorChain repository cloned

## Step 1: Clone Repository

```bash
git clone https://github.com/anchorchain/anchorchain.git
cd anchorchain
```

## Step 2: Build Node

```bash
make build
```

## Step 3: Start Local Devnet

```bash
./bin/anchorchaind devnet
```

This launches a local node with API endpoints available for development.

## Step 4: Create Chain

```bash
curl -X POST http://localhost:8080/chains
```

## Step 5: Append Entry

```bash
curl -X POST http://localhost:8080/chains/{chainId}/entries
```

Example payload:

```json
{
 "data": "hello world"
}
```

## Step 6: Retrieve Receipt

```bash
curl http://localhost:8080/entries/{entryHash}/receipt
```

## Step 7: Inspect Chain

```bash
anchor-cli chain inspect {chainId}
```

## Next Steps

- Integrate SDKs
- Implement schema validation
- Deploy nodes in distributed environments
