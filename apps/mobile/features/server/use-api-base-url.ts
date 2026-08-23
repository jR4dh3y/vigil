import { useSyncExternalStore } from "react";
import {
	getApiBaseUrl,
	getApiConfiguration,
	subscribeToApiBaseUrl,
} from "@/lib/api/config";

export function useApiBaseUrl(): string {
	return useSyncExternalStore(subscribeToApiBaseUrl, getApiBaseUrl, getApiBaseUrl);
}

export function useApiConfiguration() {
	return useSyncExternalStore(
		subscribeToApiBaseUrl,
		getApiConfiguration,
		getApiConfiguration,
	);
}
