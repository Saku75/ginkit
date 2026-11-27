import { APP_VERSION } from "$app/env/private";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = () => Response.json({ version: APP_VERSION });
