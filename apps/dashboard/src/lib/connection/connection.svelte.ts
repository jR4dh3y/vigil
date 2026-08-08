/**
 * Reactive connection store (Svelte runes). Owns the mode machine:
 *
 *   detecting → embedded | remote | none
 *
 * - `detecting`: probing same-origin / hydrating stored config
 * - `embedded`: same-origin cookie mode, no server field
 * - `remote`: persisted server + Bearer token
 * - `none`: no active server — show the Server URL gate
 *
 * The `server` query parameter takes precedence as an editable prefill (a
 * deep link); it never bypasses normalize/HTTPS/health validation.
 */
import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import { type ConnectionMode, getActiveBaseUrl, setActiveBase } from "./active";
import { resetApiClient } from "./client";
import { probeSameOrigin, testRecorderConnection } from "./connection";
import { clearToken } from "./token";
import { clearRemoteServer, getRemoteServer, normalizeServerUrl, setRemoteServer } from "./url";

export const connection = $state({
	mode: "detecting" as ConnectionMode,
	error: null as string | null,
	connecting: false,
	activeBaseUrl: getActiveBaseUrl(),
	prefill: "",
});

/** The active server origin (or empty when embedded/detecting). */
export function activeOrigin(): string {
	try {
		return new URL(
			connection.activeBaseUrl,
			typeof window !== "undefined" ? window.location.href : undefined,
		).origin;
	} catch {
		return "";
	}
}

async function applyRemoteBase(baseUrl: string, clearQuery: boolean): Promise<void> {
	setActiveBase(baseUrl, "remote");
	connection.activeBaseUrl = baseUrl;
	connection.mode = "remote";
	connection.error = null;
	resetApiClient();
	if (clearQuery) {
		await goto(resolve("/"), { replaceState: true });
	}
}

/**
 * Initialize the connection for the page. `prefill` is the raw `?server=`
 * value (or null). Resolution order: query prefill → stored remote → probe
 * same-origin → hosted gate.
 */
export async function initConnection(prefill: string | null): Promise<void> {
	connection.mode = "detecting";
	connection.error = null;

	if (prefill) {
		// Deep link: surface the address in the gate field for the user to
		// confirm through the normal validation path.
		connection.prefill = prefill;
		connection.mode = "none";
		return;
	}

	const stored = getRemoteServer();
	if (stored) {
		try {
			const normalized = normalizeServerUrl(stored);
			setActiveBase(normalized, "remote");
			connection.activeBaseUrl = normalized;
			connection.mode = "remote";
			resetApiClient();
			return;
		} catch {
			// Stored address is no longer valid (e.g. plain HTTP under HTTPS) —
			// drop it and fall through to detection.
			clearRemoteServer();
			clearToken();
		}
	}

	const sameOrigin = await probeSameOrigin();
	if (sameOrigin) {
		setActiveBase("/api/v1", "embedded");
		connection.activeBaseUrl = "/api/v1";
		connection.mode = "embedded";
		return;
	}

	connection.mode = "none";
}

/**
 * Validate and persist a remote server, clearing any stale token, and switch
 * the runtime to it. On success the caller's query cache is cleared by
 * `clearQuery`.
 */
export async function connectServer(value: string): Promise<boolean> {
	connection.connecting = true;
	connection.error = null;
	try {
		const { baseUrl } = await testRecorderConnection(value);
		setRemoteServer(baseUrl);
		clearToken();
		await applyRemoteBase(baseUrl, true);
		return true;
	} catch (cause) {
		connection.error = cause instanceof Error ? cause.message : "Could not connect to this server";
		return false;
	} finally {
		connection.connecting = false;
	}
}

/**
 * Return to the server gate (e.g. from a connection error or the "Change
 * server" affordance), prefilling the current address.
 */
export function changeServer(): void {
	connection.prefill = connection.mode === "remote" ? connection.activeBaseUrl : connection.prefill;
	connection.mode = "none";
	connection.error = null;
}
