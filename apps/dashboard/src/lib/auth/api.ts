import type { AuthStatus, LoginRequest, SetupRequest, UserPublic } from "@nvr/api-client";
import { getApiClient } from "$lib/connection";
import { clearToken } from "$lib/connection/token";

export class AuthApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "AuthApiError";
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

function throwAuthError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new AuthApiError(message, status, code);
}

/** GET /auth/status */
export async function loadStatus(): Promise<AuthStatus> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/auth/status");
	if (data) {
		return data;
	}
	throwAuthError(error, response.status, "Failed to load auth status");
}

/** POST /auth/setup — creates first admin and sets session cookie. */
export async function setup(body: SetupRequest): Promise<UserPublic> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/auth/setup", { body });
	if (data) {
		return data;
	}
	throwAuthError(error, response.status, "Setup failed");
}

/** POST /auth/login — authenticates and sets session cookie. */
export async function login(body: LoginRequest): Promise<UserPublic> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/auth/login", { body });
	if (data) {
		return data;
	}
	throwAuthError(error, response.status, "Login failed");
}

/** POST /auth/logout — clears session cookie and any remote token. */
export async function logout(): Promise<void> {
	const api = getApiClient();
	const { response } = await api.POST("/auth/logout");
	if (response.ok || response.status === 204) {
		clearToken();
		return;
	}
	throw new AuthApiError("Logout failed", response.status);
}

/** GET /auth/me */
export async function loadMe(): Promise<UserPublic> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/auth/me");
	if (data) {
		return data;
	}
	throwAuthError(error, response.status, "Not authenticated");
}
