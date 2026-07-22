export const cameraKeys = {
	all: ["cameras"] as const,
	lists: () => [...cameraKeys.all, "list"] as const,
	list: () => [...cameraKeys.lists()] as const,
	details: () => [...cameraKeys.all, "detail"] as const,
	detail: (id: string) => [...cameraKeys.details(), id] as const,
	snapshots: () => [...cameraKeys.all, "snapshot"] as const,
	snapshot: (id: string) => [...cameraKeys.snapshots(), id] as const,
};
