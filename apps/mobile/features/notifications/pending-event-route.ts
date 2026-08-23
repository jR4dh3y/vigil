import { useSyncExternalStore } from "react";

let pendingEventId: string | null = null;
const listeners = new Set<() => void>();

export function setPendingEventRoute(eventId: string): void {
	if (pendingEventId === eventId) {
		return;
	}
	pendingEventId = eventId;
	notifyListeners();
}

export function takePendingEventRoute(): string | null {
	const eventId = pendingEventId;
	if (eventId) {
		pendingEventId = null;
		notifyListeners();
	}
	return eventId;
}

export function usePendingEventRoute(): string | null {
	return useSyncExternalStore(subscribe, getPendingEventRoute, getPendingEventRoute);
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

function getPendingEventRoute(): string | null {
	return pendingEventId;
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
