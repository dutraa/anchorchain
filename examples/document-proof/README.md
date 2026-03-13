# Document Proof Demo

A minimal proof-of-existence demo for AnchorChain.

## What it does

This demo:

1. reads a local file
2. hashes it locally with SHA-256
3. anchors only the hash and basic metadata through AnchorChain
4. waits for and fetches a receipt

The full file is not stored on-chain. Only the hash and metadata are anchored.

## Files

- `src/main.ts` CLI demo
- `samples/hello.txt` tiny sample file

## Install

From the repo root:

```bash
cd sdk/js
npm install
npm run build

cd ../../examples/document-proof
npm install
npm run build
```

## Run against local devnet

Start devnet from the repo root:

```bash
make build
./bin/anchorchaind devnet
```

In another terminal run the demo:

```bash
cd examples/document-proof
npm install
npm run demo -- ./samples/hello.txt
```

The demo uses these environment variables if you need them:

- `ANCHORCHAIN_API` default: `http://127.0.0.1:8081`
- `ANCHORCHAIN_API_TOKEN` optional
- `ANCHORCHAIN_EC_PRIVATE` optional outside devnet

You can also pass flags directly:

```bash
npm run demo -- ./samples/hello.txt --api http://127.0.0.1:8081
```

To anchor into an existing chain instead of creating a new one:

```bash
npm run demo -- ./samples/hello.txt --chain <CHAIN_ID>
```

To wait longer for receipt availability:

```bash
npm run demo -- ./samples/hello.txt --attempts 48 --delay-ms 5000
```

## Test that file changes change the hash

Run the demo once with the sample file:

```bash
npm run demo -- ./samples/hello.txt
```

Then edit `samples/hello.txt`, add or remove a character, and run it again:

```bash
npm run demo -- ./samples/hello.txt
```

The local SHA-256 hash printed at the top should be different.

## Notes

- Local hashing is separate from on-chain anchoring.
- Receipt availability may take time on devnet, so the demo polls the receipt endpoint.
- If the API never returns a receipt within the retry window, the demo exits with an error rather than inventing success.

## Local devnet troubleshooting

Expected temporary states during local development include:

- writes succeed before chain reads become available
- `Missing Chain Head` while indexing catches up
- receipt polling returns `receipt not available` or `Receipt creation error`
- temporary connection failures if the local API is still starting

The demo reports these as pending or timed-out states when appropriate, but it still prints the underlying SDK/API error so you can see what actually happened.
