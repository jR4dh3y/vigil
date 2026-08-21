import { formatTimelineLabel, fractionAtTime } from "./range";
import type { CoverageBar } from "./types";

const MINUTE_MS = 60 * 1000;
const MIN_TICK_SPACING_PX = 92;
const HOUR_MS = 60 * MINUTE_MS;

const TICK_INTERVALS_MS = [
	MINUTE_MS,
	5 * MINUTE_MS,
	15 * MINUTE_MS,
	30 * MINUTE_MS,
	HOUR_MS,
	2 * HOUR_MS,
	3 * HOUR_MS,
	6 * HOUR_MS,
	12 * HOUR_MS,
	24 * HOUR_MS,
] as const;

export type TimelineBarRect = {
	start: string;
	end: string;
	left: number;
	width: number;
};

export type TimelineTick = {
	at: Date;
	fraction: number;
	label: string;
};

export function timelineTrackWidth(viewportWidth: number, zoom: number): number {
	return Math.max(1, Math.ceil(Math.max(1, viewportWidth) * Math.max(1, zoom)));
}

export function timelineTickInterval(rangeMs: number, trackWidth: number): number {
	const targetInterval = (Math.max(rangeMs, 1) * MIN_TICK_SPACING_PX) / Math.max(trackWidth, 1);
	return (
		TICK_INTERVALS_MS.find((interval) => interval >= targetInterval) ??
		TICK_INTERVALS_MS[TICK_INTERVALS_MS.length - 1]
	);
}

export function buildTimelineTicks(from: Date, to: Date, trackWidth: number): TimelineTick[] {
	const fromMs = from.getTime();
	const toMs = to.getTime();
	const rangeMs = Math.max(1, toMs - fromMs);
	const interval = timelineTickInterval(rangeMs, trackWidth);
	const firstTick = Math.ceil(fromMs / interval) * interval;
	const ticks: TimelineTick[] = [];

	for (let at = firstTick; at <= toMs; at += interval) {
		const date = new Date(at);
		ticks.push({
			at: date,
			fraction: fractionAtTime(from, to, date),
			label: formatTimelineLabel(date, rangeMs),
		});
	}

	return ticks;
}

export function coverageBarRects(
	coverage: readonly CoverageBar[],
	from: Date,
	to: Date,
): TimelineBarRect[] {
	const fromMs = from.getTime();
	const toMs = to.getTime();
	return coverage.flatMap((bar) => {
		const startMs = Date.parse(bar.start);
		const endMs = Date.parse(bar.end);
		if (
			Number.isNaN(startMs) ||
			Number.isNaN(endMs) ||
			endMs <= startMs ||
			endMs <= fromMs ||
			startMs >= toMs
		) {
			return [];
		}
		const left = fractionAtTime(from, to, new Date(Math.max(startMs, fromMs)));
		const right = fractionAtTime(from, to, new Date(Math.min(endMs, toMs)));
		return [{ start: bar.start, end: bar.end, left, width: Math.max(right - left, 0.001) }];
	});
}

export function mergeCoverageBars(coverage: readonly CoverageBar[]): CoverageBar[] {
	const ranges = coverage
		.flatMap((bar) => {
			const startMs = Date.parse(bar.start);
			const endMs = Date.parse(bar.end);
			return Number.isNaN(startMs) || Number.isNaN(endMs) || endMs <= startMs
				? []
				: [{ startMs, endMs }];
		})
		.sort((left, right) => left.startMs - right.startMs);
	const merged: Array<{ startMs: number; endMs: number }> = [];

	for (const range of ranges) {
		const previous = merged[merged.length - 1];
		if (previous && range.startMs <= previous.endMs) {
			previous.endMs = Math.max(previous.endMs, range.endMs);
			continue;
		}
		merged.push({ ...range });
	}

	return merged.map((range) => ({
		start: new Date(range.startMs).toISOString(),
		end: new Date(range.endMs).toISOString(),
	}));
}
