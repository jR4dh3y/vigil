import { useSyncExternalStore } from "react";
import { getApiBaseUrl, subscribeToApiBaseUrl } from "@/lib/api/config-state";

let pendingEventRoute: { eventId: string; recorderBaseUrl: string } | null = null;
const listeners = new Set<() => void>();

/** The route is scoped to the recorder it was raised for and dropped as soon
 * as that recorder stops being active, so switching recorders can neither open
 * a stale event against another recorder nor resurrect it when switching back. */
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

/** Observe the pending event without consuming it. */
export function peekPendingEventRoute(): string | null {
	return currentPendingEventId();
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

// A route kept while its recorder is inactive must never survive a recorder
// switch, so drop it the moment the active configuration moves elsewhere.
subscribeToApiBaseUrl(() => {
	if (!pendingEventRoute || pendingEventRoute.recorderBaseUrl === getApiBaseUrl()) {
		return;
	}
	pendingEventRoute = null;
	notifyListeners();
});

function notifyListeners(): void {
	for (const listener of listeners) {
		listener();
	}
}
