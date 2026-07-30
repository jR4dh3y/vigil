import * as SecureStore from "expo-secure-store";
import { create } from "zustand";

const STORAGE_KEY = "vigil_preferences";

type AppState = {
	armed: boolean;
	notificationsEnabled: boolean;
	lastSeenEventAt: string | null;
	setArmed: (armed: boolean) => void;
	setNotificationsEnabled: (enabled: boolean) => void;
	setLastSeenEventAt: (value: string | null) => void;
};

/** Local UI state only — server state goes through TanStack Query. */
export const useAppStore = create<AppState>((set) => ({
	armed: true,
	notificationsEnabled: false,
	lastSeenEventAt: null,
	setArmed: (armed) => {
		set({ armed });
		void persistPreferences();
	},
	setNotificationsEnabled: (notificationsEnabled) => {
		set({ notificationsEnabled });
		void persistPreferences();
	},
	setLastSeenEventAt: (lastSeenEventAt) => {
		set({ lastSeenEventAt });
		void persistPreferences();
	},
}));

type PersistedPreferences = Pick<AppState, "armed" | "notificationsEnabled" | "lastSeenEventAt">;

function isPersistedPreferences(value: unknown): value is PersistedPreferences {
	if (typeof value !== "object" || value === null) {
		return false;
	}
	const candidate = value as Record<string, unknown>;
	return (
		typeof candidate.armed === "boolean" &&
		typeof candidate.notificationsEnabled === "boolean" &&
		(candidate.lastSeenEventAt === null || typeof candidate.lastSeenEventAt === "string")
	);
}

export async function hydrateAppPreferences(): Promise<void> {
	try {
		const stored = await SecureStore.getItemAsync(STORAGE_KEY);
		const parsed: unknown = stored ? JSON.parse(stored) : null;
		if (isPersistedPreferences(parsed)) {
			useAppStore.setState(parsed);
		}
	} catch {
		// Defaults remain usable when storage is unavailable or malformed.
	}
}

let persistQueue = Promise.resolve();

async function persistPreferences(): Promise<void> {
	persistQueue = persistQueue.then(async () => {
		const { armed, notificationsEnabled, lastSeenEventAt } = useAppStore.getState();
		await SecureStore.setItemAsync(
			STORAGE_KEY,
			JSON.stringify({ armed, notificationsEnabled, lastSeenEventAt }),
		).catch(() => undefined);
	});
	return persistQueue;
}
