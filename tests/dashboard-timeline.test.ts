import { describe, expect, test } from "bun:test";
import {
	calendarGridDays,
	calendarGridRange,
	shiftCalendarMonth,
} from "../apps/dashboard/src/lib/recordings/calendar";
import {
	currentLocalDayRange,
	rangeForLocalDate,
} from "../apps/dashboard/src/lib/recordings/range";
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

	test("anchors the scale to both ends of the selected day", () => {
		const from = new Date(2026, 7, 19, 0, 0, 0, 0);
		const to = new Date(2026, 7, 19, 23, 59, 59, 999);
		const ticks = buildTimelineTicks(from, to, 1200);

		expect(ticks[0]?.at).toEqual(from);
		expect(ticks.at(-1)?.at).toEqual(to);
	});
});

describe("recording timeline browsing", () => {
	test("opens an older calendar day", () => {
		const range = rangeForLocalDate("2026-08-14");

		expect(range?.from.getFullYear()).toBe(2026);
		expect(range?.from.getMonth()).toBe(7);
		expect(range?.from.getDate()).toBe(14);
		expect(range?.to.getDate()).toBe(14);
		expect(range?.to.getHours()).toBe(23);
		expect(range?.to.getMinutes()).toBe(59);
	});

	test("defaults to midnight through the end of the current local day", () => {
		const range = currentLocalDayRange(new Date(2026, 7, 19, 14, 30));

		expect(range.from).toEqual(new Date(2026, 7, 19, 0, 0, 0, 0));
		expect(range.to).toEqual(new Date(2026, 7, 19, 23, 59, 59, 999));
	});

	test("builds a stable six-week calendar around the visible month", () => {
		const month = { year: 2026, month: 7 };
		const days = calendarGridDays(month, "2026-08-22");
		const gridRange = calendarGridRange(month);

		expect(days).toHaveLength(42);
		expect(days[0]).toMatchObject({ value: "2026-07-26", inMonth: false });
		expect(days[27]).toMatchObject({
			value: "2026-08-22",
			inMonth: true,
			isFuture: false,
		});
		expect(days[28]).toMatchObject({ value: "2026-08-23", isFuture: true });
		expect(days.at(-1)?.value).toBe("2026-09-05");
		expect(gridRange.from).toEqual(new Date(2026, 6, 26, 0, 0, 0, 0));
		expect(gridRange.to).toEqual(new Date(2026, 8, 5, 23, 59, 59, 999));
	});

	test("moves calendar months across year boundaries", () => {
		expect(shiftCalendarMonth({ year: 2026, month: 0 }, -1)).toEqual({
			year: 2025,
			month: 11,
		});
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
