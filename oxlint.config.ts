import { defineConfig } from "oxlint";

export default defineConfig({
  ignorePatterns: ["**/src/lib/api/**", "**/.svelte-kit/**", "**/build/**"],
});
