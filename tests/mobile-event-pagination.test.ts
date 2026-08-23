import { describe, expect, test } from "bun:test";
import { eventPageProgress } from "../apps/mobile/features/notifications/event-pagination";

const cursor = {
	startedAt: "2026-08-23T08:30:00Z",
	id: "4f8a3906-7fcc-458c-a0f8-19e72fa9f632",
};

describe("mobile event page progress", () => {
	test("continues with the server-provided composite cursor", () => {
		const events = Array.from({ length: 2 }, (_, index) => ({ id: String(index) }));

		expect(eventPageProgress({ events, nextCursor: cursor }, 2, false)).toEqual({
			kind: "continue",
			cursor,
		});
	});

	test("stops at the watermark or the end of the result set", () => {
		const fullPage = [{ id: "a" }, { id: "b" }];

		expect(eventPageProgress({ events: fullPage, nextCursor: cursor }, 2, true)).toEqual({
			kind: "complete",
		});
		expect(eventPageProgress({ events: fullPage }, 2, false)).toEqual({ kind: "complete" });
		expect(
			eventPageProgress({ events: fullPage.slice(0, 1), nextCursor: cursor }, 2, false),
		).toEqual({
			kind: "complete",
		});
	});
});
