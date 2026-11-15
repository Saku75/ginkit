import { createClient, type UserConfig as OpenApiConfig } from "@hey-api/openapi-ts";
import { heyApiPlugin } from "@hey-api/vite-plugin";
import adapter from "@sveltejs/adapter-node";
import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { createRequire } from "node:module";
import { defineConfig, type Plugin } from "vite";

// The client in src/lib/api is generated from the API package's committed OpenAPI document.
const openapi: OpenApiConfig = {
  input: createRequire(import.meta.url).resolve("@ginkit/api/openapi.yaml"),
  output: "src/lib/api",
  plugins: [
    "@hey-api/typescript",
    "@hey-api/sdk",
    { name: "@hey-api/client-fetch", runtimeConfigPath: "./src/lib/api-config.ts" },
  ],
};

// Regenerates the client whenever the document changes while the dev server runs.
function openapiWatch(config: OpenApiConfig): Plugin {
  const input = typeof config.input === "string" ? config.input : undefined;
  return {
    name: "openapi-watch",
    apply: "serve",
    configureServer(server) {
      if (!input) return;
      server.watcher.add(input);
      server.watcher.on("change", async (file) => {
        if (file !== input) return;
        await createClient(config);
        server.config.logger.info("regenerated src/lib/api", { timestamp: true });
      });
    },
  };
}

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit({
      compilerOptions: {
        // Force runes mode for the project, except for libraries. Can be removed in svelte 6.
        runes: ({ filename }) =>
          filename.split(/[/\\]/).includes("node_modules") ? undefined : true,
      },
      adapter: adapter(),
    }),
    heyApiPlugin({ config: openapi }),
    openapiWatch(openapi),
  ],
  server: {
    port: 3100,
    strictPort: true,
    // /api is stripped before reaching the API on 3101, the same as a reverse proxy would do.
    proxy: {
      "/api": {
        target: "http://localhost:3101",
        rewrite: (path) => path.replace(/^\/api(?=\/|$)/, "") || "/",
        configure: (proxy) =>
          proxy.on("proxyReq", (request) => request.setHeader("X-Forwarded-Prefix", "/api")),
      },
    },
  },
});
