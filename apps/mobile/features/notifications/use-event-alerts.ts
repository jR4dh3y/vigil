import type { Event } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { eventKeys, listEvents } from "@/features/events/api";
import { notifyAboutEvent } from "@/features/notifications/service";
import { getApiClient } from "@/lib/api/client";
import { useAppStore } from "@/lib/store";

export function useEventAlerts(enabled: boolean): number {
	const armed = useAppStore((state) => state.armed);
	const notificationsEnabled = useAppStore((state) => state.notificationsEnabled);
	const lastSeenEventAt = useAppStore((state) => state.lastSeenEventAt);
	const setLastSeenEventAt = useAppStore((state) => state.setLastSeenEventAt);
	const [isProcessing, setIsProcessing] = useState(false);
	// Ids already accounted for at the current watermark second. The recorder
	// timestamp has second precision, so events sharing the watermark second
	// can still arrive after it is stored and must not be dropped — nor
	// re-alerted — based on the timestamp alone.
	const seenAtWatermark = useRef<{ at: string | null; ids: Set<string> }>({
		at: null,
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
					events.filter((event) => event.createdAt === timestamp).map((event) => event.id),
				),
			};
			setLastSeenEventAt(timestamp);
		};

		const newest = events[0];
		if (!newest) {
			return;
		}
		if (!lastSeenEventAt) {
			markSeenThrough(newest.createdAt);
			return;
		}

		const lastSeenTime = Date.parse(lastSeenEventAt);
		const newestTime = Date.parse(newest.createdAt);
		if (Number.isNaN(lastSeenTime) || Number.isNaN(newestTime)) {
			markSeenThrough(newest.createdAt);
			return;
		}
		if (newestTime < lastSeenTime || isProcessing) {
			return;
		}

		if (!armed || !notificationsEnabled) {
			markSeenThrough(newest.createdAt);
			return;
		}

		const processNewEvents = async () => {
			setIsProcessing(true);
			try {
				const seenIds =
					seenAtWatermark.current.at === lastSeenEventAt
						? seenAtWatermark.current.ids
						: new Set<string>();
				const allNewEvents: Event[] = [];
				const api = getApiClient();
				let before: string | undefined;
				const limit = 100;

				while (allNewEvents.length < 1000) {
					const { data } = await api.GET("/events", {
						params: { query: { limit, before } },
					});
					if (!data?.events.length) {
						break;
					}

					const batch: Event[] = data.events;
					let reachedSeenEvents = false;
					for (const event of batch) {
						// Events arrive newest-first; the first event strictly older
						// than the watermark means every later page is accounted for.
						if (Date.parse(event.createdAt) < lastSeenTime) {
							reachedSeenEvents = true;
							break;
						}
						if (!seenIds.has(event.id)) {
							allNewEvents.push(event);
						}
					}

					if (reachedSeenEvents || batch.length < limit) {
						break;
					}

					before = batch[batch.length - 1]?.createdAt;
					if (!before) {
						break;
					}
				}

				allNewEvents.sort((a, b) => Date.parse(a.createdAt) - Date.parse(b.createdAt));
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
						if (actualNewest.createdAt === lastSeenEventAt) {
							for (const id of seenIds) {
								ids.add(id);
							}
						}
						for (const event of allNewEvents) {
							if (event.createdAt === actualNewest.createdAt) {
								ids.add(event.id);
							}
						}
						seenAtWatermark.current = { at: actualNewest.createdAt, ids };
						setLastSeenEventAt(actualNewest.createdAt);
					}
				}
			} finally {
				setIsProcessing(false);
			}
		};

		void processNewEvents();
	}, [armed, events, lastSeenEventAt, notificationsEnabled, setLastSeenEventAt, isProcessing]);

	return events.reduce((count, event) => count + (event.acknowledged ? 0 : 1), 0);
}
