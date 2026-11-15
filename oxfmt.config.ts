import { defineConfig } from "oxfmt";

export default defineConfig({
  svelte: true,
  ignorePatterns: [
    "**/node_modules/**",
    "**/.svelte-kit/**",
    "**/build/**",
    "**/.turbo/**",
    "**/src/lib/api/**",
    "apps/api/openapi.yaml",
    "apps/api/bin/**",
    "apps/api/tmp/**",
    "apps/api/internal/handlers/docs/assets/**",
    "pnpm-lock.yaml",
  ],
});
