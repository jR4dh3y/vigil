import type { DiskInfo, PatchSettingsRequest, Settings, SystemStatus } from "@nvr/api-client";
import { getApiClient } from "$lib/connection";

export class SystemApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "SystemApiError";
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

function throwSystemError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new SystemApiError(message, status, code);
}

/** GET /system/status */
export async function getSystemStatus(): Promise<SystemStatus> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/system/status");
	if (data) {
		return data;
	}
	throwSystemError(error, response.status, "Failed to load system status");
}

/** GET /system/disk */
export async function getSystemDisk(): Promise<DiskInfo> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/system/disk");
	if (data) {
		return data;
	}
	throwSystemError(error, response.status, "Failed to load disk info");
}

/** GET /settings */
export async function getSettings(): Promise<Settings> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/settings");
	if (data) {
		return data;
	}
	throwSystemError(error, response.status, "Failed to load settings");
}

/** PATCH /settings — admin only. */
export async function patchSettings(body: PatchSettingsRequest): Promise<Settings> {
	const api = getApiClient();
	const { data, error, response } = await api.PATCH("/settings", { body });
	if (data) {
		return data;
	}
	throwSystemError(error, response.status, "Failed to update settings");
}
