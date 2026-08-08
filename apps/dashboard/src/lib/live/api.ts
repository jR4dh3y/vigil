import type { LiveStream } from "@nvr/api-client";
import { getApiClient } from "$lib/connection";

export class LiveApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "LiveApiError";
		this.status = status;
		this.code = code;
	}
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

function errorDetails(body: unknown, fallback: string): { message: string; code?: string } {
	if (!isRecord(body)) {
		return { message: fallback };
	}
	const message =
		typeof body.error === "string" && body.error.trim() ? body.error.trim() : fallback;
	const code = typeof body.code === "string" ? body.code : undefined;
	return { message, code };
}

function throwLiveError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new LiveApiError(message, status, code);
}

/** POST /cameras/{id}/live — WHEP + HLS URLs and a short-lived stream token. */
export async function requestLive(cameraId: string): Promise<LiveStream> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/{id}/live", {
		params: { path: { id: cameraId } },
	});
	if (data) {
		return data;
	}
	throwLiveError(error, response.status, "Failed to start live stream");
}
