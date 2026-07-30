export const storageKeys = {
	all: ["storage"] as const,
	gdriveStatus: () => [...storageKeys.all, "gdrive", "status"] as const,
};
