import type { CreateClientConfig } from "./api/client.gen";

// /api on the page's own origin, proxied to the API. hooks.server.ts points the
// server's client at API_URL instead.
export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: "/api",
});
