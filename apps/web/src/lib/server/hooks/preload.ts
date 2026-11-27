import type { Handle } from "@sveltejs/kit/hooks";

const fontFiles = ["inter-latin-wght-normal"];

export const preloadServerHook: Handle = ({ event, resolve }) =>
  resolve(event, {
    preload: ({ type, path }) =>
      type === "font"
        ? fontFiles.some((file) => path.includes(file))
        : type === "css" || type === "js",
  });
