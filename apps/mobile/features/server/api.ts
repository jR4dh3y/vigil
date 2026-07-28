import { fetch } from "expo/fetch";
import { normalizeApiBaseUrl } from "@/lib/api/config";

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
	if (!isHealth(body)) {
		throw new Error("This address did not return a valid Vigil recorder response");
	}
	return { baseUrl, status: body.status };
}
