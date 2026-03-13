"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";

import { fetchHealth, formatMetricKey, formatMetricValue, type HealthResponse } from "@/lib/api";
import { useApiBase } from "@/lib/use-api-base";

type HealthState = "loading" | "healthy" | "degraded";

export default function HomePage() {
  const router = useRouter();
  const { apiBase, updateApiBase, ready } = useApiBase();
  const [draftApiBase, setDraftApiBase] = useState(apiBase);
  const [chainId, setChainId] = useState("");
  const [entryHash, setEntryHash] = useState("");
  const [healthData, setHealthData] = useState<HealthResponse | null>(null);
  const [healthError, setHealthError] = useState<string>("");
  const [healthState, setHealthState] = useState<HealthState>("loading");

  useEffect(() => {
    setDraftApiBase(apiBase);
  }, [apiBase]);

  useEffect(() => {
    if (!ready) {
      return;
    }

    let cancelled = false;
    setHealthData(null);
    setHealthError("");
    setHealthState("loading");

    fetchHealth(apiBase)
      .then((result) => {
        if (cancelled) {
          return;
        }
        setHealthData(result);
        setHealthState(result.status === "ok" ? "healthy" : "degraded");
      })
      .catch((error: Error) => {
        if (cancelled) {
          return;
        }
        setHealthError(error.message || "Unable to reach API.");
        setHealthState("degraded");
      });

    return () => {
      cancelled = true;
    };
  }, [apiBase, ready]);

  const metrics = (() => {
    if (!healthData?.heights) {
      return [];
    }

    const order = ["directoryblockheight", "leaderheight", "entryblockheight", "entryheight"];
    const entries = Object.entries(healthData.heights);

    return entries.sort(([left], [right]) => {
      const leftIndex = order.indexOf(left.toLowerCase());
      const rightIndex = order.indexOf(right.toLowerCase());
      const normalizedLeft = leftIndex === -1 ? Number.MAX_SAFE_INTEGER : leftIndex;
      const normalizedRight = rightIndex === -1 ? Number.MAX_SAFE_INTEGER : rightIndex;
      return normalizedLeft - normalizedRight || left.localeCompare(right);
    });
  })();

  function handleSaveApiBase(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    updateApiBase(draftApiBase);
  }

  function handleChainSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextChainId = chainId.trim();
    if (!nextChainId) {
      return;
    }
    router.push(`/chain/${encodeURIComponent(nextChainId)}?api=${encodeURIComponent(apiBase)}`);
  }

  function handleEntrySearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextEntryHash = entryHash.trim();
    if (!nextEntryHash) {
      return;
    }
    router.push(`/entry/${encodeURIComponent(nextEntryHash)}?api=${encodeURIComponent(apiBase)}`);
  }

  const statusLabel =
    healthState === "healthy" ? "Connected" : healthState === "loading" ? "Checking" : "Unreachable";
  const statusClass =
    healthState === "healthy" ? "status ok" : healthState === "loading" ? "status loading" : "status error";

  return (
    <main className="page">
      <section className="hero">
        <div className="eyebrow">AnchorChain Explorer</div>
        <h1>Read-only inspection for chains, entries, and node status.</h1>
        <p className="lede">
          Use the existing AnchorChain HTTP API to inspect chain summaries, entry payloads, and receipt
          availability. This explorer does not write or modify chain data.
        </p>
      </section>

      <div className="grid two">
        <section className="panel">
          <div className="header-row tight">
            <div>
              <h2>API Connection</h2>
              <p className="small">The selected API base URL is saved in your browser for this explorer.</p>
            </div>
            <span className={statusClass}>{statusLabel}</span>
          </div>

          <form className="form" onSubmit={handleSaveApiBase}>
            <div className="field">
              <label className="label" htmlFor="api-base-url">
                API base URL
              </label>
              <input
                id="api-base-url"
                value={draftApiBase}
                onChange={(event) => setDraftApiBase(event.target.value)}
                placeholder="http://127.0.0.1:8081"
              />
            </div>
            <div className="actions">
              <button type="submit">Save API Base URL</button>
            </div>
          </form>

          <div className="subpanel" style={{ marginTop: 18 }}>
            <div className="header-row tight">
              <div>
                <div className="label">Node health</div>
                <p className="small">Read from the connected API&apos;s `/health` endpoint.</p>
              </div>
              <span className={statusClass}>{statusLabel}</span>
            </div>

            {healthState === "loading" ? <p className="empty-copy">Loading health data...</p> : null}
            {healthError ? <div className="notice error-box">{healthError}</div> : null}
            {!healthError && metrics.length > 0 ? (
              <div className="metric-grid">
                {metrics.map(([key, value]) => (
                  <div className="metric" key={key}>
                    <div className="label">{formatMetricKey(key)}</div>
                    <div className="metric-value">{formatMetricValue(value)}</div>
                  </div>
                ))}
              </div>
            ) : null}
          </div>
        </section>

        <section className="panel">
          <div className="header-row tight">
            <div>
              <h2>Open Chain or Entry</h2>
              <p className="small">Paste a chain ID or entry hash from the connected AnchorChain API.</p>
            </div>
          </div>

          <div className="stack-lg">
            <form className="form" onSubmit={handleChainSearch}>
              <div className="field">
                <label className="label" htmlFor="chain-id-input">
                  Chain ID
                </label>
                <input
                  id="chain-id-input"
                  value={chainId}
                  onChange={(event) => setChainId(event.target.value)}
                  placeholder="Enter a chain ID"
                />
              </div>
              <div className="actions">
                <button type="submit">Open Chain Detail</button>
              </div>
            </form>

            <form className="form" onSubmit={handleEntrySearch}>
              <div className="field">
                <label className="label" htmlFor="entry-hash-input">
                  Entry hash
                </label>
                <input
                  id="entry-hash-input"
                  value={entryHash}
                  onChange={(event) => setEntryHash(event.target.value)}
                  placeholder="Enter an entry hash"
                />
              </div>
              <div className="actions">
                <button type="submit">Open Entry Detail</button>
              </div>
            </form>
          </div>
        </section>
      </div>
    </main>
  );
}

