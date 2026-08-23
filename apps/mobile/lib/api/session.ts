import * as SecureStore from "expo-secure-store";

const STORAGE_KEY = "nvr_session";

/** In-memory cache; `undefined` means not hydrated from SecureStore yet. */
let memoryToken: string | null | undefined;
let protectedSessionInvalidated = false;
// Bumped whenever a stored token is dropped so responses belonging to
// requests that started before the drop cannot restore the session.
let sessionGeneration = 0;
const invalidationListeners = new Set<() => void>();

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

export function getSessionGeneration(): number {
	return sessionGeneration;
}

export async function setSessionToken(token: string, requestGeneration?: number): Promise<void> {
	if (requestGeneration !== undefined && requestGeneration !== sessionGeneration) {
		// The response belongs to a request that started before the latest
		// token drop; accepting it would restore an invalidated session.
		return;
	}
	memoryToken = token;
	protectedSessionInvalidated = false;
	try {
		await SecureStore.setItemAsync(STORAGE_KEY, token);
	} catch {
		// Still keep the in-memory token for this process.
	}
}

export async function clearSessionToken(): Promise<void> {
	protectedSessionInvalidated = false;
	await deleteSessionToken();
}

export async function invalidateProtectedSession(): Promise<void> {
	if (protectedSessionInvalidated) {
		return;
	}
	protectedSessionInvalidated = true;
	await deleteSessionToken();
	for (const listener of invalidationListeners) {
		listener();
	}
}

export function subscribeToProtectedSessionInvalidation(listener: () => void): () => void {
	invalidationListeners.add(listener);
	return () => invalidationListeners.delete(listener);
}

async function deleteSessionToken(): Promise<void> {
	sessionGeneration += 1;
	memoryToken = null;
	try {
		await SecureStore.deleteItemAsync(STORAGE_KEY);
	} catch {
		// ignore
	}
}
