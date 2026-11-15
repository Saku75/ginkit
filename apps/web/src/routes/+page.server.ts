import { v1Example } from "#lib/api/sdk.gen.js";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ fetch }) => {
  const { data } = await v1Example({ fetch });
  return { message: data?.message ?? null };
};
