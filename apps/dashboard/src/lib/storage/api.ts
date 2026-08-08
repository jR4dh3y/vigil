import type {
	GDriveArchiveRequest,
	GDriveArchiveResponse,
	GDriveConfigurationRequest,
	GDriveConnectResponse,
	GDriveStatus,
} from "@nvr/api-client";
import { getApiClient } from "$lib/connection";

export class StorageApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "StorageApiError";
		this.status = status;
		this.code = code;
	}
}

/** PUT /storage/gdrive/configuration — saves encrypted OAuth app credentials. */
export async function putGDriveConfiguration(
	body: GDriveConfigurationRequest,
): Promise<GDriveStatus> {
	const api = getApiClient();
	const { data, error, response } = await api.PUT("/storage/gdrive/configuration", { body });
	if (data) {
		return data;
	}
	throwStorageError(error, response.status, "Failed to save Google Drive configuration");
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

function throwStorageError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new StorageApiError(message, status, code);
}

/** GET /storage/gdrive/status */
export async function getGDriveStatus(): Promise<GDriveStatus> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/storage/gdrive/status");
	if (data) {
		return data;
	}
	throwStorageError(error, response.status, "Failed to load Google Drive status");
}

/** POST /storage/gdrive/connect — returns OAuth authorization URL. */
export async function postGDriveConnect(): Promise<GDriveConnectResponse> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/storage/gdrive/connect");
	if (data) {
		return data;
	}
	throwStorageError(error, response.status, "Failed to start Google Drive connection");
}

/** DELETE /storage/gdrive/disconnect — admin only. */
export async function deleteGDriveDisconnect(): Promise<void> {
	const api = getApiClient();
	const { error, response } = await api.DELETE("/storage/gdrive/disconnect");
	if (response.ok || response.status === 204) {
		return;
	}
	throwStorageError(error, response.status, "Failed to disconnect Google Drive");
}

/** POST /storage/gdrive/archive — admin only; runs an immediate archive pass. */
export async function postGDriveArchive(
	body?: GDriveArchiveRequest,
): Promise<GDriveArchiveResponse> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/storage/gdrive/archive", {
		body: body ?? { limit: 50 },
	});
	if (data) {
		return data;
	}
	throwStorageError(error, response.status, "Failed to run Google Drive archive");
}
