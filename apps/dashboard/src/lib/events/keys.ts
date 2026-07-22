import type { ListEventsParams } from "./types";

export const eventKeys = {
	all: ["events"] as const,
	lists: () => [...eventKeys.all, "list"] as const,
	list: (params: ListEventsParams = {}) => [...eventKeys.lists(), params] as const,
};
