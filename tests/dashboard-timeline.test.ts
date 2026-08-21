import { describe, expect, test } from "bun:test";
import {
	buildTimelineTicks,
	mergeCoverageBars,
	timelineTrackWidth,
} from "../apps/dashboard/src/lib/recordings/timeline";

describe("recording timeline scale", () => {
	test("fits the range at 100% and expands it when zoomed", () => {
		expect(timelineTrackWidth(1200, 1)).toBe(1200);
		expect(timelineTrackWidth(1200, 8)).toBe(9600);
	});

	test("adds more precise timestamp ticks as the user zooms in", () => {
		const from = new Date("2026-08-14T00:00:00Z");
		const to = new Date("2026-08-21T00:00:00Z");
		const width = timelineTrackWidth(1200, 8);
		const ticks = buildTimelineTicks(from, to, width);

		expect(ticks.length).toBeGreaterThanOrEqual(80);
		expect(ticks[1]?.at.getTime() - (ticks[0]?.at.getTime() ?? 0)).toBeLessThanOrEqual(
			2 * 60 * 60 * 1000,
		);
	});
});

test("mergeCoverageBars joins adjacent camera segments", () => {
	const merged = mergeCoverageBars([
		{ start: "2026-08-20T10:01:00Z", end: "2026-08-20T10:02:00Z" },
		{ start: "2026-08-20T10:00:00Z", end: "2026-08-20T10:01:00Z" },
		{ start: "2026-08-20T10:04:00Z", end: "2026-08-20T10:05:00Z" },
	]);

	expect(merged).toEqual([
		{ start: "2026-08-20T10:00:00.000Z", end: "2026-08-20T10:02:00.000Z" },
		{ start: "2026-08-20T10:04:00.000Z", end: "2026-08-20T10:05:00.000Z" },
	]);
});
