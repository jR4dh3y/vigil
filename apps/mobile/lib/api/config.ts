import * as SecureStore from "expo-secure-store";
import {
	type ApiConfiguration,
	createInitialApiConfiguration,
	normalizeApiBaseUrl,
	resolveMediaUrlAgainstBase,
} from "@/lib/api/config-values";

export {
	MediaUrlError,
	normalizeApiBaseUrl,
	resolveMediaUrlAgainstBase,
} from "@/lib/api/config-values";

const STORAGE_KEY = "vigil_recorder_url";
const bundledApiUrl = process.env.EXPO_PUBLIC_API_URL?.trim();

let activeConfiguration = createInitialApiConfiguration(bundledApiUrl);
const listeners = new Set<() => void>();

export async function hydrateApiConfig(): Promise<void> {
	try {
		const stored = await SecureStore.getItemAsync(STORAGE_KEY);
		if (stored) {
			updateActiveConfiguration({ kind: "configured", baseUrl: normalizeApiBaseUrl(stored) });
		}
	} catch {
		// Keep the bundled or unconfigured state when secure storage is unavailable or invalid.
	}
}

export function getApiBaseUrl(): string {
	return activeConfiguration.baseUrl;
}

export function getApiConfiguration(): ApiConfiguration {
	return activeConfiguration;
}

export async function saveApiBaseUrl(value: string): Promise<string> {
	const normalized = normalizeApiBaseUrl(value);
	await SecureStore.setItemAsync(STORAGE_KEY, normalized);
	updateActiveConfiguration({ kind: "configured", baseUrl: normalized });
	return normalized;
}

export function subscribeToApiBaseUrl(listener: () => void): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}

function updateActiveConfiguration(configuration: ApiConfiguration): void {
	if (
		configuration.kind === activeConfiguration.kind &&
		configuration.baseUrl === activeConfiguration.baseUrl
	) {
		return;
	}
	activeConfiguration = configuration;
	for (const listener of listeners) {
		listener();
	}
}

export function resolveMediaUrl(path: string, token?: string): string {
	return resolveMediaUrlAgainstBase(path, activeConfiguration.baseUrl, token);
}

/** Remove the rotating token to get a stable stream endpoint identity. */
export function streamEndpoint(uri: string): string {
	const url = new URL(uri, `${new URL(activeConfiguration.baseUrl).origin}/`);
	url.searchParams.delete("token");
	return url.toString();
}
