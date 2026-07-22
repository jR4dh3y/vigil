export const liveKeys = {
	all: ["live"] as const,
	streams: () => [...liveKeys.all, "stream"] as const,
	stream: (cameraId: string) => [...liveKeys.streams(), cameraId] as const,
};
