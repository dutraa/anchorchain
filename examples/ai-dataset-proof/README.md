# AI Dataset Provenance Demo

A minimal Python example for anchoring AI dataset provenance in AnchorChain.

## What it does

This example:

1. hashes a local dataset file with SHA-256
2. anchors only the hash and small metadata in AnchorChain
3. waits for and verifies a receipt
4. lets you verify the file later by re-hashing it and comparing it with the anchored record

The full dataset contents are **not** stored on-chain. Only the hash and metadata are anchored.

## Anchored metadata

The anchored JSON payload is intentionally small and focused on:

- file name
- SHA-256 hash
- file size
- timestamp
- optional description
- optional tag

## Files

- `anchor_dataset.py` anchors a dataset proof
- `verify_dataset.py` re-hashes a file and checks it against an anchored entry
- `sample-dataset.csv` tiny local sample dataset

## Run against local devnet

Start devnet from the repo root:

```bash
make build
./bin/anchorchaind devnet
```

Then in another terminal, anchor the sample dataset:

```bash
python examples/ai-dataset-proof/anchor_dataset.py examples/ai-dataset-proof/sample-dataset.csv --description "demo training sample" --tag demo
```

The script prints the resulting `chainId` and `entryHash`.

Then verify the file later using the returned entry hash:

```bash
python examples/ai-dataset-proof/verify_dataset.py examples/ai-dataset-proof/sample-dataset.csv --entry <ENTRY_HASH>
```

## Environment and flags

Both scripts support:

- `--api` default: `http://127.0.0.1:8081`
- `--token` optional

`anchor_dataset.py` also supports:

- `--chain <CHAIN_ID>` to append to an existing chain instead of creating a new one
- `--description "..."` optional description
- `--tag ...` optional tag
- `--attempts` and `--delay-seconds` for receipt polling

`verify_dataset.py` supports:

- `--attempts` and `--delay-seconds` for receipt polling

## Example flow

Anchor once:

```bash
python examples/ai-dataset-proof/anchor_dataset.py examples/ai-dataset-proof/sample-dataset.csv --description "baseline dataset snapshot" --tag v1
```

Verify later:

```bash
python examples/ai-dataset-proof/verify_dataset.py examples/ai-dataset-proof/sample-dataset.csv --entry <ENTRY_HASH>
```

If you change the file contents and run `verify_dataset.py` again, the SHA-256 comparison should fail.

## Notes

- This is a provenance demo, not a full dataset registry.
- The example uses the existing `sdk/python` client from this repository.
- Receipt availability may take time on local devnet, so both scripts allow a patient retry window.
