import type { Event, EventCursor } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { eventKeys, listEvents } from "@/features/events/api";
import { eventPageProgress } from "@/features/notifications/event-pagination";
import {
	type EventWatermark,
	eventWatermarkAt,
	hasUnseenEventsAtWatermark,
} from "@/features/notifications/event-watermark";
import { canUseLocalNotifications } from "@/features/notifications/runtime";
import { notifyAboutEvent } from "@/features/notifications/service";
import { getApiClient } from "@/lib/api/client";
import { useAppStore } from "@/lib/store";

export function useEventAlerts(enabled: boolean): number {
	const notificationsAvailable = canUseLocalNotifications();
	const notificationsEnabled = useAppStore((state) => state.notificationsEnabled);
	const lastSeenEventAt = useAppStore((state) => state.lastSeenEventAt);
	const setLastSeenEventAt = useAppStore((state) => state.setLastSeenEventAt);
	const processing = useRef(false);
	// Ids already accounted for at the current watermark second. The recorder
	// timestamp has second precision, so events sharing the watermark second
	// can still arrive after it is stored and must not be dropped — nor
	// re-alerted — based on the timestamp alone.
	const seenAtWatermark = useRef<EventWatermark>({
		at: "",
		ids: new Set(),
	});
	const eventsQuery = useQuery({
		queryKey: eventKeys.list(false),
		queryFn: () => listEvents(false),
		enabled,
		refetchInterval: 15_000,
	});
	const events = eventsQuery.data ?? [];

	useEffect(() => {
		const markSeenThrough = (timestamp: string) => {
			seenAtWatermark.current = {
				at: timestamp,
				ids: new Set(
					events.filter((event) => event.startedAt === timestamp).map((event) => event.id),
				),
			};
			setLastSeenEventAt(timestamp);
		};

		const newest = events[0];
		if (!newest) {
			return;
		}
		if (!lastSeenEventAt) {
			markSeenThrough(newest.startedAt);
			return;
		}

		const lastSeenTime = Date.parse(lastSeenEventAt);
		const newestTime = Date.parse(newest.startedAt);
		if (Number.isNaN(lastSeenTime) || Number.isNaN(newestTime)) {
			markSeenThrough(newest.startedAt);
			return;
		}
		if (seenAtWatermark.current.at !== lastSeenEventAt) {
			seenAtWatermark.current = eventWatermarkAt(events, lastSeenEventAt);
		}
		if (
			newestTime < lastSeenTime ||
			processing.current ||
			(newestTime === lastSeenTime && !hasUnseenEventsAtWatermark(events, seenAtWatermark.current))
		) {
			return;
		}

		if (!notificationsAvailable || !notificationsEnabled) {
			markSeenThrough(newest.startedAt);
			return;
		}

		const processNewEvents = async () => {
			processing.current = true;
			try {
				const seenIds =
					seenAtWatermark.current.at === lastSeenEventAt
						? seenAtWatermark.current.ids
						: new Set<string>();
				const allNewEvents: Event[] = [];
				const api = getApiClient();
				let cursor: EventCursor | undefined;
				const limit = 100;

				while (allNewEvents.length < 1000) {
					const { data } = await api.GET("/events", {
						params: { query: { limit, cursor } },
					});
					if (!data?.events.length) {
						break;
					}

					const batch: Event[] = data.events;
					let reachedSeenEvents = false;
					for (const event of batch) {
						// Events arrive newest-first; the first event strictly older
						// than the watermark means every later page is accounted for.
						if (Date.parse(event.startedAt) < lastSeenTime) {
							reachedSeenEvents = true;
							break;
						}
						if (!seenIds.has(event.id)) {
							allNewEvents.push(event);
						}
					}

					const progress = eventPageProgress(data, limit, reachedSeenEvents);
					if (progress.kind === "complete") {
						break;
					}
					cursor = progress.cursor;
				}

				allNewEvents.sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
				const alertEvents = allNewEvents
					.filter(
						(event) =>
							!event.acknowledged &&
							(event.severity === "warning" || event.severity === "critical"),
					)
					.slice(-3);

				for (const event of alertEvents) {
					await notifyAboutEvent(event).catch(() => undefined);
				}

				if (allNewEvents.length > 0) {
					const actualNewest = allNewEvents[allNewEvents.length - 1];
					if (actualNewest) {
						const ids = new Set<string>();
						if (actualNewest.startedAt === lastSeenEventAt) {
							for (const id of seenIds) {
								ids.add(id);
							}
						}
						for (const event of allNewEvents) {
							if (event.startedAt === actualNewest.startedAt) {
								ids.add(event.id);
							}
						}
						seenAtWatermark.current = { at: actualNewest.startedAt, ids };
						setLastSeenEventAt(actualNewest.startedAt);
					}
				}
			} finally {
				processing.current = false;
			}
		};

		void processNewEvents();
	}, [events, lastSeenEventAt, notificationsAvailable, notificationsEnabled, setLastSeenEventAt]);

	return events.reduce((count, event) => count + (event.acknowledged ? 0 : 1), 0);
}
