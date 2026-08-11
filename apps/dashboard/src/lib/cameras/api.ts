import type {
	Camera,
	CreateCameraRequest,
	DiscoverCameraStreamsRequest,
	DiscoverCameraStreamsResult,
	DiscoverCamerasRequest,
	DiscoverResult,
	ProbeCameraRequest,
	ProbeResult,
	UpdateCameraRequest,
} from "@nvr/api-client";
import { getApiClient } from "$lib/connection";

export class CameraApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "CameraApiError";
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

function throwCameraError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new CameraApiError(message, status, code);
}

/** POST /cameras/discover */
export async function discoverCameras(
	body: DiscoverCamerasRequest,
): Promise<DiscoverResult["cameras"]> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/discover", { body });
	if (data) {
		return data.cameras;
	}
	throwCameraError(error, response.status, "Failed to discover cameras");
}

/** POST /cameras/discover/streams */
export async function discoverCameraStreams(
	body: DiscoverCameraStreamsRequest,
): Promise<DiscoverCameraStreamsResult> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/discover/streams", { body });
	if (data) {
		return data;
	}
	throwCameraError(error, response.status, "Failed to detect camera streams");
}

/** GET /cameras */
export async function listCameras(): Promise<Camera[]> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/cameras");
	if (data) {
		return data.cameras;
	}
	throwCameraError(error, response.status, "Failed to load cameras");
}

/** GET /cameras/{id} */
export async function getCamera(id: string): Promise<Camera> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/cameras/{id}", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwCameraError(error, response.status, "Failed to load camera");
}

/** POST /cameras */
export async function createCamera(body: CreateCameraRequest): Promise<Camera> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras", { body });
	if (data) {
		return data;
	}
	throwCameraError(error, response.status, "Failed to create camera");
}

/** PATCH /cameras/{id} */
export async function updateCamera(id: string, body: UpdateCameraRequest): Promise<Camera> {
	const api = getApiClient();
	const { data, error, response } = await api.PATCH("/cameras/{id}", {
		params: { path: { id } },
		body,
	});
	if (data) {
		return data;
	}
	throwCameraError(error, response.status, "Failed to update camera");
}

/** DELETE /cameras/{id} */
export async function deleteCamera(id: string): Promise<void> {
	const api = getApiClient();
	const { error, response } = await api.DELETE("/cameras/{id}", {
		params: { path: { id } },
	});
	if (response.ok || response.status === 204) {
		return;
	}
	throwCameraError(error, response.status, "Failed to delete camera");
}

/** POST /cameras/probe */
export async function probeCamera(body: ProbeCameraRequest): Promise<ProbeResult> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/cameras/probe", { body });
	if (data) {
		return data;
	}
	throwCameraError(error, response.status, "Failed to probe stream");
}

/** GET /cameras/{id}/snapshot — JPEG blob. */
export async function getCameraSnapshot(id: string): Promise<Blob> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/cameras/{id}/snapshot", {
		params: { path: { id } },
		parseAs: "blob",
	});
	if (response.ok && data instanceof Blob && data.size > 0) {
		return data;
	}
	throwCameraError(error, response.status, "Snapshot not available");
}
