import os
import sys
import time
from typing import Callable, Optional, TypeVar

from anchorchain import AnchorChainClient, AnchorChainSDKError

T = TypeVar("T")


def describe_error(error: Exception) -> str:
    if isinstance(error, AnchorChainSDKError):
        suffix = f" (status {error.status})" if error.status else ""
        return f"{error.kind}{suffix}: {error.message}"
    return str(error)


def retry(label: str, fn: Callable[[], T], attempts: int = 24, delay_seconds: int = 5) -> T:
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


def wait_for_chain(client: AnchorChainClient, chain_id: str):
    return retry("Chain", lambda: client.get_chain(chain_id))


def wait_for_entry(client: AnchorChainClient, entry_hash: str):
    return retry("Entry", lambda: client.get_entry(entry_hash))


def wait_for_receipt(client: AnchorChainClient, entry_hash: str):
    return retry("Receipt", lambda: client.verify_receipt(entry_hash=entry_hash, include_raw_entry=False))


def main() -> int:
    client = AnchorChainClient(
        base_url=os.getenv("ANCHORCHAIN_API", "http://127.0.0.1:8081"),
        token=os.getenv("ANCHORCHAIN_API_TOKEN"),
    )

    nonce = time.strftime("%Y-%m-%dT%H:%M:%S")

    health = client.health()
    print("[ok] Health:", health)

    created = client.create_chain(
        ext_ids=["sdk", "python", "example", nonce],
        ext_ids_encoding="utf-8",
        schema="json",
        payload={"createdBy": "sdk/python", "step": 1, "nonce": nonce},
    )
    print("[ok] Create chain:", created)

    if not created.chain_id or not created.entry_hash:
        raise RuntimeError("Chain creation did not return chain_id and entry_hash.")

    written = client.write_entry(
        created.chain_id,
        schema="json",
        payload={"step": 2, "note": "follow-up entry", "nonce": nonce},
    )
    print("[ok] Write entry:", written)

    chain = wait_for_chain(client, created.chain_id)
    print("[ok] Chain:", chain)

    entries = client.get_entries(created.chain_id, limit=10, offset=0)
    print("[ok] Entries:", entries)

    entry_hash = written.entry_hash or created.entry_hash
    if not entry_hash:
        raise RuntimeError("Entry write did not return entry_hash.")

    entry = wait_for_entry(client, entry_hash)
    print("[ok] Entry:", entry)

    receipt = wait_for_receipt(client, entry_hash)
    print("[ok] Receipt:", receipt)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except AnchorChainSDKError as error:
        if error.retriable:
            print(f"[timeout] Timed out waiting for a devnet-visible state: {describe_error(error)}", file=sys.stderr)
        else:
            print(f"[error] SDK error: {describe_error(error)}", file=sys.stderr)
        if error.body is not None:
            print(error.body, file=sys.stderr)
        raise SystemExit(1)
    except Exception as error:
        print(f"[error] {error}", file=sys.stderr)
        raise SystemExit(1)
