import type { CoverageBar } from "@nvr/api-client";
import { useMemo, useState } from "react";
import {
	type AccessibilityActionEvent,
	type GestureResponderEvent,
	Pressable,
	StyleSheet,
	Text,
	View,
} from "react-native";
import {
	fractionAtTime,
	mergeCoverageBars,
	nearestCoverageTime,
	type TimeRange,
	timeAtFraction,
} from "@/features/recordings/date";
import { formatRecordingTime } from "@/features/recordings/format";
import { colors, swatches } from "@/theme/colors";

type RecordingTimelineProps = {
	range: TimeRange;
	coverage: readonly CoverageBar[];
	selectedTime: Date | null;
	disabled?: boolean;
	onSeek: (time: Date) => void;
};

const ACCESSIBILITY_STEP_MS = 5 * 60 * 1000;

export function RecordingTimeline({
	range,
	coverage,
	selectedTime,
	disabled = false,
	onSeek,
}: RecordingTimelineProps) {
	const [width, setWidth] = useState(0);
	const mergedCoverage = useMemo(() => mergeCoverageBars(coverage), [coverage]);
	const selectedFraction = selectedTime ? fractionAtTime(range, selectedTime) : null;

	const seek = (candidate: Date) => {
		const playable = nearestCoverageTime(mergedCoverage, candidate);
		if (playable) {
			onSeek(playable);
		}
	};
	const handlePress = (event: GestureResponderEvent) => {
		if (disabled || width <= 0) {
			return;
		}
		seek(timeAtFraction(range, event.nativeEvent.locationX / width));
	};
	const handleAccessibilityAction = (event: AccessibilityActionEvent) => {
		if (disabled || !selectedTime) {
			return;
		}
		const direction = event.nativeEvent.actionName === "increment" ? 1 : -1;
		seek(new Date(selectedTime.getTime() + direction * ACCESSIBILITY_STEP_MS));
	};

	return (
		<View style={styles.container}>
			<View style={styles.header}>
				<Text selectable style={styles.title}>
					Daily coverage
				</Text>
				<Text selectable style={styles.selectedLabel}>
					{selectedTime ? formatRecordingTime(selectedTime) : "No recording selected"}
				</Text>
			</View>
			<Pressable
				accessibilityActions={[{ name: "increment" }, { name: "decrement" }]}
				accessibilityHint="Tap to seek. Swipe up or down to move five minutes."
				accessibilityLabel="Recording timeline"
				accessibilityRole="adjustable"
				accessibilityState={{ disabled }}
				accessibilityValue={{ text: selectedTime ? formatRecordingTime(selectedTime) : "Empty" }}
				disabled={disabled}
				onAccessibilityAction={handleAccessibilityAction}
				onLayout={(event) => setWidth(event.nativeEvent.layout.width)}
				onPress={handlePress}
				style={styles.track}
			>
				{mergedCoverage.map((bar) => {
					const start = new Date(bar.start);
					const end = new Date(bar.end);
					const left = fractionAtTime(range, start) * 100;
					const right = fractionAtTime(range, end) * 100;
					return (
						<View
							key={`${bar.start}-${bar.end}`}
							pointerEvents="none"
							style={[
								styles.coverage,
								{ left: `${left}%`, width: `${Math.max(right - left, 0.2)}%` },
							]}
						/>
					);
				})}
				{selectedFraction === null ? null : (
					<View
						pointerEvents="none"
						style={[styles.playhead, { left: `${selectedFraction * 100}%` }]}
					/>
				)}
			</Pressable>
			<View style={styles.axis}>
				<Text style={styles.axisLabel}>12 AM</Text>
				<Text style={styles.axisLabel}>6 AM</Text>
				<Text style={styles.axisLabel}>12 PM</Text>
				<Text style={styles.axisLabel}>6 PM</Text>
				<Text style={styles.axisLabel}>12 AM</Text>
			</View>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 20,
		gap: 10,
		padding: 16,
	},
	header: {
		alignItems: "baseline",
		flexDirection: "row",
		gap: 12,
		justifyContent: "space-between",
	},
	title: {
		color: colors.label,
		fontSize: 15,
		fontWeight: "700",
	},
	selectedLabel: {
		color: colors.secondaryLabel,
		fontSize: 12,
		fontVariant: ["tabular-nums"],
	},
	track: {
		backgroundColor: swatches.previewRaised,
		borderCurve: "continuous",
		borderRadius: 9,
		height: 42,
		overflow: "hidden",
		position: "relative",
	},
	coverage: {
		backgroundColor: colors.accent,
		bottom: 8,
		borderCurve: "continuous",
		borderRadius: 4,
		position: "absolute",
		top: 8,
	},
	playhead: {
		backgroundColor: swatches.white,
		bottom: 2,
		position: "absolute",
		top: 2,
		width: 2,
	},
	axis: {
		flexDirection: "row",
		justifyContent: "space-between",
	},
	axisLabel: {
		color: colors.secondaryLabel,
		fontSize: 9,
		fontVariant: ["tabular-nums"],
	},
});
