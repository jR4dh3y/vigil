import { isRunningInExpoGo } from "expo";
import type * as Notifications from "expo-notifications";

export type NotificationsModule = typeof Notifications;

let notificationsModulePromise: Promise<NotificationsModule> | null = null;

export function canUseLocalNotifications(): boolean {
	return process.env.EXPO_OS !== "web" && !isRunningInExpoGo();
}

/**
 * Expo Go on Android cannot load expo-notifications' remote notification
 * implementation. Keep the import lazy so Expo Router can still evaluate the
 * app routes when running in Expo Go.
 */
export async function loadNotificationsModule(): Promise<NotificationsModule | null> {
	if (!canUseLocalNotifications()) {
		return null;
	}

	notificationsModulePromise ??= import("expo-notifications");
	return notificationsModulePromise;
}
