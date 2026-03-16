import argparse
import hashlib
import json
import sys
import time
from pathlib import Path
from typing import Callable, Optional, TypeVar

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "sdk" / "python"))

from anchorchain import AnchorChainClient, AnchorChainSDKError  # noqa: E402

T = TypeVar("T")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Re-hash a local dataset file and compare it with an anchored AnchorChain entry.",
    )
    parser.add_argument("file", help="Path to the local dataset file")
    parser.add_argument("--entry", required=True, help="AnchorChain entry hash to verify against")
    parser.add_argument("--api", default="http://127.0.0.1:8081", help="AnchorChain API base URL")
    parser.add_argument("--token", help="Optional AnchorChain API token")
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


def retry(label: str, fn: Callable[[], T], attempts: int, delay_seconds: int) -> T:
    last_error: Optional[Exception] = None
    for attempt in range(1, attempts + 1):
        try:
            return fn()
        except Exception as error:
            last_error = error
            retriable = not isinstance(error, AnchorChainSDKError) or error.retriable
            if attempt == attempts or not retriable:
                break
            print(f"[pending] {label} not ready yet; retrying ({attempt}/{attempts}): {describe_error(error)}")
            time.sleep(delay_seconds)
    if last_error is None:
        raise RuntimeError(f"{label} failed without returning a result")
    raise last_error


def is_entry_lookup_pending(error: Exception) -> bool:
    return (
        isinstance(error, AnchorChainSDKError)
        and error.status == 404
        and "entry not found" in error.message.strip().lower()
    )


def wait_for_entry(client: AnchorChainClient, entry_hash: str, attempts: int, delay_seconds: int):
    last_error: Optional[Exception] = None
    for attempt in range(1, attempts + 1):
        try:
            return client.get_entry(entry_hash)
        except Exception as error:
            last_error = error
            retriable = (
                not isinstance(error, AnchorChainSDKError)
                or error.retriable
                or is_entry_lookup_pending(error)
            )
            if attempt == attempts or not retriable:
                break
            print(f"[pending] Entry not ready yet; retrying ({attempt}/{attempts}): {describe_error(error)}")
            time.sleep(delay_seconds)
    if last_error is None:
        raise RuntimeError("Entry lookup failed without returning a result")
    raise last_error


def main() -> int:
    args = parse_args()
    file_path = Path(args.file).expanduser().resolve()
    if not file_path.is_file():
        raise FileNotFoundError(f"Dataset file not found: {file_path}")

    file_sha256 = sha256_file(file_path)
    file_size = file_path.stat().st_size

    client = AnchorChainClient(base_url=args.api, token=args.token)

    entry = wait_for_entry(client, args.entry, args.attempts, args.delay_seconds)
    if not isinstance(entry.decoded_payload, dict):
        raise RuntimeError("Anchored entry does not contain a structured JSON payload.")

    payload = entry.decoded_payload
    anchored_sha256 = payload.get("sha256")
    anchored_file_name = payload.get("fileName")
    anchored_file_size = payload.get("fileSizeBytes")

    print("[ok] Anchored metadata:")
    print(json.dumps(payload, indent=2))

    print("[ok] Local dataset fingerprint:")
    print(
        json.dumps(
            {
                "file": str(file_path),
                "sha256": file_sha256,
                "fileSizeBytes": file_size,
            },
            indent=2,
        )
    )

    sha_matches = anchored_sha256 == file_sha256
    size_matches = anchored_file_size == file_size
    name_matches = anchored_file_name == file_path.name

    print("[wait] Verifying the receipt for the anchored entry...")
    receipt = retry(
        "Receipt",
        lambda: client.verify_receipt(entry_hash=args.entry, include_raw_entry=False),
        attempts=args.attempts,
        delay_seconds=args.delay_seconds,
    )

    result = {
        "entryHash": args.entry,
        "sha256Match": sha_matches,
        "fileSizeMatch": size_matches,
        "fileNameMatch": name_matches,
        "receiptVerified": receipt.success,
    }
    print("[ok] Verification result:")
    print(json.dumps(result, indent=2))

    if not sha_matches:
        print("[error] SHA-256 mismatch. The local file does not match the anchored proof.", file=sys.stderr)
        return 1

    if not size_matches or not name_matches:
        print("[warning] The hash matches, but some metadata differs.", file=sys.stderr)

    print("[ok] Dataset proof matches the anchored record.")
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
