import { useSyncExternalStore } from "react";
import { getApiBaseUrl, subscribeToApiBaseUrl } from "@/lib/api/config";

export function useApiBaseUrl(): string {
	return useSyncExternalStore(subscribeToApiBaseUrl, getApiBaseUrl, getApiBaseUrl);
}
