import * as SecureStore from "expo-secure-store";

const STORAGE_KEY = "nvr_session";

/** In-memory cache; `undefined` means not hydrated from SecureStore yet. */
let memoryToken: string | null | undefined;

export async function hydrateSession(): Promise<void> {
	if (memoryToken !== undefined) {
		return;
	}
	try {
		memoryToken = await SecureStore.getItemAsync(STORAGE_KEY);
	} catch {
		memoryToken = null;
	}
}

export async function getSessionToken(): Promise<string | null> {
	await hydrateSession();
	return memoryToken ?? null;
}

export async function setSessionToken(token: string): Promise<void> {
	memoryToken = token;
	try {
		await SecureStore.setItemAsync(STORAGE_KEY, token);
	} catch {
		// Still keep the in-memory token for this process.
	}
}

export async function clearSessionToken(): Promise<void> {
	memoryToken = null;
	try {
		await SecureStore.deleteItemAsync(STORAGE_KEY);
	} catch {
		// ignore
	}
}
