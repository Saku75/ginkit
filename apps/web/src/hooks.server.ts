import { API_URL, APP_VERSION } from "$app/env/private";
import type { HandleServerError, ServerInit } from "@sveltejs/kit/hooks";
import { sequence } from "@sveltejs/kit/hooks";
import { client } from "#lib/api/client.gen.js";
import { headersServerHook } from "#lib/server/hooks/headers.js";
import { loggingServerHook } from "#lib/server/hooks/logging.js";
import { preloadServerHook } from "#lib/server/hooks/preload.js";
import { log } from "#lib/server/log.js";

// The server's API client talks to the API directly; the browser's keeps /api.
// Call it with its own fetch, never SvelteKit's event.fetch: that one adds an
// Origin header copied from the incoming request, and the API's CORS answers
// 403 when that origin is not allowed, which is what a probe reaching this
// app by IP looks like.
export const init: ServerInit = () => {
  client.setConfig({ baseUrl: API_URL });
  log.info("starting", { version: APP_VERSION, api: API_URL });
};

export const handle = sequence(loggingServerHook, headersServerHook, preloadServerHook);

// Runs for every error except redirects. Unknown errors are what a stack trace
// is for; app errors of 500 and up carry a message worth keeping; framework
// errors such as 404s are already in the request log. Returning nothing keeps
// SvelteKit's response to the client as it is.
export const handleError: HandleServerError = (input) => {
  const request = { method: input.event.request.method, path: input.event.url.pathname };
  switch (input.kind) {
    case "unknown":
      log.error("unhandled error", {
        request,
        error: input.error,
        stack: input.error instanceof Error ? input.error.stack : undefined,
      });
      break;
    case "app":
      if (input.error.status >= 500) {
        log.error("request failed", {
          request,
          status: input.error.status,
          error: input.error.message,
        });
      }
      break;
    case "validation":
      log.warn("invalid remote call", {
        request,
        remote: input.remote?.name,
        issues: input.issues.map((issue) => issue.message),
      });
      break;
  }
};
