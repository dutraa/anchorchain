from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Union

JSONScalar = Union[str, int, float, bool, None]
JSONValue = Union[JSONScalar, Dict[str, Any], List[Any]]


@dataclass(frozen=True)
class HealthResponse:
    status: str
    heights: Dict[str, JSONValue]

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "HealthResponse":
        return cls(
            status=str(data.get("status", "")),
            heights=dict(data.get("heights") or {}),
        )


@dataclass(frozen=True)
class WriteResponse:
    success: bool
    status: str
    chain_id: Optional[str] = None
    entry_hash: Optional[str] = None
    tx_id: Optional[str] = None
    message: Optional[str] = None
    schema: Optional[str] = None
    structured: Optional[bool] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WriteResponse":
        return cls(
            success=bool(data.get("success")),
            status=str(data.get("status", "")),
            chain_id=data.get("chainId"),
            entry_hash=data.get("entryHash"),
            tx_id=data.get("txId"),
            message=data.get("message"),
            schema=data.get("schema"),
            structured=data.get("structured"),
        )


@dataclass(frozen=True)
class ChainResponse:
    chain_id: str
    entry_count: Optional[int] = None
    latest_entry_hash: Optional[str] = None
    latest_entry_timestamp: Optional[int] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ChainResponse":
        return cls(
            chain_id=str(data.get("chainId", "")),
            entry_count=data.get("entryCount"),
            latest_entry_hash=data.get("latestEntryHash"),
            latest_entry_timestamp=data.get("latestEntryTimestamp"),
        )


@dataclass(frozen=True)
class EntrySummary:
    entry_hash: str
    timestamp: Optional[int]
    ext_ids: List[str]
    schema: Optional[str]
    structured: bool

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EntrySummary":
        return cls(
            entry_hash=str(data.get("entryHash", "")),
            timestamp=data.get("timestamp"),
            ext_ids=list(data.get("extIds") or []),
            schema=data.get("schema"),
            structured=bool(data.get("structured")),
        )


@dataclass(frozen=True)
class EntriesResponse:
    chain_id: str
    entries: List[EntrySummary]
    limit: int
    offset: int
    total: int

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EntriesResponse":
        return cls(
            chain_id=str(data.get("chainId", "")),
            entries=[EntrySummary.from_dict(item) for item in data.get("entries") or []],
            limit=int(data.get("limit", 0)),
            offset=int(data.get("offset", 0)),
            total=int(data.get("total", 0)),
        )


@dataclass(frozen=True)
class EntryResponse:
    entry_hash: str
    chain_id: str
    ext_ids: List[str]
    schema: Optional[str]
    structured: bool
    content: str
    content_encoding: str
    decoded_payload: Optional[JSONValue] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EntryResponse":
        return cls(
            entry_hash=str(data.get("entryHash", "")),
            chain_id=str(data.get("chainId", "")),
            ext_ids=list(data.get("extIds") or []),
            schema=data.get("schema"),
            structured=bool(data.get("structured")),
            content=str(data.get("content", "")),
            content_encoding=str(data.get("contentEncoding", "")),
            decoded_payload=data.get("decodedPayload"),
        )


@dataclass(frozen=True)
class VerifyReceiptResponse:
    success: bool
    entry_hash: str
    receipt: Optional[JSONValue] = None
    message: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "VerifyReceiptResponse":
        return cls(
            success=bool(data.get("success")),
            entry_hash=str(data.get("entryHash", "")),
            receipt=data.get("receipt"),
            message=data.get("message"),
        )
