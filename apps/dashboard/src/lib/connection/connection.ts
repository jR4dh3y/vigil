/**
 * Connection validation: probes a candidate server's `/health` endpoint and
 * requires the Vigil JSON contract `{ "status": "ok" }` before accepting it.
 * Used both for same-origin embedded detection and for remote-server saves.
 */
import { normalizeServerUrl } from "./url";

export type ConnectionResult = {
	/** The normalized `/api/v1` base URL. */
	baseUrl: string;
};

function isHealthOk(value: unknown): boolean {
	if (typeof value !== "object" || value === null) {
		return false;
	}
	return "status" in value && value.status === "ok";
}

/**
 * Probe same-origin `/api/v1/health`. The embedded (cookie) mode is only
 * accepted when this returns the Vigil JSON contract.
 */
export async function probeSameOrigin(): Promise<boolean> {
	try {
		const response = await fetch("/api/v1/health", {
			headers: { Accept: "application/json" },
		});
		if (!response.ok) {
			return false;
		}
		const body: unknown = await response.json().catch(() => null);
		return isHealthOk(body);
	} catch {
		return false;
	}
}

/**
 * Test a remote server address through its normalized `/health` endpoint.
 * Throws a human-readable error when unreachable or not a Vigil recorder.
 */
export async function testRecorderConnection(value: string): Promise<ConnectionResult> {
	const baseUrl = normalizeServerUrl(value);

	let response: Response;
	try {
		response = await fetch(`${baseUrl}/health`, {
			headers: { Accept: "application/json" },
		});
	} catch {
		throw new Error("Could not reach a recorder at this address");
	}

	if (!response.ok) {
		throw new Error(`Recorder returned HTTP ${response.status}`);
	}

	const body: unknown = await response.json().catch(() => null);
	if (!isHealthOk(body)) {
		throw new Error("This address did not return a valid Vigil recorder response");
	}
	return { baseUrl };
}
