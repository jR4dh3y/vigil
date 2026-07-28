import type { Camera, LiveStream, PlaybackSession } from "@nvr/api-client";
import { getApiClient } from "@/lib/api/client";
import { throwApiError } from "@/lib/api/error";

export async function listCameras(): Promise<Camera[]> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/cameras");
	if (data) {
		return data.cameras;
	}
	throwApiError(error, response.status, "Could not load cameras");
}

export async function getCamera(id: string): Promise<Camera> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/cameras/{id}", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not load camera");
}

export async function getLiveStream(id: string): Promise<LiveStream> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/{id}/live", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Live stream is unavailable");
}

export async function getPlayback(
	cameraId: string,
	start: string,
	durationSec = 60,
): Promise<PlaybackSession> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/{id}/playback", {
		params: { path: { id: cameraId } },
		body: { start, durationSec },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Recorded video is unavailable");
}

export function liveRefetchInterval(expiresAt?: string): number | false {
	if (!expiresAt) {
		return false;
	}
	const expiry = Date.parse(expiresAt);
	if (Number.isNaN(expiry)) {
		return false;
	}
	return Math.max(expiry - Date.now() - 10_000, 5_000);
}

export const cameraKeys = {
	all: ["cameras"] as const,
	detail: (id: string) => ["cameras", id] as const,
	live: (id: string) => ["cameras", id, "live"] as const,
	playback: (id: string, start: string) => ["cameras", id, "playback", start] as const,
};
