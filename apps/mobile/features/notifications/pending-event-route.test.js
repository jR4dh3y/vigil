import { describe, expect, test } from "bun:test";
import {
	eventIdFromPathname,
	setPendingEventRoute,
	takePendingEventRoute,
} from "@/features/notifications/pending-event-route";
import { updateActiveConfiguration } from "@/lib/api/config-state";

function useRecorder(baseUrl) {
	updateActiveConfiguration({ kind: "configured", baseUrl });
}

describe("pending event route", () => {
	test("extracts only event detail paths", () => {
		expect(eventIdFromPathname("/event/event%20123")).toBe("event 123");
		expect(eventIdFromPathname("/camera/event-123")).toBeNull();
	});

	test("hands a pending event to authentication once", () => {
		useRecorder("http://recorder-a.example/api/v1");
		setPendingEventRoute("event-123");
		expect(takePendingEventRoute()).toBe("event-123");
		expect(takePendingEventRoute()).toBeNull();
	});

	test("drops a pending event when the recorder changes", () => {
		useRecorder("http://recorder-a.example/api/v1");
		setPendingEventRoute("event-123");
		useRecorder("http://recorder-b.example/api/v1");
		expect(takePendingEventRoute()).toBeNull();
	});
});
