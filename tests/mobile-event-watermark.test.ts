import { describe, expect, test } from "bun:test";
import {
	eventWatermarkAt,
	hasUnseenEventsAtWatermark,
} from "../apps/mobile/features/notifications/event-watermark";

const watermarkTime = "2026-08-23T08:30:00Z";

describe("mobile event watermark", () => {
	test("accounts for existing events when a persisted timestamp is hydrated", () => {
		const events = [
			{ id: "existing-a", startedAt: watermarkTime },
			{ id: "existing-b", startedAt: watermarkTime },
		];

		const watermark = eventWatermarkAt(events, watermarkTime);

		expect(hasUnseenEventsAtWatermark(events, watermark)).toBe(false);
	});

	test("detects an event that arrives later in the same timestamp second", () => {
		const watermark = eventWatermarkAt(
			[{ id: "existing", startedAt: watermarkTime }],
			watermarkTime,
		);
		const updated = [
			{ id: "late", startedAt: watermarkTime },
			{ id: "existing", startedAt: watermarkTime },
		];

		expect(hasUnseenEventsAtWatermark(updated, watermark)).toBe(true);
	});
});
