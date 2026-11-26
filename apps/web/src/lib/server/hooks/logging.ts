import type { Handle } from "@sveltejs/kit/hooks";
import { log } from "#lib/server/log.js";

// Probe traffic is noise unless the probes themselves are being debugged.
const probes = new Set(["/healthz", "/readyz", "/version"]);

// One line per request, like the API's request log: 2xx and 3xx at info, 4xx
// at warn, 5xx at error. Runs first in the sequence so latency covers everything.
export const loggingServerHook: Handle = async ({ event, resolve }) => {
  const start = performance.now();
  const response = await resolve(event);

  const attrs = {
    request: {
      method: event.request.method,
      path: event.url.pathname,
      query: event.url.search.slice(1),
      ip: clientAddress(event),
    },
    response: { status: response.status, latency: `${(performance.now() - start).toFixed(1)}ms` },
  };
  if (probes.has(event.url.pathname)) log.debug("request", attrs);
  else if (response.status >= 500) log.error("request", attrs);
  else if (response.status >= 400) log.warn("request", attrs);
  else log.info("request", attrs);

  return response;
};

function clientAddress(event: Parameters<Handle>[0]["event"]) {
  try {
    return event.getClientAddress();
  } catch {
    return undefined;
  }
}
