import type { Event } from "@nvr/api-client";
import { getApiClient } from "@/lib/api/client";
import { ApiError, throwApiError } from "@/lib/api/error";

export async function listEvents(unacknowledgedOnly: boolean, limit = 100): Promise<Event[]> {
	const api = getApiClient();
	const { data, error, response } = await api.GET("/events", {
		params: { query: { limit, unacknowledgedOnly: unacknowledgedOnly || undefined } },
	});
	if (data) {
		return data.events;
	}
	throwApiError(error, response.status, "Could not load events");
}

export async function acknowledgeEvent(id: string): Promise<Event> {
	const api = getApiClient();
	const { data, error, response } = await api.POST("/events/{id}/acknowledge", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not acknowledge event");
}

export async function getEvent(id: string): Promise<Event> {
	let before: string | undefined;
	const limit = 200;
	const maxPages = 50;
	let pagesChecked = 0;

	while (pagesChecked < maxPages) {
		const api = getApiClient();
		const { data, error, response } = await api.GET("/events", {
			params: { query: { limit, before } },
		});
		if (!data) {
			throwApiError(error, response.status, "Could not load events");
		}

		const events: Event[] = data.events;
		const event = events.find((item: Event) => item.id === id);
		if (event) {
			return event;
		}

		if (events.length < limit) {
			break;
		}

		before = events[events.length - 1]?.createdAt;
		if (!before) {
			break;
		}
		pagesChecked++;
	}

	throw new ApiError("This event is no longer in recent history", 404);
}

export const eventKeys = {
	all: ["events"] as const,
	list: (unacknowledgedOnly: boolean) => ["events", { unacknowledgedOnly }] as const,
	detail: (id: string) => ["events", id] as const,
};
