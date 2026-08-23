const monthFormatter = new Intl.DateTimeFormat(undefined, {
	month: "long",
	year: "numeric",
});

const dateFormatter = new Intl.DateTimeFormat(undefined, {
	weekday: "long",
	month: "long",
	day: "numeric",
	year: "numeric",
});

const timeFormatter = new Intl.DateTimeFormat(undefined, {
	hour: "numeric",
	minute: "2-digit",
	second: "2-digit",
});

export function formatCalendarMonth(year: number, month: number): string {
	return monthFormatter.format(new Date(year, month, 1));
}

export function formatCalendarDate(value: string): string {
	const [year, month, day] = value.split("-").map(Number);
	const date = new Date(year ?? 0, (month ?? 1) - 1, day ?? 1);
	return dateFormatter.format(date);
}

export function formatRecordingTime(value: Date): string {
	return timeFormatter.format(value);
}
