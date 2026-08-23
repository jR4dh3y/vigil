import type { RecordingDaySource } from "@nvr/api-client";
import { type ColorValue, Pressable, StyleSheet, Text, View } from "react-native";
import {
	type CalendarMonth,
	calendarGridDays,
	shiftCalendarMonth,
} from "@/features/recordings/date";
import { formatCalendarDate, formatCalendarMonth } from "@/features/recordings/format";
import { colors, swatches } from "@/theme/colors";

type RecordingCalendarProps = {
	value: string;
	maxDate: string;
	month: CalendarMonth;
	availability: ReadonlyMap<string, RecordingDaySource>;
	loading: boolean;
	error: boolean;
	onChange: (value: string) => void;
	onMonthChange: (month: CalendarMonth) => void;
};

const weekdays = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"];

export function RecordingCalendar({
	value,
	maxDate,
	month,
	availability,
	loading,
	error,
	onChange,
	onMonthChange,
}: RecordingCalendarProps) {
	const days = calendarGridDays(month, maxDate);
	const maxMonth = new Date(`${maxDate}T12:00:00`);
	const canMoveForward =
		month.year * 12 + month.month < maxMonth.getFullYear() * 12 + maxMonth.getMonth();

	return (
		<View style={styles.container}>
			<View style={styles.monthHeader}>
				<CalendarNavigationButton
					accessibilityLabel="Previous month"
					label="‹"
					onPress={() => onMonthChange(shiftCalendarMonth(month, -1))}
				/>
				<Text selectable style={styles.monthLabel}>
					{formatCalendarMonth(month.year, month.month)}
				</Text>
				<CalendarNavigationButton
					accessibilityLabel="Next month"
					disabled={!canMoveForward}
					label="›"
					onPress={() => onMonthChange(shiftCalendarMonth(month, 1))}
				/>
			</View>

			<View style={styles.grid}>
				{weekdays.map((weekday) => (
					<View key={weekday} style={styles.cell}>
						<Text style={styles.weekday}>{weekday}</Text>
					</View>
				))}
				{days.map((day) => {
					const source = availability.get(day.value);
					const selected = day.value === value;
					return (
						<View key={day.value} style={styles.cell}>
							<Pressable
								accessibilityLabel={`${formatCalendarDate(day.value)}, ${sourceLabel(source)}`}
								accessibilityRole="button"
								accessibilityState={{ disabled: day.isFuture, selected }}
								disabled={day.isFuture}
								onPress={() => onChange(day.value)}
								style={({ pressed }) => [
									styles.day,
									!day.inMonth && styles.outsideMonth,
									selected && styles.selectedDay,
									day.isFuture && styles.futureDay,
									pressed && styles.pressedDay,
								]}
							>
								<Text style={[styles.dayLabel, selected && styles.selectedDayLabel]}>
									{day.day}
								</Text>
								{source ? <AvailabilityDots source={source} /> : null}
							</Pressable>
						</View>
					);
				})}
			</View>

			<View style={styles.legend}>
				<LegendItem color={colors.accent} label="Local" />
				<LegendItem color={colors.green} label="Google Drive" />
				<Text selectable style={error ? styles.error : styles.status}>
					{loading ? "Loading days…" : error ? "Day markers unavailable" : ""}
				</Text>
			</View>
		</View>
	);
}

function CalendarNavigationButton({
	label,
	accessibilityLabel,
	onPress,
	disabled = false,
}: {
	label: string;
	accessibilityLabel: string;
	onPress: () => void;
	disabled?: boolean;
}) {
	return (
		<Pressable
			accessibilityLabel={accessibilityLabel}
			accessibilityRole="button"
			disabled={disabled}
			onPress={onPress}
			style={({ pressed }) => [
				styles.navigationButton,
				disabled && styles.navigationButtonDisabled,
				pressed && styles.navigationButtonPressed,
			]}
		>
			<Text style={styles.navigationLabel}>{label}</Text>
		</Pressable>
	);
}

function AvailabilityDots({ source }: { source: RecordingDaySource }) {
	return (
		<View pointerEvents="none" style={styles.dots}>
			{source === "local" || source === "mixed" ? (
				<View style={[styles.dot, { backgroundColor: colors.accent }]} />
			) : null}
			{source === "gdrive" || source === "mixed" ? (
				<View style={[styles.dot, { backgroundColor: colors.green }]} />
			) : null}
		</View>
	);
}

function LegendItem({ color, label }: { color: ColorValue; label: string }) {
	return (
		<View style={styles.legendItem}>
			<View style={[styles.legendDot, { backgroundColor: color }]} />
			<Text style={styles.legendLabel}>{label}</Text>
		</View>
	);
}

function sourceLabel(source: RecordingDaySource | undefined): string {
	if (source === "local") return "local recording available";
	if (source === "gdrive") return "Google Drive recording available";
	if (source === "mixed") return "local and Google Drive recordings available";
	return "no recordings";
}

const styles = StyleSheet.create({
	container: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 22,
		gap: 10,
		padding: 12,
	},
	monthHeader: {
		alignItems: "center",
		flexDirection: "row",
		justifyContent: "space-between",
	},
	monthLabel: {
		color: colors.label,
		fontSize: 15,
		fontWeight: "700",
	},
	navigationButton: {
		alignItems: "center",
		borderCurve: "continuous",
		borderRadius: 11,
		height: 36,
		justifyContent: "center",
		width: 36,
	},
	navigationButtonDisabled: {
		opacity: 0.25,
	},
	navigationButtonPressed: {
		backgroundColor: swatches.neutralSoft,
	},
	navigationLabel: {
		color: colors.label,
		fontSize: 28,
		fontWeight: "300",
		lineHeight: 30,
	},
	grid: {
		flexDirection: "row",
		flexWrap: "wrap",
	},
	cell: {
		flexBasis: "14.285714%",
		padding: 2,
	},
	weekday: {
		color: colors.secondaryLabel,
		fontSize: 10,
		fontWeight: "700",
		paddingVertical: 4,
		textAlign: "center",
	},
	day: {
		alignItems: "center",
		aspectRatio: 1,
		borderCurve: "continuous",
		borderRadius: 10,
		justifyContent: "center",
		position: "relative",
	},
	outsideMonth: {
		opacity: 0.4,
	},
	selectedDay: {
		backgroundColor: colors.label,
	},
	futureDay: {
		opacity: 0.18,
	},
	pressedDay: {
		opacity: 0.6,
	},
	dayLabel: {
		color: colors.label,
		fontSize: 12,
		fontVariant: ["tabular-nums"],
		fontWeight: "600",
	},
	selectedDayLabel: {
		color: colors.background,
	},
	dots: {
		bottom: 4,
		flexDirection: "row",
		gap: 2,
		position: "absolute",
	},
	dot: {
		borderRadius: 99,
		height: 4,
		width: 4,
	},
	legend: {
		alignItems: "center",
		borderTopColor: colors.separator,
		borderTopWidth: StyleSheet.hairlineWidth,
		flexDirection: "row",
		gap: 12,
		minHeight: 26,
		paddingHorizontal: 4,
		paddingTop: 8,
	},
	legendItem: {
		alignItems: "center",
		flexDirection: "row",
		gap: 5,
	},
	legendDot: {
		borderRadius: 99,
		height: 6,
		width: 6,
	},
	legendLabel: {
		color: colors.secondaryLabel,
		fontSize: 10,
	},
	status: {
		color: colors.secondaryLabel,
		flex: 1,
		fontSize: 10,
		textAlign: "right",
	},
	error: {
		color: colors.orange,
		flex: 1,
		fontSize: 10,
		textAlign: "right",
	},
});
