import type { RangePreset, TimeRange } from "./types";

const PRESET_MS: Record<RangePreset, number> = {
	"1h": 60 * 60 * 1000,
	"24h": 24 * 60 * 60 * 1000,
	"7d": 7 * 24 * 60 * 60 * 1000,
};

/** Default timeline window: last 24 hours ending at `now`. */
export function defaultTimeRange(now: Date = new Date()): TimeRange {
	return rangeForPreset("24h", now);
}

/** The local calendar day containing `now`. */
export function currentLocalDayRange(now: Date = new Date()): TimeRange {
	return localDayRange(now.getFullYear(), now.getMonth(), now.getDate());
}

/** Inclusive window ending at `now` for a preset duration. */
export function rangeForPreset(preset: RangePreset, now: Date = new Date()): TimeRange {
	const to = now;
	const from = new Date(to.getTime() - PRESET_MS[preset]);
	return { from, to };
}

/** YYYY-MM-DD in the browser's local timezone for a date input. */
export function localDateValue(date: Date): string {
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, "0");
	const day = String(date.getDate()).padStart(2, "0");
	return `${year}-${month}-${day}`;
}

/** One local calendar day, converted to the API's absolute time range. */
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
	return localDayRange(year, month, day);
}

function localDayRange(year: number, month: number, day: number): TimeRange {
	return {
		from: new Date(year, month, day),
		to: new Date(year, month, day + 1, 0, 0, 0, -1),
	};
}

export function toIso(date: Date): string {
	return date.toISOString();
}

export function parseIso(value: string): Date | null {
	const t = Date.parse(value);
	if (Number.isNaN(t)) {
		return null;
	}
	return new Date(t);
}

/** Keep playback seeks inside the selected range and at or before now. */
export function clampPlaybackTime(time: Date, range: TimeRange, now: Date = new Date()): Date {
	const upperBound = range.to.getTime() < now.getTime() ? range.to : now;
	return clampDate(time, range.from, upperBound);
}

/** Clamp a Date into [from, to]. */
export function clampDate(value: Date, from: Date, to: Date): Date {
	const t = value.getTime();
	if (t < from.getTime()) {
		return new Date(from.getTime());
	}
	if (t > to.getTime()) {
		return new Date(to.getTime());
	}
	return value;
}

/** Map a 0–1 fraction along the range to an absolute time. */
export function timeAtFraction(from: Date, to: Date, fraction: number): Date {
	const f = Math.min(1, Math.max(0, fraction));
	const ms = from.getTime() + f * (to.getTime() - from.getTime());
	return new Date(ms);
}

/** Map an absolute time to a 0–1 fraction of the range. */
export function fractionAtTime(from: Date, to: Date, time: Date): number {
	const span = to.getTime() - from.getTime();
	if (span <= 0) {
		return 0;
	}
	return Math.min(1, Math.max(0, (time.getTime() - from.getTime()) / span));
}

const timeFmt = new Intl.DateTimeFormat(undefined, {
	hour: "2-digit",
	minute: "2-digit",
	second: "2-digit",
	hour12: false,
});

const shortDateTimeFmt = new Intl.DateTimeFormat(undefined, {
	month: "short",
	day: "numeric",
	hour: "2-digit",
	minute: "2-digit",
	second: "2-digit",
	hour12: false,
});

const calendarDayFmt = new Intl.DateTimeFormat(undefined, {
	month: "short",
	day: "numeric",
	hour: "numeric",
	minute: "2-digit",
	hour12: true,
});

const dayFmt = new Intl.DateTimeFormat(undefined, {
	month: "short",
	day: "numeric",
	hour: "2-digit",
	minute: "2-digit",
	hour12: false,
});

/** Axis / selected-time label based on range length. */
export function formatTimelineLabel(date: Date, rangeMs: number): string {
	if (rangeMs <= 2 * 60 * 60 * 1000) {
		return timeFmt.format(date);
	}
	if (rangeMs <= 2 * 24 * 60 * 60 * 1000) {
		return calendarDayFmt.format(date);
	}
	return dayFmt.format(date);
}

/** Human-readable selected seek time. */
export function formatSelectedTime(date: Date): string {
	return shortDateTimeFmt.format(date);
}

export const RANGE_PRESET_LABELS: Record<RangePreset, string> = {
	"1h": "1h",
	"24h": "24h",
	"7d": "7d",
};

export const RANGE_PRESETS: RangePreset[] = ["1h", "24h", "7d"];
