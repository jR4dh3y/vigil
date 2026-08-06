import type { PlaybackSession, RecordingList } from "@nvr/api-client";
import { api } from "$lib/api";

export class RecordingApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "RecordingApiError";
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

function throwRecordingError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new RecordingApiError(message, status, code);
}

const DEFAULT_PLAYBACK_DURATION_SEC = 60;

/** GET /cameras/{id}/recordings?from=&to= */
export async function listRecordings(
	cameraId: string,
	from: string,
	to: string,
): Promise<RecordingList> {
	const { data, error, response } = await api.GET("/cameras/{id}/recordings", {
		params: {
			path: { id: cameraId },
			query: { from, to },
		},
	});
	if (data) {
		return data;
	}
	throwRecordingError(error, response.status, "Failed to load recordings");
}

/** POST /cameras/{id}/playback — mint a short-lived VOD session. */
export async function requestPlayback(
	cameraId: string,
	start: string,
	durationSec: number = DEFAULT_PLAYBACK_DURATION_SEC,
): Promise<PlaybackSession> {
	const { data, error, response } = await api.POST("/cameras/{id}/playback", {
		params: { path: { id: cameraId } },
		body: { start, durationSec },
	});
	if (data) {
		return data;
	}
	throwRecordingError(error, response.status, "Failed to start playback");
}
