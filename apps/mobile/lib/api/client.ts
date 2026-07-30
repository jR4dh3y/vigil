import { createApiClient, type Middleware } from "@nvr/api-client";
import { getApiBaseUrl } from "@/lib/api/config";
import { clearSessionToken, getSessionToken, setSessionToken } from "@/lib/api/session";

const SESSION_HEADER = "X-Session-Token";

const sessionMiddleware: Middleware = {
	async onRequest({ request }) {
		const token = await getSessionToken();
		if (!token) {
			// No change — omit return so openapi-fetch keeps the original Request.
			return undefined;
		}

		// Must return a *new* Request when modifying (openapi-fetch identity check).
		const headers = new Headers(request.headers);
		headers.set("Authorization", `Bearer ${token}`);
		return new Request(request, { headers });
	},
	async onResponse({ response, request }) {
		// Ignore responses from a recorder that is no longer active: a request
		// still in flight during a recorder switch must not persist the old
		// recorder's session token or clear the new recorder's session.
		if (!request.url.startsWith(getApiBaseUrl())) {
			return undefined;
		}

		const issued = response.headers.get(SESSION_HEADER);
		if (issued) {
			await setSessionToken(issued);
		}

		// Clear local token when the server rejects it (not on login/setup/status).
		const path = new URL(request.url).pathname;
		const isAuthForm =
			path.endsWith("/auth/login") || path.endsWith("/auth/setup") || path.endsWith("/auth/status");
		if (response.status === 401 && !isAuthForm) {
			await clearSessionToken();
		}

		// Side effects only. Do not return the same Response — on React Native it often
		// fails `instanceof Response` and openapi-fetch throws:
		// "onResponse: must return new Response() when modifying the response"
		return undefined;
	},
};

let cachedBaseUrl: string | null = null;
let cachedClient: ReturnType<typeof createApiClient> | null = null;

export function getApiClient(): ReturnType<typeof createApiClient> {
	const baseUrl = getApiBaseUrl();
	if (cachedClient && cachedBaseUrl === baseUrl) {
		return cachedClient;
	}

	const client = createApiClient(baseUrl, { credentials: "include" });
	client.use(sessionMiddleware);
	cachedBaseUrl = baseUrl;
	cachedClient = client;
	return client;
}
