import type { PlaybackSession, RecordingDayAvailability, RecordingList } from "@nvr/api-client";
import { getApiClient } from "@/lib/api/client";
import { throwApiError } from "@/lib/api/error";

export async function listRecordingDays(
	cameraId: string,
	from: string,
	to: string,
	timeZone: string,
): Promise<RecordingDayAvailability[]> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/recordings/days", {
		params: { query: { cameraId, from, to, timeZone } },
	});
	if (data) {
		return data.days;
	}
	throwApiError(error, response.status, "Could not load recording days");
}

export async function listCameraRecordings(
	cameraId: string,
	from: string,
	to: string,
): Promise<RecordingList> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/cameras/{id}/recordings", {
		params: { path: { id: cameraId }, query: { from, to } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not load recordings");
}

export async function requestRecordingPlayback(
	cameraId: string,
	start: string,
): Promise<PlaybackSession> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/{id}/playback", {
		params: { path: { id: cameraId } },
		body: { start, durationSec: 60 },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Recorded video is unavailable");
}

export const recordingKeys = {
	all: ["recordings"] as const,
	days: (cameraId: string, from: string, to: string, timeZone: string) =>
		["recordings", cameraId, "days", from, to, timeZone] as const,
	list: (cameraId: string, from: string, to: string) =>
		["recordings", cameraId, "list", from, to] as const,
	playback: (cameraId: string, start: string) =>
		["recordings", cameraId, "playback", start] as const,
};
