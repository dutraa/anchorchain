# AnchorChain JavaScript SDK

A minimal TypeScript client for the existing AnchorChain HTTP API.

## What it does

- checks node health
- creates chains
- writes entries
- fetches chains and entries
- verifies receipts

The SDK maps directly to the current HTTP API and does not invent protocol behavior.

## Local install

From the repo root:

```bash
cd sdk/js
npm install
npm run build
```

## Local use

After building, import the client from the generated package output in your own Node.js code, or use the included example.

```ts
import { AnchorChainClient } from "@anchorchain/sdk-js";

const client = new AnchorChainClient({
  baseUrl: "http://127.0.0.1:8081",
});

const health = await client.health();
console.log(health.status);
```

## Run the included example

The example expects a running AnchorChain HTTP API. Local devnet works well:

```bash
make build
./bin/anchorchaind devnet
```

Then in another terminal:

```bash
cd sdk/js
npm install
npm run example
```

Environment variables used by the example:

- `ANCHORCHAIN_API` default: `http://127.0.0.1:8081`
- `ANCHORCHAIN_API_TOKEN` optional

## Local devnet troubleshooting

On local devnet, write requests may succeed before follow-up reads or receipts are available.

- `Missing Chain Head` usually means the chain is not readable from the API yet.
- `receipt not available` means the entry exists but receipt data is not ready yet.
- `Receipt creation error` can appear temporarily on local devnet while receipt generation lags or fails.
- network errors such as connection refusal usually mean the local API is not up yet.

The SDK keeps the raw error body/status available on `AnchorChainSDKError` and marks common temporary states as retriable.

## Short usage example

```ts
import { AnchorChainClient } from "@anchorchain/sdk-js";

const client = new AnchorChainClient({ baseUrl: "http://127.0.0.1:8081" });

const chain = await client.createChain({
  extIds: ["demo"],
  schema: "json",
  payload: { hello: "anchorchain" },
});

console.log(chain.chainId);
```
