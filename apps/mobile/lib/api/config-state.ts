import {
	type ApiConfiguration,
	createInitialApiConfiguration,
	resolveMediaUrlAgainstBase,
} from "./config-values";

const bundledApiUrl = process.env.EXPO_PUBLIC_API_URL?.trim();

let activeConfiguration = createInitialApiConfiguration(bundledApiUrl);
const listeners = new Set<() => void>();

export function getApiBaseUrl(): string {
	return activeConfiguration.baseUrl;
}

export function getApiConfiguration(): ApiConfiguration {
	return activeConfiguration;
}

export function subscribeToApiBaseUrl(listener: () => void): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}

/** Swap the in-memory configuration and notify subscribers. */
export function updateActiveConfiguration(configuration: ApiConfiguration): void {
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
