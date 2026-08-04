import type { Event } from "@nvr/api-client";
import type * as Notifications from "expo-notifications";
import { loadNotificationsModule } from "@/features/notifications/runtime";

const ALERTS_CHANNEL = "vigil-alerts";
const IOS_PROVISIONAL_AUTHORIZATION_STATUS: Notifications.IosAuthorizationStatus = 3;

let notificationHandlerConfigured = false;

export async function configureNotifications(): Promise<void> {
	const notifications = await loadNotificationsModule();
	if (!notifications) {
		return;
	}

	if (!notificationHandlerConfigured) {
		notifications.setNotificationHandler({
			handleNotification: async () => ({
				shouldPlaySound: true,
				shouldSetBadge: false,
				shouldShowBanner: true,
				shouldShowList: true,
			}),
		});
		notificationHandlerConfigured = true;
	}

	if (process.env.EXPO_OS !== "android") {
		return;
	}

	await notifications.setNotificationChannelAsync(ALERTS_CHANNEL, {
		name: "Vigil alerts",
		description: "Warning and critical recorder events",
		importance: notifications.AndroidImportance.HIGH,
		sound: "default",
	});
}

export async function requestNotificationPermission(): Promise<boolean> {
	const notifications = await loadNotificationsModule();
	if (!notifications) {
		return false;
	}

	const current = await notifications.getPermissionsAsync();
	if (allowsNotifications(current)) {
		return true;
	}
	const requested = await notifications.requestPermissionsAsync();
	return allowsNotifications(requested);
}

export async function notifyAboutEvent(event: Event): Promise<void> {
	const notifications = await loadNotificationsModule();
	if (!notifications) {
		return;
	}

	await notifications.scheduleNotificationAsync({
		content: {
			title: event.title,
			body: event.message,
			data: { eventId: event.id },
			sound: "default",
		},
		trigger: process.env.EXPO_OS === "android" ? { channelId: ALERTS_CHANNEL } : null,
	});
}

export function eventIdFromNotificationResponse(
	response: Notifications.NotificationResponse,
): string | null {
	const eventId = response.notification.request.content.data?.eventId;
	return typeof eventId === "string" && eventId.length > 0 ? eventId : null;
}

function allowsNotifications(
	status: Awaited<ReturnType<typeof Notifications.getPermissionsAsync>>,
): boolean {
	return status.granted || status.ios?.status === IOS_PROVISIONAL_AUTHORIZATION_STATUS;
}
