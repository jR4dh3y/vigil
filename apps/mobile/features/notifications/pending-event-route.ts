import { useSyncExternalStore } from "react";
import { getApiBaseUrl } from "@/lib/api/config-state";

let pendingEventRoute: { eventId: string; recorderBaseUrl: string } | null = null;
const listeners = new Set<() => void>();

/** The route is scoped to the recorder it was raised for so switching
 * recorders cannot open a stale event against a different recorder. */
export function setPendingEventRoute(eventId: string): void {
	const recorderBaseUrl = getApiBaseUrl();
	if (
		pendingEventRoute?.eventId === eventId &&
		pendingEventRoute.recorderBaseUrl === recorderBaseUrl
	) {
		return;
	}
	pendingEventRoute = { eventId, recorderBaseUrl };
	notifyListeners();
}

export function takePendingEventRoute(): string | null {
	const eventId = currentPendingEventId();
	if (pendingEventRoute) {
		pendingEventRoute = null;
		notifyListeners();
	}
	return eventId;
}

export function usePendingEventRoute(): string | null {
	return useSyncExternalStore(subscribe, getPendingEventRoute, getPendingEventRoute);
}

function getPendingEventRoute(): string | null {
	return currentPendingEventId();
}

function currentPendingEventId(): string | null {
	if (!pendingEventRoute || pendingEventRoute.recorderBaseUrl !== getApiBaseUrl()) {
		return null;
	}
	return pendingEventRoute.eventId;
}

export function eventIdFromPathname(pathname: string): string | null {
	const match = /\/event\/([^/]+)\/?$/.exec(pathname);
	if (!match?.[1]) {
		return null;
	}
	try {
		return decodeURIComponent(match[1]);
	} catch {
		return null;
	}
}

function subscribe(listener: () => void): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}

function notifyListeners(): void {
	for (const listener of listeners) {
		listener();
	}
}
