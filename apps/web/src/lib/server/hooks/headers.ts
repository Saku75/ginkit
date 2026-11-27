import type { Handle } from "@sveltejs/kit/hooks";

export const headersServerHook: Handle = ({ event, resolve }) => {
  event.setHeaders({
    "X-Content-Type-Options": "nosniff",
    "X-Frame-Options": "SAMEORIGIN",
    "Referrer-Policy": "strict-origin-when-cross-origin",
  });
  return resolve(event);
};
