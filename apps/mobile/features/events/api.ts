import type { Event } from "@nvr/api-client";
import { api } from "@/lib/api/client";
import { throwApiError } from "@/lib/api/error";

export async function listEvents(unacknowledgedOnly: boolean): Promise<Event[]> {
	const { data, error, response } = await api.GET("/events", {
		params: { query: { limit: 100, unacknowledgedOnly: unacknowledgedOnly || undefined } },
	});
	if (data) {
		return data.events;
	}
	throwApiError(error, response.status, "Could not load events");
}

export async function acknowledgeEvent(id: string): Promise<Event> {
	const { data, error, response } = await api.POST("/events/{id}/acknowledge", {
		params: { path: { id } },
	});
	if (data) {
		return data;
	}
	throwApiError(error, response.status, "Could not acknowledge event");
}

export const eventKeys = {
	all: ["events"] as const,
	list: (unacknowledgedOnly: boolean) => ["events", { unacknowledgedOnly }] as const,
};
