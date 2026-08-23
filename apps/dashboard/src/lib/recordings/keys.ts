export const recordingKeys = {
	all: ["recordings"] as const,
	lists: () => [...recordingKeys.all, "list"] as const,
	list: (cameraId: string, from: string, to: string) =>
		[...recordingKeys.lists(), cameraId, from, to] as const,
	listMany: (cameraIds: readonly string[], from: string, to: string) =>
		[...recordingKeys.lists(), "multi", cameraIds.join(","), from, to] as const,
	days: (cameraIds: readonly string[], from: string, to: string, timeZone: string) =>
		[...recordingKeys.all, "days", cameraIds.join(","), from, to, timeZone] as const,
	playback: () => [...recordingKeys.all, "playback"] as const,
	playbackSession: (cameraId: string, start: string) =>
		[...recordingKeys.playback(), cameraId, start] as const,
};
