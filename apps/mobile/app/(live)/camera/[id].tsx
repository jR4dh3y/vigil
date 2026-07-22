import { useQuery } from "@tanstack/react-query";
import { Stack, useLocalSearchParams } from "expo-router";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { StatusDot } from "@/components/status-dot";
import { cameraKeys, getCamera, getLiveStream } from "@/features/cameras/api";
import { LivePlayer } from "@/features/cameras/components/live-player";
import { resolveMediaUrl } from "@/lib/api/config";
import { colors } from "@/theme/colors";

export default function CameraScreen() {
	const { id } = useLocalSearchParams<{ id: string }>();
	const cameraQuery = useQuery({
		queryKey: cameraKeys.detail(id),
		queryFn: () => getCamera(id),
		enabled: Boolean(id),
	});
	const streamQuery = useQuery({
		queryKey: cameraKeys.live(id),
		queryFn: () => getLiveStream(id),
		enabled: Boolean(id) && cameraQuery.data?.status === "online",
		staleTime: 45_000,
	});
	const camera = cameraQuery.data;
	const streamUrl = streamQuery.data
		? resolveMediaUrl(streamQuery.data.hlsUrl, streamQuery.data.token)
		: undefined;

	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			style={styles.screen}
		>
			<Stack.Title>{camera?.name ?? "Camera"}</Stack.Title>

			{cameraQuery.isPending ? (
				<StatePanel detail="Loading camera details…" loading title="Connecting" />
			) : cameraQuery.isError ? (
				<StatePanel
					actionLabel="Try again"
					detail={cameraQuery.error.message}
					onAction={() => cameraQuery.refetch()}
					title="Camera unavailable"
				/>
			) : camera ? (
				<>
					{streamUrl ? (
						<LivePlayer uri={streamUrl} />
					) : camera.status !== "online" ? (
						<StatePanel
							detail="The recorder cannot reach this camera right now."
							title="Camera is offline"
						/>
					) : streamQuery.isError ? (
						<StatePanel
							actionLabel="Retry stream"
							detail={streamQuery.error.message}
							onAction={() => streamQuery.refetch()}
							title="Stream unavailable"
						/>
					) : (
						<StatePanel detail="Preparing the secure live stream…" loading title="Going live" />
					)}

					<View style={styles.detailsCard}>
						<View style={styles.detailsHeader}>
							<Text selectable style={styles.detailsTitle}>
								Camera details
							</Text>
							<StatusDot
								label={camera.status.charAt(0).toUpperCase() + camera.status.slice(1)}
								tone={
									camera.status === "online"
										? "online"
										: camera.status === "offline"
											? "offline"
											: "neutral"
								}
							/>
						</View>
						<View style={styles.detailRow}>
							<Text selectable style={styles.detailLabel}>
								Address
							</Text>
							<Text selectable numberOfLines={1} style={styles.detailValue}>
								{camera.host}
							</Text>
						</View>
						<View style={styles.detailRow}>
							<Text selectable style={styles.detailLabel}>
								Streams
							</Text>
							<Text selectable style={styles.detailValue}>
								{camera.streamProfiles.length}
							</Text>
						</View>
					</View>
				</>
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
		gap: 18,
		padding: 16,
		paddingBottom: 40,
	},
	detailsCard: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 20,
		gap: 14,
		padding: 18,
	},
	detailsHeader: {
		alignItems: "center",
		flexDirection: "row",
		justifyContent: "space-between",
	},
	detailsTitle: {
		color: colors.label,
		fontSize: 17,
		fontWeight: "700",
	},
	detailRow: {
		alignItems: "center",
		borderTopColor: colors.separator,
		borderTopWidth: StyleSheet.hairlineWidth,
		flexDirection: "row",
		gap: 16,
		justifyContent: "space-between",
		paddingTop: 14,
	},
	detailLabel: {
		color: colors.secondaryLabel,
		fontSize: 14,
	},
	detailValue: {
		color: colors.label,
		flex: 1,
		fontSize: 14,
		fontWeight: "600",
		textAlign: "right",
	},
});
