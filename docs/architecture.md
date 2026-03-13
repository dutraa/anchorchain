# AnchorChain Architecture

AnchorChain modernizes the original Factom Protocol stack with updated tooling while preserving the core data model (chains, entries, and receipts).

## High-Level Components
- **anchorchaind (daemon)** — wraps the vendored Factom node (`node/`) and exposes a simplified runtime plus a developer-friendly HTTP API. It can run against public networks or the built-in devnet profile.
- **HTTP API layer (`api/`)** — provides REST-style endpoints for creating chains, writing entries, querying chain state, fetching entries, verifying receipts, and reporting node health. Requests are translated into Factom RPC calls under the hood.
- **anchor-cli (`cmd/anchor-cli`)** — a thin Go CLI that talks to the HTTP API. It offers subcommands such as `chain create`, `entry write`, `receipt verify`, and `node health`, returning either human-readable summaries or raw JSON.
- **Wallet helpers (`wallet/`)** — legacy Factom wallet utilities that `anchorchaind wallet ...` reuses until a standalone wallet/CLI replacement is ready.
- **Vendored protocol (`node/`)** — the upstream Factom implementation, left untouched to maintain protocol compatibility.

## Data Model
- **Chains**: ordered sequences of entries identified by a unique chain ID (a hash of the first entry). They act as namespaces for application data.
- **Entries**: immutable payloads up to ~10 KB. Entries contain ExtIDs (metadata) plus content (opaque bytes or structured JSON). Entries are paid for with Entry Credits derived from Factoids.
- **Receipts**: Merkle proofs showing an entry’s inclusion in a directory block and, ultimately, anchoring into external blockchains (e.g., Bitcoin). Receipts can be retrieved and locally verified for audit trails.

## Devnet Workflow
1. `./bin/anchorchaind devnet` spins up a single-node network with 60-second blocks and pre-funded addresses.
2. The daemon exposes the HTTP API on `127.0.0.1:8081` and the legacy Factom RPC on `localhost:8088`.
3. `anchor-cli` defaults to this API endpoint, so developers can create chains (`chain create`), append entries (`entry write`), and fetch receipts without additional configuration.
4. Because the devnet seeds a demo Entry Credit key, write operations work out of the box. For custom networks, pass `--ec-key` (or set `ANCHORCHAIN_EC_PRIVATE`).

## Deployment Considerations
- AnchorChain is still a developer preview. Secure HTTP access with tokens (`ANCHORCHAIN_API_TOKEN`) and limit exposure unless absolutely necessary.
- When targeting production networks, configure the daemon’s Factom RPC endpoint (`FACTOMD_RPC_ADDR`) and wallet/key management according to your environment.

This architecture allows new tooling to evolve around the original Factom consensus engine while keeping the operational surface approachable for modern dev workflows.
