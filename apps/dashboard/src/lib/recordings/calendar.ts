import { localDateValue, rangeForLocalDate } from "./range";
import type { TimeRange } from "./types";

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

/** Six complete weeks keep the calendar popover stable between months. */
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
