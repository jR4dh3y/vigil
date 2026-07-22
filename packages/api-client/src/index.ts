import createClient from "openapi-fetch";
import type { paths } from "./generated/schema";

export type { paths };
export type ApiClient = ReturnType<typeof createClient<paths>>;

export function createApiClient(baseUrl: string): ApiClient {
	return createClient<paths>({ baseUrl });
}
