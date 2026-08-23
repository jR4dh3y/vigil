import type { EventCursor } from "@nvr/api-client";

export type EventPageProgress = { kind: "complete" } | { kind: "continue"; cursor: EventCursor };

export function eventPageProgress(
	page: { events: readonly unknown[]; nextCursor?: EventCursor },
	limit: number,
	reachedWatermark: boolean,
): EventPageProgress {
	if (reachedWatermark || page.events.length < limit || !page.nextCursor) {
		return { kind: "complete" };
	}
	return { kind: "continue", cursor: page.nextCursor };
}
