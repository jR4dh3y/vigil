import type { EventSeverity } from "./types";

export function severityLabel(severity: EventSeverity): string {
	switch (severity) {
		case "critical":
			return "Critical";
		case "warning":
			return "Warning";
		default:
			return "Info";
	}
}

export function severityBadgeClass(severity: EventSeverity): string {
	switch (severity) {
		case "critical":
			return "border-red-500/30 bg-red-500/10 text-red-300";
		case "warning":
			return "border-amber-500/30 bg-amber-500/10 text-amber-300";
		default:
			return "border-sky-500/30 bg-sky-500/10 text-sky-300";
	}
}

export function formatEventTime(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return iso;
	}
	return date.toLocaleString(undefined, {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
	});
}
