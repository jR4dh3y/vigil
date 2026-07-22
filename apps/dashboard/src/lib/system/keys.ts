export const systemKeys = {
	all: ["system"] as const,
	status: () => [...systemKeys.all, "status"] as const,
	disk: () => [...systemKeys.all, "disk"] as const,
	settings: () => [...systemKeys.all, "settings"] as const,
};
