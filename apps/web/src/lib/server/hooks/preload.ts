import type { Handle } from "@sveltejs/kit/hooks";

const fontFiles = ["inter-latin-wght-normal"];

export const preloadServerHook: Handle = async ({ event, resolve }) => {
  return await resolve(event, {
    preload: ({ type, path }) => {
      switch (type) {
        case "font":
          return fontFiles.some((file) => path.includes(file));
        case "css":
        case "js":
          return true;
        default:
          return false;
      }
    },
  });
};
