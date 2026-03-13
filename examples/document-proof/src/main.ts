import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import path from "node:path";

import { AnchorChainClient, AnchorChainSDKError, type VerifyReceiptResponse, type WriteResponse } from "@anchorchain/sdk-js";

type Options = {
  filePath: string;
  api: string;
  token?: string;
  chainId?: string;
  ecKey?: string;
  attempts: number;
  delayMs: number;
};

function parseArgs(argv: string[]): Options {
  if (argv.length === 0) {
    throw new Error("Usage: npm run demo -- <file> [--api URL] [--chain CHAIN_ID] [--token TOKEN] [--ec-key KEY]");
  }

  const options: Options = {
    filePath: "",
    api: process.env.ANCHORCHAIN_API ?? "http://127.0.0.1:8081",
    token: process.env.ANCHORCHAIN_API_TOKEN,
    ecKey: process.env.ANCHORCHAIN_EC_PRIVATE,
    attempts: 24,
    delayMs: 5000,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];

    switch (arg) {
      case "--api":
        options.api = argv[++index] ?? options.api;
        break;
      case "--token":
        options.token = argv[++index] ?? options.token;
        break;
      case "--chain":
        options.chainId = argv[++index];
        break;
      case "--ec-key":
        options.ecKey = argv[++index] ?? options.ecKey;
        break;
      case "--attempts":
        options.attempts = Number(argv[++index] ?? options.attempts);
        break;
      case "--delay-ms":
        options.delayMs = Number(argv[++index] ?? options.delayMs);
        break;
      default:
        if (arg.startsWith("--")) {
          throw new Error(`Unknown option: ${arg}`);
        }
        if (!options.filePath) {
          options.filePath = arg;
        }
        break;
    }
  }

  if (!options.filePath) {
    throw new Error("A file path is required.");
  }

  return options;
}

function sha256Hex(input: Buffer) {
  return createHash("sha256").update(input).digest("hex");
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function describeError(error: unknown) {
  if (error instanceof AnchorChainSDKError) {
    const suffix = error.status ? ` (status ${error.status})` : "";
    return `${error.kind}${suffix}: ${error.message}`;
  }

  return error instanceof Error ? error.message : String(error);
}

async function waitForReceipt(
  client: AnchorChainClient,
  entryHash: string,
  attempts: number,
  delayMs: number,
): Promise<VerifyReceiptResponse> {
  let lastError: unknown;

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      return await client.verifyReceipt({ entryHash, includeRawEntry: false });
    } catch (error) {
      lastError = error;
      const retriable = !(error instanceof AnchorChainSDKError) || error.retriable;
      if (attempt === attempts) {
        break;
      }
      if (!retriable) {
        break;
      }
      console.log(`[pending] Receipt not available yet; retrying (${attempt}/${attempts}): ${describeError(error)}`);
      await sleep(delayMs);
    }
  }

  throw lastError;
}

function buildProofPayload(filePath: string, fileBytes: Buffer, hash: string) {
  return {
    type: "document-proof/v1",
    algorithm: "sha256",
    fileName: path.basename(filePath),
    fileSizeBytes: fileBytes.byteLength,
    fileSha256: hash,
    hashedAt: new Date().toISOString(),
    note: "Only the hash and metadata are anchored on-chain. The file contents stay local.",
  };
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const absolutePath = path.resolve(options.filePath);
  const fileBytes = await readFile(absolutePath);
  const fileSha256 = sha256Hex(fileBytes);
  const payload = buildProofPayload(absolutePath, fileBytes, fileSha256);

  console.log("[ok] Local document hash:");
  console.log(JSON.stringify({ file: absolutePath, sha256: fileSha256, sizeBytes: fileBytes.byteLength }, null, 2));

  const client = new AnchorChainClient({
    baseUrl: options.api,
    token: options.token,
  });

  const health = await client.health();
  console.log("[ok] AnchorChain API health:");
  console.log(JSON.stringify(health, null, 2));

  let writeResult: WriteResponse;
  let chainId = options.chainId;

  if (chainId) {
    console.log(`Anchoring proof into existing chain ${chainId}...`);
    writeResult = await client.writeEntry(chainId, {
      schema: "json",
      payload,
      ecPrivateKey: options.ecKey,
    });
  } else {
    console.log("[pending] Creating a new proof chain and anchoring the first proof entry...");
    writeResult = await client.createChain({
      extIds: ["document-proof", "sha256"],
      extIdsEncoding: "utf-8",
      schema: "json",
      payload,
      ecPrivateKey: options.ecKey,
    });
    chainId = writeResult.chainId;
  }

  if (!writeResult.entryHash) {
    throw new Error("AnchorChain did not return an entry hash.");
  }

  console.log("[ok] Anchor result:");
  console.log(JSON.stringify(writeResult, null, 2));

  console.log("[pending] Waiting for receipt data from the API...");
  const receipt = await waitForReceipt(client, writeResult.entryHash, options.attempts, options.delayMs);
  console.log("[ok] Verified receipt:");
  console.log(JSON.stringify(receipt, null, 2));

  console.log("[ok] Summary:");
  console.log(
    JSON.stringify(
      {
        chainId,
        entryHash: writeResult.entryHash,
        fileSha256,
      },
      null,
      2,
    ),
  );
}

main().catch((error) => {
  if (error instanceof AnchorChainSDKError) {
    if (error.retriable) {
      console.error(`[timeout] Timed out waiting for local devnet readiness: ${describeError(error)}`);
      console.error("Writes may succeed before reads or receipts become available on local devnet.");
    } else {
      console.error(`[error] SDK error: ${describeError(error)}`);
    }
    if (error.body !== undefined) {
      console.error(JSON.stringify(error.body, null, 2));
    }
  } else if (error instanceof Error) {
    console.error(`[error] ${error.message}`);
  } else {
    console.error("[error]", error);
  }
  process.exit(1);
});
