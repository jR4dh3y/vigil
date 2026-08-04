import type { NotificationResponse } from "expo-notifications";
import { useRouter } from "expo-router";
import { useEffect } from "react";
import {
	loadNotificationsModule,
	type NotificationsModule,
} from "@/features/notifications/runtime";
import {
	configureNotifications,
	eventIdFromNotificationResponse,
} from "@/features/notifications/service";

export function useNotificationRouting(): void {
	const { push } = useRouter();

	useEffect(() => {
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

			const openResponse = (response: NotificationResponse) => {
				const eventId = eventIdFromNotificationResponse(response);
				if (eventId) {
					push({ pathname: "/event/[id]", params: { id: eventId } });
				}
			};

			try {
				const lastResponse = notifications.getLastNotificationResponse();
				if (lastResponse) {
					openResponse(lastResponse);
					notifications.clearLastNotificationResponse();
				}
			} catch {
				// Notification response APIs may be unavailable in a partial runtime.
			}

			if (!disposed) {
				subscription = notifications.addNotificationResponseReceivedListener(openResponse);
			}
		};

		void initialize().catch(() => undefined);
		return () => {
			disposed = true;
			subscription?.remove();
		};
	}, [push]);
}
