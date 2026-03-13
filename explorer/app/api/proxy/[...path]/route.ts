import { NextRequest } from "next/server";

function joinPath(basePath: string, suffix: string) {
  const left = basePath.replace(/\/$/, "");
  const right = suffix.replace(/^\//, "");
  return `${left}/${right}`;
}

function parseBase(rawBase: string | null) {
  if (!rawBase) {
    throw new Error("Missing API base URL.");
  }

  const base = new URL(rawBase);
  if (base.protocol !== "http:" && base.protocol !== "https:") {
    throw new Error("API base URL must use http or https.");
  }

  return base;
}

async function forward(request: NextRequest, path: string[]) {
  try {
    const incoming = new URL(request.url);
    const base = parseBase(incoming.searchParams.get("base"));

    const target = new URL(joinPath(base.pathname || "/", path.join("/")), base);
    incoming.searchParams.forEach((value, key) => {
      if (key !== "base") {
        target.searchParams.set(key, value);
      }
    });

    const headers = new Headers();
    const contentType = request.headers.get("content-type");
    if (contentType) {
      headers.set("content-type", contentType);
    }

    const init: RequestInit = {
      method: request.method,
      headers,
      cache: "no-store",
    };

    if (request.method !== "GET" && request.method !== "HEAD") {
      init.body = await request.text();
    }

    const upstream = await fetch(target, init);
    const responseHeaders = new Headers();
    const upstreamType = upstream.headers.get("content-type");
    if (upstreamType) {
      responseHeaders.set("content-type", upstreamType);
    }

    return new Response(await upstream.text(), {
      status: upstream.status,
      headers: responseHeaders,
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unable to reach AnchorChain API.";
    return Response.json({ error: message }, { status: 400 });
  }
}

export async function GET(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  const { path } = await context.params;
  return forward(request, path);
}

export async function POST(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  const { path } = await context.params;
  return forward(request, path);
}
