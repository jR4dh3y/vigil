import { create } from "zustand";

type AppState = {
	armed: boolean;
	setArmed: (armed: boolean) => void;
};

/** Local UI state only — server state goes through TanStack Query. */
export const useAppStore = create<AppState>((set) => ({
	armed: true,
	setArmed: (armed) => set({ armed }),
}));
