import { APP_VERSION } from "$app/env/private";
import { json } from "@sveltejs/kit";
import { version } from "#lib/api/sdk.gen.js";
import type { RequestHandler } from "./$types";

// Ready only when the API reports the same build version, so a rolling deploy
// holds traffic from this app until the API has rolled too.
export const GET: RequestHandler = async () => {
  const result = await version({ signal: AbortSignal.timeout(2000) }).catch(() => undefined);
  const api = result?.data?.version ?? null;

  if (api === null) {
    const detail = result?.response ? `API answered ${result.response.status}` : "API unreachable";
    return json({ status: "unavailable", version: APP_VERSION, api, detail }, { status: 503 });
  }
  if (api !== APP_VERSION) {
    return json(
      { status: "unavailable", version: APP_VERSION, api, detail: "API version mismatch" },
      { status: 503 },
    );
  }
  return json({ status: "ok", version: APP_VERSION, api });
};
