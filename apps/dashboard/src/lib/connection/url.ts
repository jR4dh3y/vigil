/**
 * Pure server-address helpers plus localStorage persistence.
 *
 * Normalization contract (shared with the mobile client):
 * - require a protocol, defaulting to the page protocol
 * - reject non-HTTP(S) and embedded credentials
 * - strip query/hash
 * - append exactly `/api/v1`
 * - reject a plain-HTTP address when the page is served over HTTPS
 *
 * All localStorage access degrades to in-memory state so the dashboard never
 * crashes when storage is unavailable (private mode, sandboxed iframe, …).
 */

const REMOTE_SERVER_KEY = "nvr_remote_server";

/**
 * Normalize a user-supplied recorder address to a canonical `/api/v1` base URL.
 * Throws a human-readable error on invalid input.
 */
export function normalizeServerUrl(input: string): string {
	const candidate = input.trim();
	if (!candidate) {
		throw new Error("Enter the address of your recorder");
	}

	const pageIsHttps = typeof window !== "undefined" && window.location.protocol === "https:";
	const withProtocol = /^https?:\/\//i.test(candidate)
		? candidate
		: `${pageIsHttps ? "https://" : "http://"}${candidate}`;

	let url: URL;
	try {
		url = new URL(withProtocol);
	} catch {
		throw new Error("That doesn't look like a valid server address");
	}

	if (url.protocol !== "http:" && url.protocol !== "https:") {
		throw new Error("Server address must use HTTP or HTTPS");
	}
	if (url.username || url.password) {
		throw new Error("Server address must not contain credentials");
	}
	if (pageIsHttps && url.protocol === "http:") {
		throw new Error("This page is served over HTTPS; use an HTTPS server address");
	}

	url.search = "";
	url.hash = "";
	const path = url.pathname.replace(/\/+$/, "");
	url.pathname = path.endsWith("/api/v1") ? path : `${path}/api/v1`;
	return url.toString().replace(/\/$/, "");
}

export type ServerRef = {
	/** Scheme + host + port, e.g. `https://recorder.example:8443`. */
	origin: string;
	/** Compact host display, e.g. `recorder.example:8443`. */
	display: string;
	/** The canonical `/api/v1` base URL. */
	baseUrl: string;
};

/** Derive display/origin from a normalized `/api/v1` base URL. */
export function deriveServerRef(baseUrl: string): ServerRef {
	const url = new URL(baseUrl);
	const port = url.port ? `:${url.port}` : "";
	return {
		origin: url.origin,
		display: `${url.hostname}${port}`,
		baseUrl,
	};
}

let memoryRemoteServer: string | null | undefined;

function hydrateRemoteServer(): string | null {
	if (memoryRemoteServer !== undefined) {
		return memoryRemoteServer;
	}
	try {
		memoryRemoteServer = window.localStorage.getItem(REMOTE_SERVER_KEY);
	} catch {
		memoryRemoteServer = null;
	}
	return memoryRemoteServer;
}

/** The stored remote `/api/v1` base URL, or null when unset. */
export function getRemoteServer(): string | null {
	if (typeof window === "undefined") {
		return null;
	}
	return hydrateRemoteServer();
}

/** Persist a single remote server base URL. */
export function setRemoteServer(baseUrl: string): void {
	memoryRemoteServer = baseUrl;
	if (typeof window === "undefined") {
		return;
	}
	try {
		window.localStorage.setItem(REMOTE_SERVER_KEY, baseUrl);
	} catch {
		// Keep the in-memory value for this process.
	}
}

/** Clear the stored remote server. */
export function clearRemoteServer(): void {
	memoryRemoteServer = null;
	if (typeof window === "undefined") {
		return;
	}
	try {
		window.localStorage.removeItem(REMOTE_SERVER_KEY);
	} catch {
		// ignore
	}
}
