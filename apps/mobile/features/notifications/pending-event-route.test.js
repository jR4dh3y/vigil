import { describe, expect, test } from "bun:test";
import {
	eventIdFromPathname,
	setPendingEventRoute,
	takePendingEventRoute,
} from "@/features/notifications/pending-event-route";

describe("pending event route", () => {
	test("extracts only event detail paths", () => {
		expect(eventIdFromPathname("/event/event%20123")).toBe("event 123");
		expect(eventIdFromPathname("/camera/event-123")).toBeNull();
	});

	test("hands a pending event to authentication once", () => {
		setPendingEventRoute("event-123");
		expect(takePendingEventRoute()).toBe("event-123");
		expect(takePendingEventRoute()).toBeNull();
	});
});
