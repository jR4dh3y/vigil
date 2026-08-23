import * as SecureStore from "expo-secure-store";
import { create } from "zustand";
import { hydrateStoredPreferences } from "@/lib/preferences";

const STORAGE_KEY = "vigil_preferences";

type AppState = {
	notificationsEnabled: boolean;
	lastSeenEventAt: string | null;
	setNotificationsEnabled: (enabled: boolean) => void;
	setLastSeenEventAt: (value: string | null) => void;
};

/** Local UI state only — server state goes through TanStack Query. */
export const useAppStore = create<AppState>((set) => ({
	notificationsEnabled: false,
	lastSeenEventAt: null,
	setNotificationsEnabled: (notificationsEnabled) => {
		set({ notificationsEnabled });
		void persistPreferences();
	},
	setLastSeenEventAt: (lastSeenEventAt) => {
		set({ lastSeenEventAt });
		void persistPreferences();
	},
}));

export async function hydrateAppPreferences(): Promise<void> {
	try {
		const stored = await SecureStore.getItemAsync(STORAGE_KEY);
		const parsed: unknown = stored ? JSON.parse(stored) : null;
		const hydrated = hydrateStoredPreferences(parsed);
		if (hydrated) {
			useAppStore.setState(hydrated.preferences);
			if (hydrated.needsWrite) {
				await persistPreferences();
			}
		}
	} catch {
		// Defaults remain usable when storage is unavailable or malformed.
	}
}

let persistQueue = Promise.resolve();

async function persistPreferences(): Promise<void> {
	persistQueue = persistQueue.then(async () => {
		const { notificationsEnabled, lastSeenEventAt } = useAppStore.getState();
		await SecureStore.setItemAsync(
			STORAGE_KEY,
			JSON.stringify({ notificationsEnabled, lastSeenEventAt }),
		).catch(() => undefined);
	});
	return persistQueue;
}
