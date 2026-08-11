import * as SecureStore from "expo-secure-store";

const STORAGE_KEY = "vigil_recorder_url";
const DEFAULT_API_URL = "http://127.0.0.1:8080/api/v1";
const configuredApiUrl = process.env.EXPO_PUBLIC_API_URL ?? DEFAULT_API_URL;

let activeApiBaseUrl = normalizeApiBaseUrl(configuredApiUrl);
const listeners = new Set<() => void>();

export function normalizeApiBaseUrl(value: string): string {
	const candidate = value.trim();
	if (!candidate) {
		throw new Error("Enter the address of your recorder");
	}

	const withProtocol = /^https?:\/\//i.test(candidate) ? candidate : `http://${candidate}`;
	const url = new URL(withProtocol);
	if (url.protocol !== "http:" && url.protocol !== "https:") {
		throw new Error("Recorder address must use HTTP or HTTPS");
	}
	if (url.username || url.password) {
		throw new Error("Recorder address must not contain credentials");
	}

	url.search = "";
	url.hash = "";
	const path = url.pathname.replace(/\/+$/, "");
	url.pathname = path.endsWith("/api/v1") ? path : `${path}/api/v1`;
	return url.toString().replace(/\/$/, "");
}

export async function hydrateApiConfig(): Promise<void> {
	try {
		const stored = await SecureStore.getItemAsync(STORAGE_KEY);
		if (stored) {
			updateActiveUrl(normalizeApiBaseUrl(stored));
		}
	} catch {
		// Keep the bundled URL when secure storage is unavailable or invalid.
	}
}

export function getApiBaseUrl(): string {
	return activeApiBaseUrl;
}

export async function saveApiBaseUrl(value: string): Promise<string> {
	const normalized = normalizeApiBaseUrl(value);
	await SecureStore.setItemAsync(STORAGE_KEY, normalized);
	updateActiveUrl(normalized);
	return normalized;
}

export function subscribeToApiBaseUrl(listener: () => void): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}

function updateActiveUrl(value: string): void {
	if (value === activeApiBaseUrl) {
		return;
	}
	activeApiBaseUrl = value;
	for (const listener of listeners) {
		listener();
	}
}

export function resolveMediaUrl(path: string, token?: string): string {
	const baseOrigin = new URL(activeApiBaseUrl).origin;
	const url = new URL(path, `${baseOrigin}/`);
	if (token) {
		url.searchParams.set("token", token);
	}
	return url.toString();
}

/** Remove the rotating token to get a stable stream endpoint identity. */
export function streamEndpoint(uri: string): string {
	const url = new URL(uri, `${new URL(activeApiBaseUrl).origin}/`);
	url.searchParams.delete("token");
	return url.toString();
}
