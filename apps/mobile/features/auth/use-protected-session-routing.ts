import { usePathname, useRouter } from "expo-router";
import { useEffect, useRef } from "react";
import {
	eventIdFromPathname,
	setPendingEventRoute,
} from "@/features/notifications/pending-event-route";
import { subscribeToProtectedSessionInvalidation } from "@/lib/api/session";
import { queryClient } from "@/lib/query-client";

export function useProtectedSessionRouting(): void {
	const pathname = usePathname();
	const router = useRouter();
	const pathnameRef = useRef(pathname);
	pathnameRef.current = pathname;

	useEffect(() => {
		return subscribeToProtectedSessionInvalidation(() => {
			const eventId = eventIdFromPathname(pathnameRef.current);
			if (eventId) {
				setPendingEventRoute(eventId);
			}
			queryClient.clear();
			router.replace("/login");
		});
	}, [router]);
}
