import type { NotificationResponse } from "expo-notifications";
import { useRouter } from "expo-router";
import { useEffect } from "react";
import { setPendingEventRoute } from "@/features/notifications/pending-event-route";
import {
	loadNotificationsModule,
	type NotificationsModule,
} from "@/features/notifications/runtime";
import {
	configureNotifications,
	eventIdFromNotificationResponse,
} from "@/features/notifications/service";
import { getApiConfiguration } from "@/lib/api/config";
import { getSessionToken } from "@/lib/api/session";

export function useNotificationRouting(enabled: boolean): void {
	const { push } = useRouter();

	useEffect(() => {
		if (!enabled) {
			return;
		}
		let disposed = false;
		let subscription: ReturnType<
			NotificationsModule["addNotificationResponseReceivedListener"]
		> | null = null;

		const initialize = async () => {
			const notifications = await loadNotificationsModule();
			if (disposed || !notifications) {
				return;
			}

			await configureNotifications().catch(() => undefined);

			const openResponse = async (response: NotificationResponse) => {
				const eventId = eventIdFromNotificationResponse(response);
				if (!eventId) {
					return;
				}
				const token = await getSessionToken();
				if (disposed) {
					return;
				}
				if (getApiConfiguration().kind === "configured" && token) {
					push({ pathname: "/event/[id]", params: { id: eventId } });
					return;
				}
				setPendingEventRoute(eventId);
				push("/");
			};

			try {
				const lastResponse = notifications.getLastNotificationResponse();
				if (lastResponse) {
					void openResponse(lastResponse);
					notifications.clearLastNotificationResponse();
				}
			} catch {
				// Notification response APIs may be unavailable in a partial runtime.
			}

			if (!disposed) {
				subscription = notifications.addNotificationResponseReceivedListener((response) => {
					void openResponse(response);
				});
			}
		};

		void initialize().catch(() => undefined);
		return () => {
			disposed = true;
			subscription?.remove();
		};
	}, [enabled, push]);
}
