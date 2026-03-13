"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

import CopyButton from "@/components/copy-button";
import { ChainEntriesResponse, ChainSummary, ExplorerAPIError, fetchChain, fetchChainEntries, formatTimestamp } from "@/lib/api";
import { useApiBase } from "@/lib/use-api-base";

function describeChainError(error: Error) {
  if (error instanceof ExplorerAPIError && error.status === 404) {
    return "This chain is not available on the connected API yet. It may be missing or still indexing.";
  }

  return error.message || "Chain lookup failed.";
}

function describeEntriesError(error: Error) {
  if (error instanceof ExplorerAPIError && error.status === 404) {
    return "Recent entries are not available for this chain yet.";
  }

  return error.message || "Entry listing failed.";
}

export default function ChainPage({ chainId }: { chainId: string }) {
  const searchParams = useSearchParams();
  const { apiBase, ready } = useApiBase(searchParams.get("api") ?? undefined);
  const [chain, setChain] = useState<ChainSummary | null>(null);
  const [entries, setEntries] = useState<ChainEntriesResponse | null>(null);
  const [chainError, setChainError] = useState<string>("");
  const [entriesError, setEntriesError] = useState<string>("");

  useEffect(() => {
    if (!ready) {
      return;
    }

    let cancelled = false;
    setChain(null);
    setEntries(null);
    setChainError("");
    setEntriesError("");

    fetchChain(chainId, apiBase)
      .then((result) => {
        if (!cancelled) {
          setChain(result);
        }
      })
      .catch((error: Error) => {
        if (!cancelled) {
          setChainError(describeChainError(error));
        }
      });

    fetchChainEntries(chainId, apiBase, 25)
      .then((result) => {
        if (!cancelled) {
          setEntries(result);
        }
      })
      .catch((error: Error) => {
        if (!cancelled) {
          setEntriesError(describeEntriesError(error));
        }
      });

    return () => {
      cancelled = true;
    };
  }, [apiBase, chainId, ready]);

  return (
    <main className="page">
      <div className="header-row">
        <div className="page-heading">
          <div className="eyebrow">Chain detail</div>
          <h1 className="page-title">{chainId}</h1>
          <p>Connected API: {apiBase}</p>
        </div>
        <div className="actions align-start">
          <CopyButton value={chainId} label="Copy chain ID" />
          <Link className="button-link secondary" href={`/?api=${encodeURIComponent(apiBase)}`}>
            Back home
          </Link>
        </div>
      </div>

      <div className="grid two">
        <section className="panel">
          <div className="header-row tight">
            <div>
              <h2>Chain metadata</h2>
              <p className="small">Summary returned by the AnchorChain HTTP API.</p>
            </div>
          </div>

          {!chain && !chainError ? <p className="empty-copy">Loading chain summary...</p> : null}
          {chainError ? <div className="notice error-box">{chainError}</div> : null}
          {chain ? (
            <div className="kv compact-kv">
              <div className="kv-row">
                <div className="label">Chain ID</div>
                <div className="inline-value-row">
                  <div className="value code">{chain.chainId}</div>
                  <CopyButton value={chain.chainId} label="Copy" />
                </div>
              </div>
              <div className="kv-row">
                <div className="label">Entry count</div>
                <div className="value strong-value">{chain.entryCount ?? "Not available"}</div>
              </div>
              <div className="kv-row">
                <div className="label">Latest entry hash</div>
                <div className="inline-value-row">
                  <div className="value code">{chain.latestEntryHash ?? "Not available"}</div>
                  {chain.latestEntryHash ? <CopyButton value={chain.latestEntryHash} label="Copy" /> : null}
                </div>
              </div>
              <div className="kv-row">
                <div className="label">Latest entry timestamp</div>
                <div className="value">{formatTimestamp(chain.latestEntryTimestamp)}</div>
              </div>
            </div>
          ) : null}
        </section>

        <section className="panel">
          <div className="header-row tight">
            <div>
              <h2>Recent entries</h2>
              <p className="small">Showing up to 25 entries from `/chains/{'{chainId}'}/entries`.</p>
            </div>
          </div>

          {!entries && !entriesError ? <p className="empty-copy">Loading recent entries...</p> : null}
          {entriesError ? <div className="notice error-box">{entriesError}</div> : null}
          {entries ? (
            <div className="list">
              {entries.entries.length === 0 ? <p className="empty-copy">No entries were returned for this chain.</p> : null}
              {entries.entries.map((entry) => (
                <Link
                  className="entry-card"
                  key={entry.entryHash}
                  href={`/entry/${encodeURIComponent(entry.entryHash)}?api=${encodeURIComponent(apiBase)}`}
                >
                  <div className="entry-card-top">
                    <span className="entry-hash code">{entry.entryHash}</span>
                    <span className="muted small">{formatTimestamp(entry.timestamp)}</span>
                  </div>
                  <div className="entry-card-meta">
                    <span className="pill neutral">Structured: {entry.structured ? "yes" : "no"}</span>
                    <span className="pill neutral">Schema: {entry.schema || "Not available"}</span>
                    <span className="pill neutral">ExtIDs: {entry.extIds.length}</span>
                  </div>
                  <span className="small accent-link">Open entry detail</span>
                </Link>
              ))}
            </div>
          ) : null}
        </section>
      </div>
    </main>
  );
}
