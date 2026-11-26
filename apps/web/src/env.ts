import { defineEnvVars } from "@sveltejs/kit/env";
import * as v from "valibot";

export const variables = defineEnvVars({
  API_URL: {
    schema: v.optional(v.pipe(v.string(), v.url()), "http://localhost:3101"),
    description:
      "Where server-side rendering reaches the API. The browser uses /api on its own origin.",
  },
  LOG_LEVEL: {
    schema: v.optional(v.picklist(["debug", "info", "warn", "error"]), "info"),
    description: "Lowest level written to the log. debug also logs probe requests.",
  },
  LOG_FORMAT: {
    schema: v.optional(v.picklist(["text", "json"]), "text"),
    description: "text for people, json for log pipelines; the same two shapes as the API.",
  },
  APP_VERSION: {
    static: true,
    schema: v.optional(v.string(), "dev"),
    description:
      "Build version, inlined at build time. /readyz fails while the API reports a different one.",
  },
});
