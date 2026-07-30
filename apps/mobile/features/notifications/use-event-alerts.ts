import type { Event } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
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
	const eventsQuery = useQuery({
		queryKey: eventKeys.list(false),
		queryFn: () => listEvents(false),
		enabled,
		refetchInterval: 15_000,
	});
	const events = eventsQuery.data ?? [];

	useEffect(() => {
		const newest = events[0];
		if (!newest) {
			return;
		}
		if (!lastSeenEventAt) {
			setLastSeenEventAt(newest.createdAt);
			return;
		}

		const lastSeenTime = Date.parse(lastSeenEventAt);
		const newestTime = Date.parse(newest.createdAt);
		if (Number.isNaN(lastSeenTime) || Number.isNaN(newestTime)) {
			setLastSeenEventAt(newest.createdAt);
			return;
		}
		if (newestTime <= lastSeenTime || isProcessing) {
			return;
		}

		if (!armed || !notificationsEnabled) {
			setLastSeenEventAt(newest.createdAt);
			return;
		}

		const processNewEvents = async () => {
			setIsProcessing(true);
			try {
				const allNewEvents: Event[] = [];
				const api = getApiClient();
				let before: string | undefined = undefined;
				const limit = 100;

				while (allNewEvents.length < 1000) {
					const { data } = await api.GET("/events", {
						params: { query: { limit, before } },
					});
					if (!data?.events.length) {
						break;
					}

					const batch: Event[] = data.events;
					const newInBatch = batch.filter((event: Event) => Date.parse(event.createdAt) > lastSeenTime);
					allNewEvents.push(...newInBatch);

					if (newInBatch.length < batch.length) {
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
