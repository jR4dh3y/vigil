export type {
	Event,
	EventList,
	EventSeverity,
} from "@nvr/api-client";

export type ListEventsParams = {
	limit?: number;
	before?: string;
	cameraId?: string;
	type?: string;
	unacknowledgedOnly?: boolean;
};
