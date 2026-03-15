from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Callable, Dict, Literal, Mapping, Optional, Union
from urllib.parse import quote

import requests

from .models import (
    ChainResponse,
    EntriesResponse,
    EntryResponse,
    HealthResponse,
    VerifyReceiptResponse,
    WriteResponse,
)

AnchorChainErrorKind = Literal[
    "api_unreachable",
    "missing_chain_head",
    "receipt_not_ready",
    "receipt_creation_error",
    "bad_request",
    "server_error",
    "unknown",
]


@dataclass
class AnchorChainSDKError(Exception):
    message: str
    kind: AnchorChainErrorKind
    retriable: bool
    status: Optional[int] = None
    body: Any = None
    cause: Optional[BaseException] = None

    def __post_init__(self) -> None:
        super().__init__(self.message)


def _normalize_base_url(base_url: Optional[str]) -> str:
    value = (base_url or "http://127.0.0.1:8081").strip()
    if not value:
        return "http://127.0.0.1:8081"
    if value.lower().startswith(("http://", "https://")):
        return value.rstrip("/")
    return f"http://{value}".rstrip("/")


def _resolve_message(body: Any, fallback: str) -> str:
    if not isinstance(body, dict):
        return fallback
    for key in ("error", "message", "status"):
        value = body.get(key)
        if isinstance(value, str) and value.strip():
            return value
    return fallback


def _classify_api_error(status: int, message: str) -> tuple[AnchorChainErrorKind, bool]:
    normalized = message.strip().lower()
    if "missing chain head" in normalized:
        return "missing_chain_head", status in (404, 202)
    if "chain is pending confirmation" in normalized:
        return "missing_chain_head", True
    if "receipt not available" in normalized:
        return "receipt_not_ready", status == 404
    if "receipt creation error" in normalized:
        return "receipt_creation_error", status >= 500
    if status == 400:
        return "bad_request", False
    if status >= 500:
        return "server_error", True
    return "unknown", False


def _classify_transport_error(error: BaseException) -> AnchorChainSDKError:
    message = str(error) or "Unable to reach AnchorChain API"
    return AnchorChainSDKError(
        message=message,
        kind="api_unreachable",
        retriable=True,
        cause=error,
    )


class AnchorChainClient:
    def __init__(
        self,
        *,
        base_url: Optional[str] = None,
        token: Optional[str] = None,
        session: Optional[requests.Session] = None,
        timeout: Union[float, tuple[float, float]] = 30,
    ) -> None:
        self.base_url = _normalize_base_url(base_url)
        self.token = token
        self.session = session or requests.Session()
        self.timeout = timeout

    def health(self) -> HealthResponse:
        return self._request("GET", "/health", response_factory=HealthResponse.from_dict)

    def create_chain(
        self,
        *,
        ext_ids: Optional[list[str]] = None,
        ext_ids_encoding: Optional[str] = None,
        content: Optional[str] = None,
        content_encoding: Optional[str] = None,
        schema: Optional[str] = None,
        payload: Any = None,
        ec_private_key: Optional[str] = None,
    ) -> WriteResponse:
        body = self._build_write_body(
            ext_ids=ext_ids,
            ext_ids_encoding=ext_ids_encoding,
            content=content,
            content_encoding=content_encoding,
            schema=schema,
            payload=payload,
            ec_private_key=ec_private_key,
        )
        return self._request("POST", "/chains", body=body, response_factory=WriteResponse.from_dict)

    def write_entry(
        self,
        chain_id: str,
        *,
        ext_ids: Optional[list[str]] = None,
        ext_ids_encoding: Optional[str] = None,
        content: Optional[str] = None,
        content_encoding: Optional[str] = None,
        schema: Optional[str] = None,
        payload: Any = None,
        ec_private_key: Optional[str] = None,
    ) -> WriteResponse:
        encoded_chain_id = quote(chain_id.strip(), safe="")
        body = self._build_write_body(
            ext_ids=ext_ids,
            ext_ids_encoding=ext_ids_encoding,
            content=content,
            content_encoding=content_encoding,
            schema=schema,
            payload=payload,
            ec_private_key=ec_private_key,
        )
        return self._request(
            "POST",
            f"/chains/{encoded_chain_id}/entries",
            body=body,
            response_factory=WriteResponse.from_dict,
        )

    def get_chain(self, chain_id: str) -> ChainResponse:
        encoded_chain_id = quote(chain_id.strip(), safe="")
        return self._request("GET", f"/chains/{encoded_chain_id}", response_factory=ChainResponse.from_dict)

    def get_entries(
        self,
        chain_id: str,
        *,
        limit: Optional[int] = None,
        offset: Optional[int] = None,
    ) -> EntriesResponse:
        encoded_chain_id = quote(chain_id.strip(), safe="")
        params: Dict[str, int] = {}
        if limit is not None:
            params["limit"] = limit
        if offset is not None:
            params["offset"] = offset
        return self._request(
            "GET",
            f"/chains/{encoded_chain_id}/entries",
            params=params or None,
            response_factory=EntriesResponse.from_dict,
        )

    def get_entry(self, entry_hash: str) -> EntryResponse:
        encoded_entry_hash = quote(entry_hash.strip(), safe="")
        return self._request("GET", f"/entries/{encoded_entry_hash}", response_factory=EntryResponse.from_dict)

    def verify_receipt(
        self,
        request: Optional[Union[str, Mapping[str, Any]]] = None,
        *,
        entry_hash: Optional[str] = None,
        include_raw_entry: Optional[bool] = None,
    ) -> VerifyReceiptResponse:
        if isinstance(request, str):
            body: Dict[str, Any] = {"entryHash": request}
        elif request is not None:
            body = dict(request)
        else:
            if not entry_hash:
                raise ValueError("entry_hash is required")
            body = {"entryHash": entry_hash}

        if entry_hash and "entryHash" not in body:
            body["entryHash"] = entry_hash
        if include_raw_entry is not None:
            body["includeRawEntry"] = include_raw_entry

        return self._request(
            "POST",
            "/receipts/verify",
            body=body,
            response_factory=VerifyReceiptResponse.from_dict,
        )

    def createChain(self, request: Mapping[str, Any]) -> WriteResponse:
        return self.create_chain(**self._translate_write_request(request))

    def writeEntry(self, chain_id: str, request: Mapping[str, Any]) -> WriteResponse:
        return self.write_entry(chain_id, **self._translate_write_request(request))

    def getChain(self, chain_id: str) -> ChainResponse:
        return self.get_chain(chain_id)

    def getEntries(self, chain_id: str, options: Optional[Mapping[str, Any]] = None) -> EntriesResponse:
        options = dict(options or {})
        return self.get_entries(chain_id, limit=options.get("limit"), offset=options.get("offset"))

    def getEntry(self, entry_hash: str) -> EntryResponse:
        return self.get_entry(entry_hash)

    def verifyReceipt(
        self,
        request: Optional[Union[str, Mapping[str, Any]]] = None,
        *,
        entry_hash: Optional[str] = None,
        include_raw_entry: Optional[bool] = None,
    ) -> VerifyReceiptResponse:
        return self.verify_receipt(request=request, entry_hash=entry_hash, include_raw_entry=include_raw_entry)

    def _request(
        self,
        method: str,
        path: str,
        *,
        body: Optional[Mapping[str, Any]] = None,
        params: Optional[Mapping[str, Any]] = None,
        response_factory: Callable[[Dict[str, Any]], Any],
    ) -> Any:
        try:
            response = self.session.request(
                method=method,
                url=f"{self.base_url}{path}",
                headers=self._build_headers(has_body=body is not None),
                json=body,
                params=params,
                timeout=self.timeout,
            )
        except requests.RequestException as error:
            raise _classify_transport_error(error) from error

        content_type = response.headers.get("content-type", "")
        response_body: Any
        if "application/json" in content_type.lower():
            response_body = response.json()
        else:
            response_body = response.text

        if not response.ok or self._is_api_error_body(response_body):
            message = _resolve_message(
                response_body,
                f"AnchorChain API request failed with status {response.status_code}",
            )
            kind, retriable = _classify_api_error(response.status_code, message)
            raise AnchorChainSDKError(
                message=message,
                status=response.status_code,
                body=response_body,
                kind=kind,
                retriable=retriable,
            )

        if not isinstance(response_body, dict):
            raise AnchorChainSDKError(
                message="AnchorChain API returned a non-JSON response",
                kind="unknown",
                retriable=False,
                status=response.status_code,
                body=response_body,
            )

        return response_factory(response_body)

    def _is_api_error_body(self, response_body: Any) -> bool:
        if not isinstance(response_body, dict):
            return False
        error_value = response_body.get("error")
        if isinstance(error_value, str) and error_value.strip():
            return True
        return False

    def _build_headers(self, *, has_body: bool) -> Dict[str, str]:
        headers: Dict[str, str] = {}
        if has_body:
            headers["Content-Type"] = "application/json"
        if self.token:
            headers["X-Anchorchain-Api-Token"] = self.token
        return headers

    def _build_write_body(
        self,
        *,
        ext_ids: Optional[list[str]],
        ext_ids_encoding: Optional[str],
        content: Optional[str],
        content_encoding: Optional[str],
        schema: Optional[str],
        payload: Any,
        ec_private_key: Optional[str],
    ) -> Dict[str, Any]:
        body: Dict[str, Any] = {}
        if ext_ids is not None:
            body["extIds"] = ext_ids
        if ext_ids_encoding is not None:
            body["extIdsEncoding"] = ext_ids_encoding
        if content is not None:
            body["content"] = content
        if content_encoding is not None:
            body["contentEncoding"] = content_encoding
        if schema is not None:
            body["schema"] = schema
        if payload is not None:
            body["payload"] = payload
        if ec_private_key is not None:
            body["ecPrivateKey"] = ec_private_key
        return body

    def _translate_write_request(self, request: Mapping[str, Any]) -> Dict[str, Any]:
        return {
            "ext_ids": request.get("extIds"),
            "ext_ids_encoding": request.get("extIdsEncoding"),
            "content": request.get("content"),
            "content_encoding": request.get("contentEncoding"),
            "schema": request.get("schema"),
            "payload": request.get("payload"),
            "ec_private_key": request.get("ecPrivateKey"),
        }
