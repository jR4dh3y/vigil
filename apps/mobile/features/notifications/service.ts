import type { Event } from "@nvr/api-client";
import * as Notifications from "expo-notifications";

const ALERTS_CHANNEL = "vigil-alerts";

if (process.env.EXPO_OS !== "web") {
	Notifications.setNotificationHandler({
		handleNotification: async () => ({
			shouldPlaySound: true,
			shouldSetBadge: false,
			shouldShowBanner: true,
			shouldShowList: true,
		}),
	});
}

export async function configureNotifications(): Promise<void> {
	if (process.env.EXPO_OS !== "android") {
		return;
	}
	await Notifications.setNotificationChannelAsync(ALERTS_CHANNEL, {
		name: "Vigil alerts",
		description: "Warning and critical recorder events",
		importance: Notifications.AndroidImportance.HIGH,
		sound: "default",
	});
}

export async function requestNotificationPermission(): Promise<boolean> {
	const current = await Notifications.getPermissionsAsync();
	if (allowsNotifications(current)) {
		return true;
	}
	const requested = await Notifications.requestPermissionsAsync();
	return allowsNotifications(requested);
}

export async function notifyAboutEvent(event: Event): Promise<void> {
	await Notifications.scheduleNotificationAsync({
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
	return status.granted || status.ios?.status === Notifications.IosAuthorizationStatus.PROVISIONAL;
}
