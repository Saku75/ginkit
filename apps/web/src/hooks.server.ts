import { API_URL } from "$app/env/private";
import type { ServerInit } from "@sveltejs/kit/hooks";
import { sequence } from "@sveltejs/kit/hooks";
import { client } from "#lib/api/client.gen.js";
import { headersServerHook } from "#lib/server/hooks/headers.js";
import { preloadServerHook } from "#lib/server/hooks/preload.js";

// The server's API client talks to the API directly; the browser's keeps /api.
// Call it with its own fetch, never SvelteKit's event.fetch: that one adds an
// Origin header copied from the incoming request, and the API's CORS answers
// 403 when that origin is not allowed, which is what a probe reaching this
// app by IP looks like.
export const init: ServerInit = () => {
  client.setConfig({ baseUrl: API_URL });
};

export const handle = sequence(headersServerHook, preloadServerHook);
