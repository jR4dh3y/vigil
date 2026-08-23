import { describe, expect, test } from "bun:test";
import {
	calendarGridDays,
	calendarGridRange,
	mergeCoverageBars,
	nearestCoverageTime,
	rangeForLocalDate,
	shiftCalendarMonth,
	timeAtFraction,
} from "../apps/mobile/features/recordings/date";

describe("mobile recording history dates", () => {
	test("opens one complete local calendar day", () => {
		const range = rangeForLocalDate("2026-08-23");
		expect(range?.from.getFullYear()).toBe(2026);
		expect(range?.from.getMonth()).toBe(7);
		expect(range?.from.getDate()).toBe(23);
		expect(range?.from.getHours()).toBe(0);
		expect(range?.to.getDate()).toBe(23);
		expect(range?.to.getHours()).toBe(23);
		expect(range?.to.getMilliseconds()).toBe(999);
	});

	test("rejects impossible and malformed calendar dates", () => {
		expect(rangeForLocalDate("2026-02-29")).toBeNull();
		expect(rangeForLocalDate("08/23/2026")).toBeNull();
	});

	test("keeps a stable six-week grid and moves across years", () => {
		const month = shiftCalendarMonth({ year: 2026, month: 0 }, -1);
		expect(month).toEqual({ year: 2025, month: 11 });

		const days = calendarGridDays(month, "2026-08-23");
		const range = calendarGridRange(month);
		expect(days).toHaveLength(42);
		expect(days[0]?.value).toBe("2025-11-30");
		expect(range.from.getDay()).toBe(0);
		expect(range.to.getTime() - range.from.getTime()).toBe(42 * 24 * 60 * 60 * 1000 - 1);
	});
});

describe("mobile recording history seeking", () => {
	const coverage = [
		{ start: "2026-08-23T00:00:00.000Z", end: "2026-08-23T00:01:00.000Z" },
		{ start: "2026-08-23T00:01:00.000Z", end: "2026-08-23T00:02:00.000Z" },
		{ start: "2026-08-23T00:04:00.000Z", end: "2026-08-23T00:05:00.000Z" },
	];

	test("merges adjacent coverage before rendering", () => {
		expect(mergeCoverageBars(coverage)).toEqual([
			{ start: "2026-08-23T00:00:00.000Z", end: "2026-08-23T00:02:00.000Z" },
			{ start: "2026-08-23T00:04:00.000Z", end: "2026-08-23T00:05:00.000Z" },
		]);
	});

	test("keeps an in-coverage seek and snaps a gap to the nearest recording", () => {
		const inside = new Date("2026-08-23T00:00:37.000Z");
		expect(nearestCoverageTime(coverage, inside)?.toISOString()).toBe(inside.toISOString());
		expect(nearestCoverageTime(coverage, new Date("2026-08-23T00:03:30.000Z"))?.toISOString()).toBe(
			"2026-08-23T00:04:00.000Z",
		);
	});

	test("clamps timeline fractions to the selected day", () => {
		const range = {
			from: new Date("2026-08-23T00:00:00.000Z"),
			to: new Date("2026-08-24T00:00:00.000Z"),
		};
		expect(timeAtFraction(range, -1).toISOString()).toBe(range.from.toISOString());
		expect(timeAtFraction(range, 0.5).toISOString()).toBe("2026-08-23T12:00:00.000Z");
		expect(timeAtFraction(range, 2).toISOString()).toBe(range.to.toISOString());
	});
});
