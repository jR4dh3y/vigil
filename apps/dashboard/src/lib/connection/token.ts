/**
 * Remote session-token persistence. Only used in hosted (remote-server) mode;
 * embedded mode relies on HttpOnly cookies and never stores a token.
 *
 * localStorage failures degrade to in-memory state for the current page.
 */

const TOKEN_KEY = "nvr_session";

let memoryToken: string | null | undefined;

function hydrateToken(): string | null {
	if (memoryToken !== undefined) {
		return memoryToken;
	}
	try {
		memoryToken = window.localStorage.getItem(TOKEN_KEY);
	} catch {
		memoryToken = null;
	}
	return memoryToken;
}

/** The currently stored remote session token, or null. */
export function getToken(): string | null {
	if (typeof window === "undefined") {
		return null;
	}
	return hydrateToken();
}

/** Persist the remote session token. */
export function setToken(token: string): void {
	memoryToken = token;
	if (typeof window === "undefined") {
		return;
	}
	try {
		window.localStorage.setItem(TOKEN_KEY, token);
	} catch {
		// Keep the in-memory token for this process.
	}
}

/** Clear the remote session token. */
export function clearToken(): void {
	memoryToken = null;
	if (typeof window === "undefined") {
		return;
	}
	try {
		window.localStorage.removeItem(TOKEN_KEY);
	} catch {
		// ignore
	}
}
