import type { Camera, Event } from "@nvr/api-client";
import { Link } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { StatusDot } from "@/components/status-dot";
import { LivePlayer } from "@/features/cameras/components/live-player";
import { eventEyebrow, formatEventDateTime, severityLabel } from "@/features/events/format";
import { colors, swatches } from "@/theme/colors";

type EventDetailProps = {
	event?: Event;
	camera?: Camera;
	playbackUrl?: string;
	playbackStartOffsetSec: number;
	eventPending: boolean;
	eventError: Error | null;
	playbackPending: boolean;
	playbackError: Error | null;
	acknowledgeError: Error | null;
	acknowledging: boolean;
	onRetryEvent: () => void;
	onRetryPlayback: () => void;
	onPlaybackEnded: () => void;
	onAcknowledge: () => void;
};

export function EventDetail({
	event,
	camera,
	playbackUrl,
	playbackStartOffsetSec,
	eventPending,
	eventError,
	playbackPending,
	playbackError,
	acknowledgeError,
	acknowledging,
	onRetryEvent,
	onRetryPlayback,
	onPlaybackEnded,
	onAcknowledge,
}: EventDetailProps) {
	if (eventPending) {
		return (
			<ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={styles.content}>
				<StatePanel detail="Loading event details…" loading title="Checking activity" />
			</ScrollView>
		);
	}
	if (eventError || !event) {
		return (
			<ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={styles.content}>
				<StatePanel
					actionLabel="Try again"
					detail={eventError?.message ?? "This event could not be found."}
					onAction={onRetryEvent}
					title="Event unavailable"
				/>
			</ScrollView>
		);
	}

	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			style={styles.screen}
		>
			<View style={styles.header}>
				<Text selectable style={styles.eyebrow}>
					{eventEyebrow(event)} · {severityLabel(event.severity)}
				</Text>
				<Text selectable style={styles.title}>
					{event.title}
				</Text>
				<Text selectable style={styles.time}>
					{formatEventDateTime(event.startedAt)}
				</Text>
			</View>

			<View style={styles.card}>
				<Text selectable style={styles.message}>
					{event.message}
				</Text>
				{event.endedAt ? (
					<Text selectable style={styles.detail}>
						Ended {formatEventDateTime(event.endedAt)}
					</Text>
				) : null}
			</View>

			{event.cameraId ? (
				<View style={styles.section}>
					<View style={styles.sectionHeader}>
						<Text selectable style={styles.sectionTitle}>
							Recorded video
						</Text>
						{camera ? (
							<StatusDot
								label={camera.name}
								tone={camera.status === "unknown" ? "neutral" : camera.status}
							/>
						) : (
							<Text style={styles.detail}>Camera</Text>
						)}
					</View>
					<View style={styles.player}>
						{playbackError ? (
							<StatePanel
								actionLabel="Retry"
								detail={playbackError.message}
								onAction={onRetryPlayback}
								title="Playback unavailable"
							/>
						) : playbackUrl ? (
							<LivePlayer
								loop={false}
								nativeControls
								onPlaybackEnded={onPlaybackEnded}
								onRetry={onRetryPlayback}
								startOffsetSec={playbackStartOffsetSec}
								uri={playbackUrl}
							/>
						) : (
							<StatePanel
								detail="Starting at the event time…"
								loading={playbackPending}
								title="Playback"
							/>
						)}
					</View>
					<Link href={{ pathname: "/camera/[id]", params: { id: event.cameraId } }} asChild>
						<Pressable
							accessibilityRole="button"
							style={({ pressed }) => [styles.secondaryAction, pressed ? styles.pressed : null]}
						>
							<Text style={styles.secondaryActionLabel}>Open live camera</Text>
						</Pressable>
					</Link>
				</View>
			) : (
				<StatePanel
					detail="This recorder event is not associated with a camera."
					title="System event"
				/>
			)}

			{event.acknowledged ? (
				<View style={styles.reviewed}>
					<Text style={styles.reviewedLabel}>Reviewed</Text>
				</View>
			) : (
				<Pressable
					accessibilityRole="button"
					disabled={acknowledging}
					onPress={onAcknowledge}
					style={({ pressed }) => [styles.primaryAction, pressed ? styles.pressed : null]}
				>
					<Text style={styles.primaryActionLabel}>
						{acknowledging ? "Marking as reviewed…" : "Mark as reviewed"}
					</Text>
				</Pressable>
			)}
			{acknowledgeError ? (
				<Text selectable style={styles.error}>
					{acknowledgeError.message}
				</Text>
			) : null}
		</ScrollView>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		gap: 22,
		padding: 16,
		paddingBottom: 40,
	},
	header: {
		gap: 6,
	},
	eyebrow: {
		color: colors.accent,
		fontSize: 11,
		fontWeight: "800",
		letterSpacing: 1,
	},
	title: {
		color: colors.label,
		fontSize: 27,
		fontWeight: "800",
		letterSpacing: -0.6,
	},
	time: {
		color: colors.secondaryLabel,
		fontSize: 13,
		fontVariant: ["tabular-nums"],
	},
	card: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 20,
		gap: 10,
		padding: 18,
	},
	message: {
		color: colors.label,
		fontSize: 16,
		lineHeight: 23,
	},
	detail: {
		color: colors.secondaryLabel,
		fontSize: 13,
	},
	section: {
		gap: 12,
	},
	sectionHeader: {
		alignItems: "center",
		flexDirection: "row",
		gap: 12,
		justifyContent: "space-between",
	},
	sectionTitle: {
		color: colors.label,
		fontSize: 18,
		fontWeight: "700",
	},
	player: {
		aspectRatio: 16 / 9,
		backgroundColor: swatches.preview,
		borderCurve: "continuous",
		borderRadius: 20,
		overflow: "hidden",
	},
	secondaryAction: {
		alignItems: "center",
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 13,
		minHeight: 44,
		justifyContent: "center",
		paddingHorizontal: 14,
	},
	secondaryActionLabel: {
		color: colors.accent,
		fontSize: 14,
		fontWeight: "700",
	},
	primaryAction: {
		alignItems: "center",
		backgroundColor: colors.label,
		borderCurve: "continuous",
		borderRadius: 14,
		minHeight: 48,
		justifyContent: "center",
		paddingHorizontal: 16,
	},
	primaryActionLabel: {
		color: colors.background,
		fontSize: 15,
		fontWeight: "700",
	},
	pressed: {
		opacity: 0.65,
	},
	reviewed: {
		alignItems: "center",
		backgroundColor: swatches.greenSoft,
		borderCurve: "continuous",
		borderRadius: 14,
		padding: 14,
	},
	reviewedLabel: {
		color: colors.green,
		fontSize: 14,
		fontWeight: "700",
	},
	error: {
		color: colors.red,
		fontSize: 13,
		lineHeight: 18,
		textAlign: "center",
	},
});
