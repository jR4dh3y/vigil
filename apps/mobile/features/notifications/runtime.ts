import { isRunningInExpoGo } from "expo";
import type * as Notifications from "expo-notifications";
import { Platform } from "react-native";

export type NotificationsModule = typeof Notifications;

let notificationsModulePromise: Promise<NotificationsModule> | null = null;

/**
 * Expo Go on Android cannot load expo-notifications' remote notification
 * implementation. Keep the import lazy so Expo Router can still evaluate the
 * app routes when running in Expo Go.
 */
export async function loadNotificationsModule(): Promise<NotificationsModule | null> {
	if (Platform.OS === "web" || isRunningInExpoGo()) {
		return null;
	}

	notificationsModulePromise ??= import("expo-notifications");
	return notificationsModulePromise;
}
