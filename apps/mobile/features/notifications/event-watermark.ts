import type { Event } from "@nvr/api-client";

export type EventWatermark = {
	at: string;
	ids: Set<string>;
};

export function eventWatermarkAt(
	events: readonly Pick<Event, "id" | "startedAt">[],
	timestamp: string,
): EventWatermark {
	return {
		at: timestamp,
		ids: new Set(events.filter((event) => event.startedAt === timestamp).map((event) => event.id)),
	};
}

export function hasUnseenEventsAtWatermark(
	events: readonly Pick<Event, "id" | "startedAt">[],
	watermark: EventWatermark,
): boolean {
	return events.some((event) => event.startedAt === watermark.at && !watermark.ids.has(event.id));
}
