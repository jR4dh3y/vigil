import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { eventKeys, listEvents } from "@/features/events/api";
import { notifyAboutEvent } from "@/features/notifications/service";
import { useAppStore } from "@/lib/store";

export function useEventAlerts(enabled: boolean): number {
	const armed = useAppStore((state) => state.armed);
	const notificationsEnabled = useAppStore((state) => state.notificationsEnabled);
	const lastSeenEventAt = useAppStore((state) => state.lastSeenEventAt);
	const setLastSeenEventAt = useAppStore((state) => state.setLastSeenEventAt);
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
		if (newestTime <= lastSeenTime) {
			return;
		}
		const newEvents = events
			.filter((event) => Date.parse(event.createdAt) > lastSeenTime)
			.slice()
			.reverse();
		setLastSeenEventAt(newest.createdAt);

		if (!armed || !notificationsEnabled) {
			return;
		}
		const alertEvents = newEvents
			.filter(
				(event) =>
					!event.acknowledged && (event.severity === "warning" || event.severity === "critical"),
			)
			.slice(-3);
		for (const event of alertEvents) {
			void notifyAboutEvent(event).catch(() => undefined);
		}
	}, [armed, events, lastSeenEventAt, notificationsEnabled, setLastSeenEventAt]);

	return events.reduce((count, event) => count + (event.acknowledged ? 0 : 1), 0);
}
