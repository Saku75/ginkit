import { defineEnvVars } from "@sveltejs/kit/env";
import * as v from "valibot";

export const variables = defineEnvVars({
  API_URL: {
    schema: v.optional(v.pipe(v.string(), v.url()), "http://localhost:3101"),
    description:
      "Where server-side rendering reaches the API. The browser uses /api on its own origin.",
  },
});
