import type { Event } from "@nvr/api-client";
import { api } from "$lib/api";
import type { ListEventsParams } from "./types";

export class EventApiError extends Error {
	readonly status: number;
	readonly code?: string;

	constructor(message: string, status: number, code?: string) {
		super(message);
		this.name = "EventApiError";
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

function throwEventError(body: unknown, status: number, fallback: string): never {
	const { message, code } = errorDetails(body, fallback);
	throw new EventApiError(message, status, code);
}

/** GET /events */
export async function listEvents(params: ListEventsParams = {}): Promise<Event[]> {
	const { data, error, response } = await api.GET("/events", {
		params: {
			query: {
				limit: params.limit,
				before: params.before,
				cameraId: params.cameraId,
				type: params.type,
				unacknowledgedOnly: params.unacknowledgedOnly,
			},
		},
	});
	if (data) {
		return data.events;
	}
	throwEventError(error, response.status, "Failed to load events");
}

/** POST /events/{id}/acknowledge */
export async function acknowledgeEvent(id: string): Promise<Event> {
	const { data, error, response } = await api.POST("/events/{id}/acknowledge", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwEventError(error, response.status, "Failed to acknowledge event");
}
