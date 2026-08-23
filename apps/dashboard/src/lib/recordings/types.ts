export type {
	CoverageBar,
	PlaybackRequest,
	PlaybackSession,
	RecordingDayAvailability,
	RecordingDayList,
	RecordingDaySource,
	RecordingList,
	RecordingSegment,
} from "@nvr/api-client";

/** Inclusive time window for recording list queries. */
export type TimeRange = {
	from: Date;
	to: Date;
};

export type RangePreset = "1h" | "24h" | "7d";
