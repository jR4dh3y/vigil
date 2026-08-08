/**
 * The active API base URL and connection mode, shared between the runtime
 * client middleware and the reactive connection store. Kept framework-free so
 * `client.ts` and pure helpers can read the current value without runes.
 */

export type ConnectionMode = "detecting" | "embedded" | "remote" | "none";

const INITIAL_BASE = "/api/v1";

let activeBaseUrl = INITIAL_BASE;
let activeMode: ConnectionMode = "detecting";

/** The canonical `/api/v1` base URL the API client should target right now. */
export function getActiveBaseUrl(): string {
	return activeBaseUrl;
}

/** The current connection mode. */
export function getActiveMode(): ConnectionMode {
	return activeMode;
}

/** Update the active base URL/mode and notify subscribers. */
export function setActiveBase(baseUrl: string, mode: ConnectionMode): void {
	if (baseUrl === activeBaseUrl && mode === activeMode) {
		return;
	}
	activeBaseUrl = baseUrl;
	activeMode = mode;
}
