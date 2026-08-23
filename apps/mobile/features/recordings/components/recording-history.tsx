import type { ReactNode } from "react";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { LivePlayer } from "@/features/cameras/components/live-player";
import { RecordingCalendar } from "@/features/recordings/components/recording-calendar";
import { RecordingTimeline } from "@/features/recordings/components/recording-timeline";
import { formatCalendarDate } from "@/features/recordings/format";
import type { useRecordingHistory } from "@/features/recordings/use-recording-history";
import { colors, swatches } from "@/theme/colors";

type RecordingHistoryProps = {
	history: ReturnType<typeof useRecordingHistory>;
};

export function RecordingHistory({ history }: RecordingHistoryProps) {
	if (history.cameraPending) {
		return (
			<HistoryScroll>
				<StatePanel detail="Loading camera…" loading title="Opening history" />
			</HistoryScroll>
		);
	}
	if (history.cameraError || !history.camera) {
		return (
			<HistoryScroll>
				<StatePanel
					actionLabel="Try again"
					detail={history.cameraError?.message ?? "This camera could not be found."}
					onAction={() => history.retryCamera()}
					title="Camera unavailable"
				/>
			</HistoryScroll>
		);
	}

	const playbackError = history.playbackUrlError ?? history.playbackError;
	const hasCoverage = history.coverage.length > 0;

	return (
		<HistoryScroll>
			<View style={styles.intro}>
				<Text selectable style={styles.title}>
					{formatCalendarDate(history.selectedDay)}
				</Text>
				<Text selectable style={styles.detail}>
					Tap any marked date, then seek anywhere inside the day’s recorded coverage.
				</Text>
			</View>

			<RecordingCalendar
				availability={history.availability}
				error={history.daysError}
				loading={history.daysLoading}
				maxDate={history.maxDate}
				month={history.month}
				onChange={history.selectDay}
				onMonthChange={history.setMonth}
				value={history.selectedDay}
			/>

			{history.recordingsPending ? (
				<StatePanel detail="Loading daily coverage…" loading title="Checking recordings" />
			) : history.recordingsError ? (
				<StatePanel
					actionLabel="Try again"
					detail={history.recordingsError.message}
					onAction={() => history.retryRecordings()}
					title="Recordings unavailable"
				/>
			) : !history.dayRange || !hasCoverage ? (
				<StatePanel
					detail="Choose another date. Days with retained video have a colored marker."
					title="No video on this day"
				/>
			) : (
				<>
					<RecordingTimeline
						coverage={history.coverage}
						disabled={history.playbackPending}
						onSeek={history.seek}
						range={history.dayRange}
						selectedTime={history.selectedTime}
					/>

					<View style={styles.playerSection}>
						<View style={styles.playerHeader}>
							<Text selectable style={styles.playerTitle}>
								Playback
							</Text>
							{history.playbackSession ? (
								<View style={styles.sourceBadge}>
									<Text style={styles.sourceLabel}>
										{history.playbackSession.source === "gdrive" ? "Google Drive" : "Local"}
									</Text>
								</View>
							) : null}
						</View>
						<View style={styles.player}>
							{playbackError ? (
								<StatePanel
									actionLabel="Retry"
									detail={playbackError.message}
									onAction={() => history.retryPlayback()}
									title="Playback unavailable"
								/>
							) : history.playbackUrl && history.playbackSession ? (
								<LivePlayer
									loop={false}
									nativeControls
									onPlaybackEnded={history.continuePlayback}
									onRetry={() => history.retryPlayback()}
									startOffsetSec={history.playbackSession.startOffsetSec}
									uri={history.playbackUrl}
								/>
							) : (
								<StatePanel
									detail="Preparing video at the selected time…"
									loading={history.playbackPending}
									title="Starting playback"
								/>
							)}
						</View>
					</View>
				</>
			)}
		</HistoryScroll>
	);
}

function HistoryScroll({ children }: { children: ReactNode }) {
	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			style={styles.screen}
		>
			{children}
		</ScrollView>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		gap: 18,
		padding: 16,
		paddingBottom: 40,
	},
	intro: {
		gap: 5,
	},
	title: {
		color: colors.label,
		fontSize: 24,
		fontWeight: "800",
		letterSpacing: -0.5,
	},
	detail: {
		color: colors.secondaryLabel,
		fontSize: 13,
		lineHeight: 19,
	},
	playerSection: {
		gap: 10,
	},
	playerHeader: {
		alignItems: "center",
		flexDirection: "row",
		justifyContent: "space-between",
	},
	playerTitle: {
		color: colors.label,
		fontSize: 18,
		fontWeight: "700",
	},
	sourceBadge: {
		backgroundColor: swatches.neutralSoft,
		borderCurve: "continuous",
		borderRadius: 99,
		paddingHorizontal: 9,
		paddingVertical: 5,
	},
	sourceLabel: {
		color: colors.secondaryLabel,
		fontSize: 10,
		fontWeight: "700",
		textTransform: "uppercase",
	},
	player: {
		aspectRatio: 16 / 9,
		backgroundColor: swatches.preview,
		borderCurve: "continuous",
		borderRadius: 20,
		overflow: "hidden",
	},
});
