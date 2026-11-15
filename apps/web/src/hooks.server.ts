import { API_URL } from "$app/env/private";
import type { ServerInit } from "@sveltejs/kit/hooks";
import { sequence } from "@sveltejs/kit/hooks";
import { client } from "#lib/api/client.gen.js";
import { headersServerHook } from "#lib/server/hooks/headers.js";
import { preloadServerHook } from "#lib/server/hooks/preload.js";

// The server's API client talks to the API directly; the browser's keeps /api.
export const init: ServerInit = () => {
  client.setConfig({ baseUrl: API_URL });
};

export const handle = sequence(headersServerHook, preloadServerHook);
