import { APP_VERSION } from "$app/env/private";
import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = () => json({ version: APP_VERSION });
