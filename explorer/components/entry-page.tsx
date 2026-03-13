"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import CopyButton from "@/components/copy-button";
import { EntryDetail, ReceiptVerifyResponse, ExplorerAPIError, fetchEntry, fetchReceipt } from "@/lib/api";
import { useApiBase } from "@/lib/use-api-base";

type ReceiptState = "loading" | "available" | "pending" | "error";

function describeEntryError(error: Error) {
  if (error instanceof ExplorerAPIError && error.status === 404) {
    return "This entry is not available on the connected API. Check the hash or wait for indexing.";
  }

  return error.message || "Entry lookup failed.";
}

function classifyReceipt(error: Error): { state: ReceiptState; message: string } {
  if (error instanceof ExplorerAPIError && error.status === 404) {
    return {
      state: "pending",
      message: "The API does not have a receipt for this entry yet.",
    };
  }

  return {
    state: "error",
    message: error.message || "The API could not return receipt data.",
  };
}

function formatPayload(entry: EntryDetail) {
  return JSON.stringify(entry.decodedPayload ?? entry.content, null, 2);
}

export default function EntryPage({ entryHash }: { entryHash: string }) {
  const searchParams = useSearchParams();
  const { apiBase, ready } = useApiBase(searchParams.get("api") ?? undefined);
  const [entry, setEntry] = useState<EntryDetail | null>(null);
  const [receipt, setReceipt] = useState<ReceiptVerifyResponse | null>(null);
  const [entryError, setEntryError] = useState<string>("");
  const [receiptMessage, setReceiptMessage] = useState<string>("");
  const [receiptState, setReceiptState] = useState<ReceiptState>("loading");

  useEffect(() => {
    if (!ready) {
      return;
    }

    let cancelled = false;
    setEntry(null);
    setReceipt(null);
    setEntryError("");
    setReceiptMessage("");
    setReceiptState("loading");

    fetchEntry(entryHash, apiBase)
      .then((result) => {
        if (!cancelled) {
          setEntry(result);
        }
      })
      .catch((error: Error) => {
        if (!cancelled) {
          setEntryError(describeEntryError(error));
        }
      });

    fetchReceipt(entryHash, apiBase)
      .then((result) => {
        if (!cancelled) {
          setReceipt(result);
          setReceiptState("available");
          setReceiptMessage(result.message || "Receipt data returned by the connected API.");
        }
      })
      .catch((error: Error) => {
        if (!cancelled) {
          const next = classifyReceipt(error);
          setReceiptState(next.state);
          setReceiptMessage(next.message);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [apiBase, entryHash, ready]);

  const payload = entry ? formatPayload(entry) : "";
  const receiptClass =
    receiptState === "available"
      ? "status ok"
      : receiptState === "pending"
        ? "status pending"
        : receiptState === "loading"
          ? "status loading"
          : "status error";
  const receiptLabel =
    receiptState === "available"
      ? "Available"
      : receiptState === "pending"
        ? "Pending"
        : receiptState === "loading"
          ? "Checking"
          : "Unavailable";

  return (
    <main className="page">
      <div className="header-row">
        <div className="page-heading">
          <div className="eyebrow">Entry detail</div>
          <h1 className="page-title">{entryHash}</h1>
          <p>Connected API: {apiBase}</p>
        </div>
        <div className="actions align-start">
          <CopyButton value={entryHash} label="Copy entry hash" />
          <Link className="button-link secondary" href={`/?api=${encodeURIComponent(apiBase)}`}>
            Back home
          </Link>
          {entry?.chainId ? (
            <Link className="button-link secondary" href={`/chain/${encodeURIComponent(entry.chainId)}?api=${encodeURIComponent(apiBase)}`}>
              Open chain
            </Link>
          ) : null}
        </div>
      </div>

      <div className="grid two">
        <section className="panel">
          <div className="header-row tight">
            <div>
              <h2>Entry metadata</h2>
              <p className="small">Returned directly by `/entries/{'{entryHash}'}`.</p>
            </div>
          </div>

          {!entry && !entryError ? <p className="empty-copy">Loading entry metadata...</p> : null}
          {entryError ? <div className="notice error-box">{entryError}</div> : null}
          {entry ? (
            <div className="kv compact-kv">
              <div className="kv-row">
                <div className="label">Entry hash</div>
                <div className="inline-value-row">
                  <div className="value code">{entry.entryHash}</div>
                  <CopyButton value={entry.entryHash} label="Copy" />
                </div>
              </div>
              <div className="kv-row">
                <div className="label">Chain ID</div>
                <div className="inline-value-row">
                  <div className="value code">{entry.chainId}</div>
                  <CopyButton value={entry.chainId} label="Copy" />
                </div>
              </div>
              <div className="kv-row">
                <div className="label">Schema</div>
                <div className="value">{entry.schema || "Not available"}</div>
              </div>
              <div className="kv-row">
                <div className="label">Structured</div>
                <div className="value">{entry.structured ? "yes" : "no"}</div>
              </div>
              <div className="kv-row">
                <div className="label">Content encoding</div>
                <div className="value">{entry.contentEncoding}</div>
              </div>
            </div>
          ) : null}
        </section>

        <section className="panel">
          <div className="header-row tight">
            <div>
              <h2>Receipt status</h2>
              <p className="small">Read-only check against `/receipts/verify`.</p>
            </div>
            <span className={receiptClass}>{receiptLabel}</span>
          </div>

          {receiptState === "loading" ? <p className="empty-copy">Checking receipt availability...</p> : null}
          {receiptMessage ? (
            <div className={receiptState === "error" ? "notice error-box" : receiptState === "pending" ? "notice pending-box" : "notice success-box"}>
              {receiptMessage}
            </div>
          ) : null}
          {receipt?.receipt ? (
            <div className="kv compact-kv" style={{ marginTop: 16 }}>
              <div className="kv-row">
                <div className="label">Entry hash</div>
                <div className="value code">{receipt.entryHash}</div>
              </div>
              <div className="kv-row">
                <div className="label">Receipt data</div>
                <div className="value">Raw JSON is available below.</div>
              </div>
            </div>
          ) : null}
        </section>
      </div>

      {entry ? (
        <div className="grid two" style={{ marginTop: 18 }}>
          <section className="panel">
            <div className="header-row tight">
              <div>
                <h2>ExtIDs</h2>
                <p className="small">Base64 values returned by the API.</p>
              </div>
            </div>
            <div className="list">
              {entry.extIds.length === 0 ? <p className="empty-copy">No ExtIDs returned for this entry.</p> : null}
              {entry.extIds.map((extId, index) => (
                <div className="item" key={`${extId}-${index}`}>
                  <div className="label">ExtID {index}</div>
                  <div className="code">{extId}</div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel">
            <div className="header-row tight">
              <div>
                <h2>Payload</h2>
                <p className="small">Rendered exactly from the HTTP API response.</p>
              </div>
            </div>
            <pre>{payload}</pre>
          </section>
        </div>
      ) : null}

      {receipt?.receipt ? (
        <section className="panel" style={{ marginTop: 18 }}>
          <details className="details-block">
            <summary>Raw receipt JSON</summary>
            <pre style={{ marginTop: 16 }}>{JSON.stringify(receipt.receipt, null, 2)}</pre>
          </details>
        </section>
      ) : null}
    </main>
  );
}

