export type APIErrorBody = {
  error?: string;
  message?: string;
  status?: string;
};

export type HealthResponse = {
  status: string;
  heights: Record<string, number | string | null>;
};

export type WriteResponse = {
  success: boolean;
  status: string;
  chainId?: string;
  entryHash?: string;
  txId?: string;
  message?: string;
  schema?: string;
  structured?: boolean;
};

export type CreateChainRequest = {
  extIds?: string[];
  extIdsEncoding?: string;
  content?: string;
  contentEncoding?: string;
  schema?: string;
  payload?: unknown;
  ecPrivateKey?: string;
};

export type WriteEntryRequest = {
  extIds?: string[];
  extIdsEncoding?: string;
  content?: string;
  contentEncoding?: string;
  schema?: string;
  payload?: unknown;
  ecPrivateKey?: string;
};

export type ChainResponse = {
  chainId: string;
  entryCount?: number;
  latestEntryHash?: string;
  latestEntryTimestamp?: number;
};

export type EntrySummary = {
  entryHash: string;
  timestamp?: number;
  extIds: string[];
  schema?: string;
  structured: boolean;
};

export type EntriesResponse = {
  chainId: string;
  entries: EntrySummary[];
  limit: number;
  offset: number;
  total: number;
};

export type EntryResponse = {
  entryHash: string;
  chainId: string;
  extIds: string[];
  schema?: string;
  structured: boolean;
  content: string;
  contentEncoding: string;
  decodedPayload?: unknown;
};

export type VerifyReceiptRequest = {
  entryHash: string;
  includeRawEntry?: boolean;
};

export type VerifyReceiptResponse = {
  success: boolean;
  entryHash: string;
  receipt?: unknown;
  message?: string;
};

export type GetEntriesOptions = {
  limit?: number;
  offset?: number;
};

export type AnchorChainClientOptions = {
  baseUrl?: string;
  token?: string;
  fetch?: typeof fetch;
};

export type AnchorChainErrorKind =
  | "api_unreachable"
  | "missing_chain_head"
  | "receipt_not_ready"
  | "receipt_creation_error"
  | "bad_request"
  | "server_error"
  | "unknown";

export class AnchorChainSDKError extends Error {
  status?: number;
  body?: unknown;
  kind: AnchorChainErrorKind;
  retriable: boolean;
  cause?: unknown;

  constructor(
    message: string,
    options: {
      status?: number;
      body?: unknown;
      kind: AnchorChainErrorKind;
      retriable: boolean;
      cause?: unknown;
    },
  ) {
    super(message);
    this.name = "AnchorChainSDKError";
    this.status = options.status;
    this.body = options.body;
    this.kind = options.kind;
    this.retriable = options.retriable;
    this.cause = options.cause;
  }
}

function normalizeBaseUrl(baseUrl: string | undefined) {
  const value = (baseUrl ?? "http://127.0.0.1:8081").trim();
  if (!value) {
    return "http://127.0.0.1:8081";
  }
  if (/^https?:\/\//i.test(value)) {
    return value.replace(/\/$/, "");
  }
  return `http://${value}`.replace(/\/$/, "");
}

function resolveMessage(body: unknown, fallback: string) {
  if (!body || typeof body !== "object") {
    return fallback;
  }

  const apiBody = body as APIErrorBody;
  return apiBody.error || apiBody.message || apiBody.status || fallback;
}

function normalizeMessage(message: string) {
  return message.trim().toLowerCase();
}

function classifyAPIError(status: number, message: string) {
  const normalized = normalizeMessage(message);

  if (normalized.includes("missing chain head")) {
    return {
      kind: "missing_chain_head" as const,
      retriable: status === 404 || status === 202,
    };
  }

  if (normalized.includes("receipt not available")) {
    return {
      kind: "receipt_not_ready" as const,
      retriable: status === 404,
    };
  }

  if (normalized.includes("receipt creation error")) {
    return {
      kind: "receipt_creation_error" as const,
      retriable: status >= 500,
    };
  }

  if (status === 400) {
    return {
      kind: "bad_request" as const,
      retriable: false,
    };
  }

  if (status >= 500) {
    return {
      kind: "server_error" as const,
      retriable: true,
    };
  }

  return {
    kind: "unknown" as const,
    retriable: false,
  };
}

function classifyTransportError(error: unknown) {
  const message = error instanceof Error ? error.message : "Unable to reach AnchorChain API";
  return new AnchorChainSDKError(message, {
    kind: "api_unreachable",
    retriable: true,
    cause: error,
  });
}

export class AnchorChainClient {
  readonly baseUrl: string;
  readonly token?: string;
  private readonly fetchImpl: typeof fetch;

  constructor(options: AnchorChainClientOptions = {}) {
    if (!options.fetch && typeof globalThis.fetch !== "function") {
      throw new Error("fetch is not available in this runtime. Use Node 18+ or provide a fetch implementation.");
    }

    this.baseUrl = normalizeBaseUrl(options.baseUrl);
    this.token = options.token;
    this.fetchImpl = options.fetch ?? globalThis.fetch;
  }

  async health() {
    return this.request<HealthResponse>("GET", "/health");
  }

  async createChain(request: CreateChainRequest) {
    return this.request<WriteResponse>("POST", "/chains", request);
  }

  async writeEntry(chainId: string, request: WriteEntryRequest) {
    const encodedChainId = encodeURIComponent(chainId.trim());
    return this.request<WriteResponse>("POST", `/chains/${encodedChainId}/entries`, request);
  }

  async getChain(chainId: string) {
    const encodedChainId = encodeURIComponent(chainId.trim());
    return this.request<ChainResponse>("GET", `/chains/${encodedChainId}`);
  }

  async getEntries(chainId: string, options: GetEntriesOptions = {}) {
    const encodedChainId = encodeURIComponent(chainId.trim());
    const params = new URLSearchParams();
    if (options.limit !== undefined) {
      params.set("limit", String(options.limit));
    }
    if (options.offset !== undefined) {
      params.set("offset", String(options.offset));
    }

    const suffix = params.size > 0 ? `?${params.toString()}` : "";
    return this.request<EntriesResponse>("GET", `/chains/${encodedChainId}/entries${suffix}`);
  }

  async getEntry(entryHash: string) {
    const encodedEntryHash = encodeURIComponent(entryHash.trim());
    return this.request<EntryResponse>("GET", `/entries/${encodedEntryHash}`);
  }

  async verifyReceipt(request: VerifyReceiptRequest | string) {
    const body = typeof request === "string" ? { entryHash: request } : request;
    return this.request<VerifyReceiptResponse>("POST", "/receipts/verify", body);
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    let response: Response;
    try {
      response = await this.fetchImpl(`${this.baseUrl}${path}`, {
        method,
        headers: this.buildHeaders(body !== undefined),
        body: body !== undefined ? JSON.stringify(body) : undefined,
      });
    } catch (error) {
      throw classifyTransportError(error);
    }

    const contentType = response.headers.get("content-type") ?? "";
    const isJSON = contentType.includes("application/json");
    const responseBody = isJSON ? await response.json() : await response.text();

    if (!response.ok) {
      const message = resolveMessage(responseBody, `AnchorChain API request failed with status ${response.status}`);
      const classification = classifyAPIError(response.status, message);
      throw new AnchorChainSDKError(
        message,
        {
          status: response.status,
          body: responseBody,
          kind: classification.kind,
          retriable: classification.retriable,
        },
      );
    }

    return responseBody as T;
  }

  private buildHeaders(hasBody: boolean) {
    const headers: Record<string, string> = {};

    if (hasBody) {
      headers["Content-Type"] = "application/json";
    }

    if (this.token) {
      headers["X-Anchorchain-Api-Token"] = this.token;
    }

    return headers;
  }
}
