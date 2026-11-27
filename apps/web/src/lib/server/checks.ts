import { APP_VERSION } from "$app/env/private";
import { version } from "#lib/api/sdk.gen.js";

// Check mirrors handlers.Check in the API: run resolves while its dependency
// is ready and rejects with the reason otherwise. Names must be unique.
export type Check = { name: string; run: () => Promise<void> };

// The API must answer and be built from the same version, so a rolling deploy
// holds traffic from this app until the API has rolled too.
const api: Check = {
  name: "api",
  run: async () => {
    const result = await version({ signal: AbortSignal.timeout(2000) });
    if (!result.data) {
      throw new Error(
        result.response ? `API answered ${result.response.status}` : "API unreachable",
      );
    }
    if (result.data.version !== APP_VERSION) {
      throw new Error(`API is ${result.data.version}, this build is ${APP_VERSION}`);
    }
  },
};

export const checks: Check[] = [api];
