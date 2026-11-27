import { defineConfig } from "oxlint";

export default defineConfig({
  plugins: ["typescript", "unicorn", "oxc"],
  categories: {
    correctness: "error",
    suspicious: "warn",
    pedantic: "warn",
    perf: "warn",
  },
  rules: {
    // Both fire on shapes other packages' types fix: parameters that cannot be
    // made readonly, and the `export {}` that makes src/app.d.ts a module.
    "typescript/prefer-readonly-parameter-types": "off",
    "unicorn/require-module-specifiers": "off",
  },
  options: {
    typeAware: true,
    typeCheck: true,
  },
  env: {
    builtin: true,
  },
  ignorePatterns: ["**/src/lib/api/**", "**/.svelte-kit/**", "**/build/**"],
});
