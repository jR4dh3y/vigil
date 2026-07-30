import type { AuthStatus, LoginRequest, SetupRequest, UserPublic } from "@nvr/api-client";
import { getApiClient } from "@/lib/api/client";
import { throwApiError } from "@/lib/api/error";
import { clearSessionToken } from "@/lib/api/session";

export async function loadAuthStatus(): Promise<AuthStatus> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/auth/status");
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not reach the recorder");
}

export async function login(body: LoginRequest): Promise<UserPublic> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/auth/login", { body });
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Login failed");
}

export async function setup(body: SetupRequest): Promise<UserPublic> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/auth/setup", { body });
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Setup failed");
}

export async function logout(): Promise<void> {
	try {
		const api = getApiClient();
		await api.POST("/auth/logout");
	} finally {
		await clearSessionToken();
	}
}
