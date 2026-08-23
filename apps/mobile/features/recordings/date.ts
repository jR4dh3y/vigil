import type { CoverageBar } from "@nvr/api-client";

export type TimeRange = {
	from: Date;
	to: Date;
};

export type CalendarMonth = {
	year: number;
	month: number;
};

export type CalendarDay = {
	value: string;
	day: number;
	inMonth: boolean;
	isFuture: boolean;
};

export function localDateValue(date: Date): string {
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, "0");
	const day = String(date.getDate()).padStart(2, "0");
	return `${year}-${month}-${day}`;
}

export function rangeForLocalDate(value: string): TimeRange | null {
	const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
	if (!match) {
		return null;
	}

	const year = Number(match[1]);
	const month = Number(match[2]) - 1;
	const day = Number(match[3]);
	const from = new Date(year, month, day);
	if (from.getFullYear() !== year || from.getMonth() !== month || from.getDate() !== day) {
		return null;
	}

	return {
		from,
		to: new Date(year, month, day + 1, 0, 0, 0, -1),
	};
}

export function calendarMonthForDate(date: Date): CalendarMonth {
	return { year: date.getFullYear(), month: date.getMonth() };
}

export function calendarMonthForValue(value: string): CalendarMonth | null {
	const range = rangeForLocalDate(value);
	return range ? calendarMonthForDate(range.from) : null;
}

export function shiftCalendarMonth(month: CalendarMonth, offset: number): CalendarMonth {
	return calendarMonthForDate(new Date(month.year, month.month + offset, 1));
}

export function calendarGridDays(month: CalendarMonth, maxDate: string): CalendarDay[] {
	const first = new Date(month.year, month.month, 1);
	const start = new Date(month.year, month.month, 1 - first.getDay());
	return Array.from({ length: 42 }, (_, index) => {
		const date = new Date(start.getFullYear(), start.getMonth(), start.getDate() + index);
		const value = localDateValue(date);
		return {
			value,
			day: date.getDate(),
			inMonth: date.getMonth() === month.month && date.getFullYear() === month.year,
			isFuture: value > maxDate,
		};
	});
}

export function calendarGridRange(month: CalendarMonth): TimeRange {
	const first = new Date(month.year, month.month, 1);
	const from = new Date(month.year, month.month, 1 - first.getDay());
	return {
		from,
		to: new Date(from.getFullYear(), from.getMonth(), from.getDate() + 42, 0, 0, 0, -1),
	};
}

export function fractionAtTime(range: TimeRange, time: Date): number {
	const span = range.to.getTime() - range.from.getTime();
	if (span <= 0) {
		return 0;
	}
	return Math.min(1, Math.max(0, (time.getTime() - range.from.getTime()) / span));
}

export function timeAtFraction(range: TimeRange, fraction: number): Date {
	const clamped = Math.min(1, Math.max(0, fraction));
	return new Date(range.from.getTime() + clamped * (range.to.getTime() - range.from.getTime()));
}

/** Pick a playable instant, preferring the requested point and then the next recording. */
export function nearestCoverageTime(
	coverage: readonly CoverageBar[],
	requested: Date,
): Date | null {
	const requestedTime = requested.getTime();
	let previousEnd: number | null = null;

	for (const bar of coverage) {
		const start = Date.parse(bar.start);
		const end = Date.parse(bar.end);
		if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) {
			continue;
		}
		if (requestedTime >= start && requestedTime < end) {
			return requested;
		}
		if (requestedTime < start) {
			if (previousEnd === null || start - requestedTime <= requestedTime - previousEnd) {
				return new Date(start);
			}
			return new Date(Math.max(previousEnd - 1, 0));
		}
		previousEnd = end;
	}

	return previousEnd === null ? null : new Date(Math.max(previousEnd - 1, 0));
}

export function mergeCoverageBars(coverage: readonly CoverageBar[]): CoverageBar[] {
	const sorted = coverage
		.map((bar) => ({ start: Date.parse(bar.start), end: Date.parse(bar.end) }))
		.filter((bar) => Number.isFinite(bar.start) && Number.isFinite(bar.end) && bar.end > bar.start)
		.sort((left, right) => left.start - right.start);
	const merged: Array<{ start: number; end: number }> = [];

	for (const bar of sorted) {
		const last = merged[merged.length - 1];
		if (last && bar.start <= last.end) {
			last.end = Math.max(last.end, bar.end);
		} else {
			merged.push({ ...bar });
		}
	}

	return merged.map((bar) => ({
		start: new Date(bar.start).toISOString(),
		end: new Date(bar.end).toISOString(),
	}));
}

export function resolvedTimeZone(): string {
	return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
}
