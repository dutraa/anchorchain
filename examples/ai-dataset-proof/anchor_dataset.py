import argparse
import hashlib
import json
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Callable, Optional, TypeVar

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "sdk" / "python"))

from anchorchain import AnchorChainClient, AnchorChainSDKError  # noqa: E402

T = TypeVar("T")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Hash a local dataset file and anchor only its hash plus metadata in AnchorChain.",
    )
    parser.add_argument("file", help="Path to the local dataset file")
    parser.add_argument("--api", default="http://127.0.0.1:8081", help="AnchorChain API base URL")
    parser.add_argument("--token", help="Optional AnchorChain API token")
    parser.add_argument("--chain", dest="chain_id", help="Existing chain ID to append to")
    parser.add_argument("--description", help="Optional short description")
    parser.add_argument("--tag", help="Optional short tag")
    parser.add_argument("--attempts", type=int, default=48, help="Receipt polling attempts")
    parser.add_argument("--delay-seconds", type=int, default=8, help="Delay between receipt polls")
    return parser.parse_args()


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def describe_error(error: Exception) -> str:
    if isinstance(error, AnchorChainSDKError):
        suffix = f" (status {error.status})" if error.status else ""
        return f"{error.kind}{suffix}: {error.message}"
    return str(error)


def retry(
    label: str,
    fn: Callable[[], T],
    attempts: int,
    delay_seconds: int,
    should_retry: Optional[Callable[[Exception], bool]] = None,
) -> T:
    last_error: Optional[Exception] = None
    for attempt in range(1, attempts + 1):
        try:
            return fn()
        except Exception as error:
            last_error = error
            retriable = not isinstance(error, AnchorChainSDKError) or error.retriable
            if should_retry is not None:
                retriable = retriable or should_retry(error)
            if attempt == attempts or not retriable:
                break
            print(f"[pending] {label} not ready yet; retrying ({attempt}/{attempts}): {describe_error(error)}")
            time.sleep(delay_seconds)
    if last_error is None:
        raise RuntimeError(f"{label} failed without returning a result")
    raise last_error


def wait_for_receipt(client: AnchorChainClient, entry_hash: str, attempts: int, delay_seconds: int):
    return retry(
        "Receipt",
        lambda: client.verify_receipt(entry_hash=entry_hash, include_raw_entry=False),
        attempts=attempts,
        delay_seconds=delay_seconds,
    )


def build_payload(file_path: Path, sha256: str, description: Optional[str], tag: Optional[str]) -> dict:
    payload = {
        "type": "ai-dataset-proof/v1",
        "fileName": file_path.name,
        "sha256": sha256,
        "fileSizeBytes": file_path.stat().st_size,
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }
    if description:
        payload["description"] = description
    if tag:
        payload["tag"] = tag
    return payload


def main() -> int:
    args = parse_args()
    file_path = Path(args.file).expanduser().resolve()
    if not file_path.is_file():
        raise FileNotFoundError(f"Dataset file not found: {file_path}")

    file_sha256 = sha256_file(file_path)
    payload = build_payload(file_path, file_sha256, args.description, args.tag)

    print("[ok] Local dataset fingerprint:")
    print(
        json.dumps(
            {
                "file": str(file_path),
                "sha256": file_sha256,
                "fileSizeBytes": file_path.stat().st_size,
            },
            indent=2,
        )
    )

    client = AnchorChainClient(base_url=args.api, token=args.token)

    health = client.health()
    print("[ok] AnchorChain API health:")
    print(json.dumps({"status": health.status, "heights": health.heights}, indent=2))

    if args.chain_id:
        print(f"[pending] Anchoring into existing dataset chain {args.chain_id}...")
        write_result = client.write_entry(
            args.chain_id,
            schema="json",
            payload=payload,
        )
        chain_id = args.chain_id
    else:
        print("[pending] Creating a new dataset proof chain and anchoring the first entry...")
        write_result = client.create_chain(
            ext_ids=["ai-dataset-proof", "sha256"],
            ext_ids_encoding="utf-8",
            schema="json",
            payload=payload,
        )
        chain_id = write_result.chain_id

    if not write_result.entry_hash:
        raise RuntimeError("AnchorChain did not return an entry hash.")

    print("[ok] Anchor result:")
    print(
        json.dumps(
            {
                "chainId": chain_id,
                "entryHash": write_result.entry_hash,
                "status": write_result.status,
                "message": write_result.message,
            },
            indent=2,
        )
    )

    print("[wait] Waiting for receipt confirmation. Local devnet receipts can take a bit longer.")
    receipt = wait_for_receipt(client, write_result.entry_hash, args.attempts, args.delay_seconds)
    print("[ok] Receipt verified:")
    print(json.dumps({"entryHash": receipt.entry_hash, "success": receipt.success}, indent=2))

    print("[ok] Verification command:")
    print(
        "python examples/ai-dataset-proof/verify_dataset.py "
        f"\"{file_path}\" --entry {write_result.entry_hash} --api {args.api}"
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except AnchorChainSDKError as error:
        if error.retriable:
            print(f"[timeout] Timed out waiting for devnet readiness: {describe_error(error)}", file=sys.stderr)
        else:
            print(f"[error] SDK error: {describe_error(error)}", file=sys.stderr)
        if error.body is not None:
            print(json.dumps(error.body, indent=2), file=sys.stderr)
        raise SystemExit(1)
    except Exception as error:
        print(f"[error] {error}", file=sys.stderr)
        raise SystemExit(1)
