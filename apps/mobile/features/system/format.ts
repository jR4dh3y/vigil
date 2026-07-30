const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB", "PB"] as const;

export function formatBytes(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes < 0) {
		return "—";
	}
	if (bytes === 0) {
		return "0 B";
	}
	const exponent = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), BYTE_UNITS.length - 1);
	const value = bytes / 1024 ** exponent;
	const decimals = exponent === 0 ? 0 : value >= 10 ? 1 : 2;
	return `${value.toFixed(decimals)} ${BYTE_UNITS[exponent]}`;
}

export function formatDiskUsage(usedBytes: number, totalBytes: number): string {
	return `${formatBytes(usedBytes)} of ${formatBytes(totalBytes)}`;
}
