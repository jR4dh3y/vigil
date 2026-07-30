import * as Notifications from "expo-notifications";
import { useRouter } from "expo-router";
import { useEffect } from "react";
import {
	configureNotifications,
	eventIdFromNotificationResponse,
} from "@/features/notifications/service";

export function useNotificationRouting(): void {
	const { push } = useRouter();

	useEffect(() => {
		if (process.env.EXPO_OS === "web") {
			return;
		}
		void configureNotifications().catch(() => undefined);

		const openResponse = (response: Notifications.NotificationResponse) => {
			const eventId = eventIdFromNotificationResponse(response);
			if (eventId) {
				push({ pathname: "/event/[id]", params: { id: eventId } });
			}
		};

		try {
			const lastResponse = Notifications.getLastNotificationResponse();
			if (lastResponse) {
				openResponse(lastResponse);
				Notifications.clearLastNotificationResponse();
			}
		} catch {
			// Notification response APIs are unavailable on some web runtimes.
		}

		const subscription = Notifications.addNotificationResponseReceivedListener(openResponse);
		return () => subscription.remove();
	}, [push]);
}
