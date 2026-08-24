import type { Event, EventSeverity } from "@nvr/api-client";

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
	dateStyle: "medium",
	timeStyle: "short",
});

export function formatEventDateTime(value: string): string {
	return dateTimeFormatter.format(new Date(value));
}

export function eventEyebrow(event: Event): string {
	return event.type.replace(/[._-]+/g, " ").toUpperCase();
}

export function severityLabel(severity: EventSeverity): string {
	if (severity === "critical") return "Critical";
	if (severity === "warning") return "Attention";
	return "Activity";
}
