import { fetch } from "expo/fetch";
import { normalizeApiBaseUrl } from "@/lib/api/config";

const CONNECTION_TIMEOUT_MS = 8_000;

export type RecorderConnection = {
	baseUrl: string;
	status: "ok" | "degraded";
};

function isHealth(value: unknown): value is { status: "ok" | "degraded" } {
	if (typeof value !== "object" || value === null || !("status" in value)) {
		return false;
	}
	return value.status === "ok" || value.status === "degraded";
}

export async function testRecorderConnection(value: string): Promise<RecorderConnection> {
	const baseUrl = normalizeApiBaseUrl(value);
	const controller = new AbortController();
	const timeout = setTimeout(() => controller.abort(), CONNECTION_TIMEOUT_MS);
	let response: Response;
	try {
		response = await fetch(`${baseUrl}/health`, {
			headers: { Accept: "application/json" },
			signal: controller.signal,
		});
	} catch {
		if (controller.signal.aborted) {
			throw new Error(
				"The recorder did not respond within 8 seconds. Check the address and make sure this device is on the same LAN or Tailscale network.",
			);
		}
		throw new Error("Could not reach a recorder at this address");
	} finally {
		clearTimeout(timeout);
	}

	if (!response.ok) {
		throw new Error(`Recorder returned HTTP ${response.status}`);
	}

	const body: unknown = await response.json().catch(() => null);
	if (!isHealth(body)) {
		throw new Error("This address did not return a valid Vigil recorder response");
	}
	return { baseUrl, status: body.status };
}
