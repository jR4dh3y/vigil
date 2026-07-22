import type { Event, EventSeverity } from "@nvr/api-client";

const timeFormatter = new Intl.DateTimeFormat(undefined, {
	hour: "numeric",
	minute: "2-digit",
});

const dayFormatter = new Intl.DateTimeFormat(undefined, {
	month: "short",
	day: "numeric",
});

export function formatEventTime(value: string): string {
	const date = new Date(value);
	const today = new Date();
	const isToday = date.toDateString() === today.toDateString();
	return isToday
		? timeFormatter.format(date)
		: `${dayFormatter.format(date)}, ${timeFormatter.format(date)}`;
}

export function eventEyebrow(event: Event): string {
	return event.type.replace(/[._-]+/g, " ").toUpperCase();
}

export function severityLabel(severity: EventSeverity): string {
	if (severity === "critical") return "Critical";
	if (severity === "warning") return "Attention";
	return "Activity";
}
