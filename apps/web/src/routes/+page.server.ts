import { error } from "@sveltejs/kit";
import { v1Example } from "#lib/api/sdk.gen.js";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async () => {
  const result = await v1Example();
  if (!result.data) error(502, result.error.detail ?? "The API did not answer");
  return { message: result.data.message };
};
