import { defineEnvVars } from "@sveltejs/kit/env";
import * as v from "valibot";

export const variables = defineEnvVars({
  API_URL: {
    schema: v.optional(v.pipe(v.string(), v.url()), "http://localhost:3101"),
    description:
      "Where server-side rendering reaches the API. The browser uses /api on its own origin.",
  },
  APP_VERSION: {
    static: true,
    schema: v.optional(v.string(), "dev"),
    description:
      "Build version, inlined at build time. /readyz fails while the API reports a different one.",
  },
});
