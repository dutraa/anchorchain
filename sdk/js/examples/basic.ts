import { AnchorChainClient, AnchorChainSDKError } from "../src/index.ts";

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

function describeError(error: unknown) {
  if (error instanceof AnchorChainSDKError) {
    const suffix = error.status ? ` (status ${error.status})` : "";
    return `${error.kind}${suffix}: ${error.message}`;
  }
  return error instanceof Error ? error.message : String(error);
}

async function retry<T>(label: string, fn: () => Promise<T>, options?: { attempts?: number; delayMs?: number }) {
  const attempts = options?.attempts ?? 24;
  const delayMs = options?.delayMs ?? 5000;
  let lastError: unknown;

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      const retriable = !(error instanceof AnchorChainSDKError) || error.retriable;
      if (attempt === attempts) {
        break;
      }
      if (!retriable) {
        break;
      }

      console.log(`[pending] ${label} not ready yet; retrying (${attempt}/${attempts}): ${describeError(error)}`);
      await sleep(delayMs);
    }
  }

  throw lastError;
}

async function waitForChain(client: AnchorChainClient, chainId: string) {
  return retry("Chain", async () => client.getChain(chainId));
}

async function waitForEntry(client: AnchorChainClient, entryHash: string) {
  return retry("Entry", async () => client.getEntry(entryHash));
}

async function waitForReceipt(client: AnchorChainClient, entryHash: string) {
  return retry("Receipt", async () => client.verifyReceipt({ entryHash, includeRawEntry: false }));
}

async function main() {
  const client = new AnchorChainClient({
    baseUrl: process.env.ANCHORCHAIN_API ?? "http://127.0.0.1:8081",
    token: process.env.ANCHORCHAIN_API_TOKEN,
  });

  const nonce = new Date().toISOString();

  const health = await client.health();
  console.log("[ok] Health:", health);

  const created = await client.createChain({
    extIds: ["sdk", "example", nonce],
    extIdsEncoding: "utf-8",
    schema: "json",
    payload: { createdBy: "sdk/js", step: 1, nonce },
  });
  console.log("[ok] Create chain:", created);

  if (!created.chainId || !created.entryHash) {
    throw new Error("Chain creation did not return chainId and entryHash.");
  }

  const written = await client.writeEntry(created.chainId, {
    schema: "json",
    payload: { step: 2, note: "follow-up entry", nonce },
  });
  console.log("[ok] Write entry:", written);

  const chain = await waitForChain(client, created.chainId);
  console.log("[ok] Chain:", chain);

  const entries = await client.getEntries(created.chainId, { limit: 10, offset: 0 });
  console.log("[ok] Entries:", entries);

  const entryHash = written.entryHash ?? created.entryHash;
  const entry = await waitForEntry(client, entryHash);
  console.log("[ok] Entry:", entry);

  const receipt = await waitForReceipt(client, entryHash);
  console.log("[ok] Receipt:", receipt);
}

main().catch((error) => {
  if (error instanceof AnchorChainSDKError) {
    if (error.retriable) {
      console.error(`[timeout] Timed out waiting for a devnet-visible state: ${describeError(error)}`);
    } else {
      console.error(`[error] SDK error: ${describeError(error)}`);
    }
    console.error(error.body);
  } else {
    console.error("[error]", error);
  }
  process.exit(1);
});
