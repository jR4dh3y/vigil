import type { CreateUserRequest, UserPublic } from "@nvr/api-client";
import { getApiClient } from "$lib/connection";

export class UserApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "UserApiError";
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

function throwUserError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new UserApiError(message, status, code);
}

/** GET /users — admin only. */
export async function listUsers(): Promise<UserPublic[]> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/users");
	if (data) {
		return data.users;
	}
	throwUserError(error, response.status, "Failed to load users");
}

/** POST /users — admin only. */
export async function createUser(body: CreateUserRequest): Promise<UserPublic> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/users", { body });
	if (data) {
		return data;
	}
	throwUserError(error, response.status, "Failed to create user");
}

/** DELETE /users/{id} — admin only. */
export async function deleteUser(id: string): Promise<void> {
	const api = getApiClient();
	const { error, response } = await api.DELETE("/users/{id}", {
		params: { path: { id } },
	});
	if (response.ok || response.status === 204) {
		return;
	}
	throwUserError(error, response.status, "Failed to delete user");
}
