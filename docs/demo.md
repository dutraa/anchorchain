# AnchorChain Quick Demo

Spin up a devnet and exercise the CLI in under two minutes.

## 1. Build the binaries
```bash
make build
```
This compiles `anchorchaind` and `anchor-cli` into `./bin/`.

## 2. Launch the devnet
```bash
./bin/anchorchaind devnet
```
A single-node network starts locally (60-second blocks) and exposes the HTTP API on `http://127.0.0.1:8081`. Leave this running in a terminal.

## 3. Check node health
```bash
./bin/anchor-cli node health
```
Confirms the daemon is reachable and shows current directory/entry heights.

## 4. Create a chain
```bash
./bin/anchor-cli chain create \
  --extid demo --schema json \
  --payload '{"hello":"anchorchain"}'
```
Returns the new chain ID, entry hash, and txid.

## 5. Append an entry
```bash
./bin/anchor-cli entry write \
  --chain <CHAIN_ID> \
  --schema json --payload '{"step":2}'
```
Uses the devnet Entry Credit key to add another entry.

## 6. Verify a receipt
```bash
./bin/anchor-cli receipt verify --entry <ENTRY_HASH>
```
Fetches the Merkle receipt for the entry, proving it landed on the chain.

That’s it—once you see the receipt, you’ve exercised chains, entries, receipts, and the HTTP/CLI stack end-to-end. Press `Ctrl+C` in the devnet terminal to shut it down.
