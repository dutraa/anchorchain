# AnchorChain HTTP API

Developer-facing reference for the routes currently implemented in [api/server.go](/C:/Users/tonyd/OneDrive/Desktop/projects/anchorchain/api/server.go).

Base URL in local devnet: `http://127.0.0.1:8081`

If `ANCHORCHAIN_API_TOKEN` is configured on the daemon, send it as `X-Anchorchain-Api-Token: <token>`. `Authorization: Bearer <token>` is also accepted.

## `GET /health`

Returns current node health and chain heights.

- Method: `GET`
- Path: `/health`
- Request body example: none
- Response body example:

```json
{
  "status": "ok",
  "heights": {
    "directoryblockheight": 12,
    "entryblockheight": 12,
    "entryheight": 34
  }
}
```

- Common error cases:
  - `502` if the upstream Factom RPC is unavailable
- Curl:

```bash
curl -s http://127.0.0.1:8081/health
```

## `POST /chains`

Creates a new chain by committing and revealing its first entry.

- Method: `POST`
- Path: `/chains`
- Request body example:

```json
{
  "extIds": ["demo", "chain"],
  "extIdsEncoding": "utf-8",
  "schema": "json",
  "payload": {
    "hello": "anchorchain"
  }
}
```

- Response body example:

```json
{
  "success": true,
  "status": "pending",
  "chainId": "8888888888888888888888888888888888888888888888888888888888888888",
  "entryHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "txId": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "message": "chain submitted",
  "schema": "json",
  "structured": true
}
```

- Common error cases:
  - `400` invalid JSON
  - `400` `ecPrivateKey is required` outside devnet
  - `400` `payload is required when schema is set`
  - `400` `payload must be valid JSON`
  - `400` unsupported `contentEncoding`
  - `400` invalid base64 in `content` or `extIds`
  - `400` invalid EC private key
  - `502` commit or reveal failure from upstream RPC
- Curl:

```bash
curl -s -X POST http://127.0.0.1:8081/chains \
  -H "Content-Type: application/json" \
  -d '{
    "extIds": ["demo", "chain"],
    "extIdsEncoding": "utf-8",
    "schema": "json",
    "payload": {"hello":"anchorchain"}
  }'
```

Notes:

- If `schema` is set, the server stores compact JSON from `payload` and prepends `schema:<name>` to the ExtIDs.
- If `schema` is not set, use `content` plus optional `contentEncoding` (`utf-8` or `base64`).
- In devnet, the daemon injects a demo EC key when `ecPrivateKey` is omitted.

## `POST /chains/{chainId}/entries`

Appends an entry to an existing chain.

- Method: `POST`
- Path: `/chains/{chainId}/entries`
- Request body example:

```json
{
  "extIds": ["demo"],
  "extIdsEncoding": "utf-8",
  "schema": "json",
  "payload": {
    "step": 2
  }
}
```

- Response body example:

```json
{
  "success": true,
  "status": "pending",
  "chainId": "8888888888888888888888888888888888888888888888888888888888888888",
  "entryHash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
  "txId": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
  "message": "entry submitted",
  "schema": "json",
  "structured": true
}
```

- Common error cases:
  - `400` invalid JSON
  - `400` missing or invalid `ecPrivateKey`
  - `400` `payload` or `schema` mismatch
  - `400` unsupported `contentEncoding`
  - `502` commit or reveal failure from upstream RPC
- Curl:

```bash
curl -s -X POST http://127.0.0.1:8081/chains/<CHAIN_ID>/entries \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "json",
    "payload": {"step":2}
  }'
```

## `GET /chains/{chainId}`

Returns a compact chain summary.

- Method: `GET`
- Path: `/chains/{chainId}`
- Request body example: none
- Response body example:

```json
{
  "chainId": "8888888888888888888888888888888888888888888888888888888888888888",
  "entryCount": 2,
  "latestEntryHash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
  "latestEntryTimestamp": 1710000000
}
```

- Common error cases:
  - `202` chain is pending confirmation
  - `404` chain not found
  - `400` invalid chain identifier passed through to upstream RPC
  - `502` upstream RPC failure
- Curl:

```bash
curl -s http://127.0.0.1:8081/chains/<CHAIN_ID>
```

## `GET /chains/{chainId}/entries`

Lists entries on a chain with offset and limit pagination.

- Method: `GET`
- Path: `/chains/{chainId}/entries`
- Request body example: none
- Response body example:

```json
{
  "chainId": "8888888888888888888888888888888888888888888888888888888888888888",
  "entries": [
    {
      "entryHash": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "timestamp": 1710000000,
      "extIds": ["ZGVtbw==", "Y2hhaW4="],
      "schema": "json",
      "structured": true
    }
  ],
  "limit": 50,
  "offset": 0,
  "total": 2
}
```

- Common error cases:
  - `400` `limit must be a positive integer`
  - `400` `offset must be zero or a positive integer`
  - `202` chain is pending confirmation
  - `404` chain not found
  - `502` upstream RPC failure
- Curl:

```bash
curl -s "http://127.0.0.1:8081/chains/<CHAIN_ID>/entries?limit=10&offset=0"
```

Notes:

- Default `limit` is `50`.
- Maximum `limit` is `500`.
- Returned `extIds` are base64-encoded.

## `GET /entries/{entryHash}`

Loads a single entry by hash.

- Method: `GET`
- Path: `/entries/{entryHash}`
- Request body example: none
- Response body example:

```json
{
  "entryHash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
  "chainId": "8888888888888888888888888888888888888888888888888888888888888888",
  "extIds": ["c2NoZW1hOmpzb24=", "ZGVtbw=="],
  "schema": "json",
  "structured": true,
  "content": "{\"step\":2}",
  "contentEncoding": "utf-8",
  "decodedPayload": {
    "step": 2
  }
}
```

- Common error cases:
  - `404` entry not found
  - `400` invalid entry hash passed through to upstream RPC
  - `502` upstream RPC failure
  - `502` if a structured entry cannot be decoded as JSON
- Curl:

```bash
curl -s http://127.0.0.1:8081/entries/<ENTRY_HASH>
```

Notes:

- For structured entries, `content` is returned as UTF-8 JSON text and `decodedPayload` is included.
- For non-structured entries, `content` is base64 and `contentEncoding` is `base64`.

## `POST /receipts/verify`

Fetches the receipt for an entry and optionally includes raw entry bytes.

- Method: `POST`
- Path: `/receipts/verify`
- Request body example:

```json
{
  "entryHash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
  "includeRawEntry": false
}
```

- Response body example:

```json
{
  "success": true,
  "entryHash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
  "receipt": {
    "entry": {
      "raw": ""
    }
  },
  "message": "receipt retrieved"
}
```

- Common error cases:
  - `400` invalid JSON
  - `400` `entryHash is required`
  - `404` receipt not available
  - `400` invalid entry hash passed through to upstream RPC
  - `502` upstream RPC failure
- Curl:

```bash
curl -s -X POST http://127.0.0.1:8081/receipts/verify \
  -H "Content-Type: application/json" \
  -d '{
    "entryHash": "<ENTRY_HASH>",
    "includeRawEntry": false
  }'
```
