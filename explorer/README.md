# AnchorChain Explorer

Minimal read-only explorer for the existing AnchorChain HTTP API.

## What it can do

- configure an API base URL
- show node health and returned height values
- search by chain ID
- search by entry hash
- inspect chain summaries and recent entries
- inspect entry details and receipt availability
- copy chain IDs and entry hashes from detail pages

## What it cannot do

- create chains or write entries
- change protocol or daemon behavior
- bypass API-side authentication requirements

The explorer stays read-only and only uses the existing HTTP API.

## Prerequisites

- Node.js 20+
- npm
- a running AnchorChain API, such as local devnet at `http://127.0.0.1:8081`

## Install

From the repo root:

```bash
cd explorer
npm install
```

## Run locally

```bash
npm run dev
```

Then open:

```text
http://localhost:3000
```

## Test against local devnet

In one terminal from the repo root:

```bash
make build
./bin/anchorchaind devnet
```

In a second terminal:

```bash
cd explorer
npm install
npm run dev
```

Open `http://localhost:3000`, keep the API base URL set to `http://127.0.0.1:8081`, then:

1. confirm the health card shows the API status and height values
2. create test data with `anchor-cli` from a separate terminal if needed
3. open a chain ID and inspect the recent entries list
4. open an entry hash and inspect the payload and receipt status

Example write flow from the repo root:

```bash
./bin/anchor-cli chain create --extid demo --schema json --payload '{"hello":"anchorchain"}'
./bin/anchor-cli entry write --chain <CHAIN_ID> --schema json --payload '{"step":2}'
./bin/anchor-cli receipt verify --entry <ENTRY_HASH>
```

## Production note

This explorer does not add token-entry UI. If your API requires a token, the explorer would need to be extended before it can browse that endpoint.
