import { create } from "zustand";

type AppState = {
	armed: boolean;
	notificationsEnabled: boolean;
	setArmed: (armed: boolean) => void;
	setNotificationsEnabled: (enabled: boolean) => void;
};

/** Local UI state only — server state goes through TanStack Query. */
export const useAppStore = create<AppState>((set) => ({
	armed: true,
	notificationsEnabled: true,
	setArmed: (armed) => set({ armed }),
	setNotificationsEnabled: (notificationsEnabled) => set({ notificationsEnabled }),
}));
