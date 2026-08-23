import * as SecureStore from "expo-secure-store";
import { updateActiveConfiguration } from "./config-state";
import { normalizeApiBaseUrl } from "./config-values";

export {
	getApiBaseUrl,
	getApiConfiguration,
	resolveMediaUrl,
	streamEndpoint,
	subscribeToApiBaseUrl,
} from "./config-state";
export {
	MediaUrlError,
	normalizeApiBaseUrl,
	resolveMediaUrlAgainstBase,
} from "./config-values";

const STORAGE_KEY = "vigil_recorder_url";

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

export async function saveApiBaseUrl(value: string): Promise<string> {
	const normalized = normalizeApiBaseUrl(value);
	await SecureStore.setItemAsync(STORAGE_KEY, normalized);
	updateActiveConfiguration({ kind: "configured", baseUrl: normalized });
	return normalized;
}
