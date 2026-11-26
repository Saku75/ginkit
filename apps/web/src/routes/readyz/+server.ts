import { json } from "@sveltejs/kit";
import { checks } from "#lib/server/checks.js";
import type { RequestHandler } from "./$types";

// Same contract as the API's /readyz: 200 with every check's result, or 503 in
// the API's error format naming each check that failed.
export const GET: RequestHandler = async () => {
  const results = await Promise.allSettled(checks.map((check) => check.run()));
  const failed = results.flatMap((result, i) =>
    result.status === "rejected"
      ? [{ location: `checks.${checks[i].name}`, message: reason(result.reason) }]
      : [],
  );

  if (failed.length > 0) {
    return json(
      { title: "Service Unavailable", status: 503, detail: "service is not ready", errors: failed },
      { status: 503, headers: { "content-type": "application/problem+json" } },
    );
  }
  return json({
    status: "ok",
    checks: Object.fromEntries(checks.map((check) => [check.name, "ok"])),
  });
};

const reason = (error: unknown) => (error instanceof Error ? error.message : String(error));
