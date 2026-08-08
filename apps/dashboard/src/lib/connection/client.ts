/**
 * Runtime OpenAPI client, cached per active base URL.
 *
 * Remote (hosted) mode mirrors the mobile middleware contract:
 * - adds `Authorization: Bearer <token>`
 * - captures the exposed `X-Session-Token` response header
 * - ignores responses whose request no longer targets the active server, so a
 *   stale in-flight request can never persist the *old* server's token or clear
 *   the *new* server's session
 * - clears the token on an authenticated 401 (not login/setup/status)
 *
 * Embedded (same-origin) mode keeps the relative `/api/v1` base and relies on
 * HttpOnly cookies via `credentials: include`; no Bearer is added.
 */
import { type ApiClient, createApiClient, type Middleware } from "@nvr/api-client";
import { getActiveBaseUrl, getActiveMode } from "./active";
import { clearToken, getToken, setToken } from "./token";

const SESSION_HEADER = "X-Session-Token";

/** True when `requestUrl` targets the currently active server base. */
function isActiveRequest(requestUrl: string): boolean {
	const base = getActiveBaseUrl();

	// The base may be relative (`/api/v1`) or absolute (`https://host/api/v1`);
	// `request.url` is always absolute once the Request is constructed.
	if (/^https?:\/\//i.test(base)) {
		return requestUrl.startsWith(base);
	}

	try {
		const absolute = new URL(base, window.location.href).href;
		return requestUrl.startsWith(absolute);
	} catch {
		return requestUrl.startsWith(base);
	}
}

const sessionMiddleware: Middleware = {
	onRequest({ request }) {
		if (getActiveMode() !== "remote") {
			// Embedded mode: cookies only, never a Bearer header.
			return undefined;
		}
		const token = getToken();
		if (!token) {
			return undefined;
		}
		const headers = new Headers(request.headers);
		headers.set("Authorization", `Bearer ${token}`);
		// openapi-fetch requires a *new* Request when the headers are modified.
		return new Request(request, { headers });
	},
	onResponse({ response, request }) {
		// A response from a server that is no longer active must not affect the
		// current session (token persistence or clearing).
		if (!isActiveRequest(request.url)) {
			return undefined;
		}

		if (getActiveMode() === "remote") {
			const issued = response.headers.get(SESSION_HEADER);
			if (issued) {
				setToken(issued);
			}

			const pathname = new URL(request.url).pathname;
			const isAuthForm =
				pathname.endsWith("/auth/login") ||
				pathname.endsWith("/auth/setup") ||
				pathname.endsWith("/auth/status");
			if (response.status === 401 && !isAuthForm) {
				clearToken();
			}
		}

		// Side effects only — return the original response untouched.
		return undefined;
	},
};

let cachedBaseUrl: string | null = null;
let cachedClient: ApiClient | null = null;

/** Reset the client cache so the next call rebuilds against the active base. */
export function resetApiClient(): void {
	cachedBaseUrl = null;
	cachedClient = null;
}

/** The runtime client for the current active server, cached per base URL. */
export function getApiClient(): ApiClient {
	const baseUrl = getActiveBaseUrl();
	if (cachedClient && cachedBaseUrl === baseUrl) {
		return cachedClient;
	}

	// Embedded (same-origin) mode relies on HttpOnly cookies, so it sends
	// credentials. Remote Bearer mode carries auth in the header and must not
	// leak cookies to the remote recorder.
	const credentials = getActiveMode() === "embedded" ? "include" : "omit";
	const client = createApiClient(baseUrl, { credentials });
	client.use(sessionMiddleware);
	cachedBaseUrl = baseUrl;
	cachedClient = client;
	return client;
}
