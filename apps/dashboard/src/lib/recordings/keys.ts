export const recordingKeys = {
	all: ["recordings"] as const,
	lists: () => [...recordingKeys.all, "list"] as const,
	list: (cameraId: string, from: string, to: string) =>
		[...recordingKeys.lists(), cameraId, from, to] as const,
	playback: () => [...recordingKeys.all, "playback"] as const,
	playbackSession: (cameraId: string, start: string) =>
		[...recordingKeys.playback(), cameraId, start] as const,
};
