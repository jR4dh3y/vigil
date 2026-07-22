import type { DiskInfo } from "./types";

const UNITS = ["B", "KB", "MB", "GB", "TB", "PB"] as const;

export function formatBytes(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes < 0) {
		return "—";
	}
	if (bytes === 0) {
		return "0 B";
	}
	const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), UNITS.length - 1);
	const value = bytes / 1024 ** exp;
	const decimals = exp === 0 ? 0 : value >= 100 ? 0 : value >= 10 ? 1 : 2;
	return `${value.toFixed(decimals)} ${UNITS[exp]}`;
}

export function diskUsageLabel(disk: DiskInfo): string {
	return `${formatBytes(disk.usedBytes)} / ${formatBytes(disk.totalBytes)}`;
}

export function diskBarClass(usedPercent: number): string {
	if (usedPercent >= 90) {
		return "bg-red-500";
	}
	if (usedPercent >= 75) {
		return "bg-amber-500";
	}
	return "bg-emerald-500";
}

export function healthBadgeClass(status: "ok" | "degraded"): string {
	return status === "ok"
		? "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
		: "border-amber-500/30 bg-amber-500/10 text-amber-300";
}
