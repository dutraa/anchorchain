export const DEFAULT_API_BASE = "http://127.0.0.1:8081";
export const API_BASE_STORAGE_KEY = "anchorchain.apiBase";

export type APIErrorPayload = {
  error?: string;
  message?: string;
  status?: string;
};

export type HealthResponse = {
  status: string;
  heights: Record<string, number | string | null>;
};

export type ChainSummary = {
  chainId: string;
  entryCount?: number;
  latestEntryHash?: string;
  latestEntryTimestamp?: number;
};

export type ChainEntriesResponse = {
  chainId: string;
  entries: EntrySummary[];
  limit: number;
  offset: number;
  total: number;
};

export type EntrySummary = {
  entryHash: string;
  timestamp?: number;
  extIds: string[];
  schema?: string;
  structured: boolean;
};

export type EntryDetail = {
  entryHash: string;
  chainId: string;
  extIds: string[];
  schema?: string;
  structured: boolean;
  content: string;
  contentEncoding: string;
  decodedPayload?: unknown;
};

export type ReceiptVerifyResponse = {
  success: boolean;
  entryHash: string;
  receipt?: unknown;
  message?: string;
};

export class ExplorerAPIError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ExplorerAPIError";
    this.status = status;
  }
}

export function normalizeApiBase(input: string | null | undefined): string {
  const trimmed = (input ?? "").trim();
  if (!trimmed) {
    return DEFAULT_API_BASE;
  }

  if (/^https?:\/\//i.test(trimmed)) {
    return trimmed.replace(/\/$/, "");
  }

  return `http://${trimmed}`.replace(/\/$/, "");
}

function buildProxyURL(path: string, apiBase: string, query?: Record<string, string | number | undefined>) {
  const trimmedPath = path.replace(/^\/+/, "");
  const params = new URLSearchParams({ base: normalizeApiBase(apiBase) });

  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined && value !== "") {
        params.set(key, String(value));
      }
    }
  }

  return `/api/proxy/${trimmedPath}?${params.toString()}`;
}

export async function explorerFetch<T>(
  path: string,
  apiBase: string,
  init?: RequestInit,
  query?: Record<string, string | number | undefined>,
): Promise<T> {
  const response = await fetch(buildProxyURL(path, apiBase, query), {
    ...init,
    headers: {
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  });

  const contentType = response.headers.get("content-type") ?? "";
  const isJSON = contentType.includes("application/json");
  const payload = isJSON ? await response.json() : await response.text();

  if (!response.ok) {
    const errorPayload = (isJSON ? payload : {}) as APIErrorPayload;
    const fallback = typeof payload === "string" ? payload : response.statusText;
    const message = errorPayload.error || errorPayload.message || errorPayload.status || fallback || "Request failed";
    throw new ExplorerAPIError(response.status, message);
  }

  return payload as T;
}

export async function fetchHealth(apiBase: string) {
  return explorerFetch<HealthResponse>("health", apiBase);
}

export async function fetchChain(chainId: string, apiBase: string) {
  return explorerFetch<ChainSummary>(`chains/${encodeURIComponent(chainId)}`, apiBase);
}

export async function fetchChainEntries(chainId: string, apiBase: string, limit = 25) {
  return explorerFetch<ChainEntriesResponse>(`chains/${encodeURIComponent(chainId)}/entries`, apiBase, undefined, { limit });
}

export async function fetchEntry(entryHash: string, apiBase: string) {
  return explorerFetch<EntryDetail>(`entries/${encodeURIComponent(entryHash)}`, apiBase);
}

export async function fetchReceipt(entryHash: string, apiBase: string) {
  return explorerFetch<ReceiptVerifyResponse>(`receipts/verify`, apiBase, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ entryHash, includeRawEntry: false }),
  });
}

export function formatTimestamp(timestamp?: number) {
  if (!timestamp) {
    return "Not available";
  }

  return new Date(timestamp * 1000).toLocaleString();
}

export function formatMetricKey(key: string) {
  const labels: Record<string, string> = {
    directoryblockheight: "Directory block",
    leaderheight: "Leader",
    entryblockheight: "Entry block",
    entryheight: "Entry",
  };

  return labels[key.toLowerCase()] ?? key;
}

export function formatMetricValue(value: number | string | null | undefined) {
  if (value === null || value === undefined || value === "") {
    return "Not available";
  }

  if (typeof value === "number") {
    return value.toLocaleString();
  }

  return String(value);
}
