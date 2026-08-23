import { useQuery } from "@tanstack/react-query";
import { Stack, useLocalSearchParams } from "expo-router";
import { useIsFocused } from "expo-router/react-navigation";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { StatusDot } from "@/components/status-dot";
import { cameraKeys, getCamera, getLiveStream, liveRefetchInterval } from "@/features/cameras/api";
import { LiveStreamPlayer } from "@/features/cameras/components/live-stream-player";
import { resolveLiveStream } from "@/features/cameras/media";
import { colors, swatches } from "@/theme/colors";

export default function CameraScreen() {
	const { id } = useLocalSearchParams<{ id: string }>();
	const isFocused = useIsFocused();
	const cameraQuery = useQuery({
		queryKey: cameraKeys.detail(id),
		queryFn: () => getCamera(id),
		enabled: Boolean(id) && isFocused,
		refetchInterval: 30_000,
	});
	const streamQuery = useQuery({
		queryKey: cameraKeys.live(id),
		queryFn: () => getLiveStream(id),
		enabled:
			Boolean(id) &&
			isFocused &&
			cameraQuery.data?.status === "online" &&
			cameraQuery.data?.enabled,
		staleTime: 45_000,
		refetchInterval: (query) => liveRefetchInterval(query.state.data?.expiresAt),
	});
	const camera = cameraQuery.data;
	const media = resolveLiveStream(streamQuery.data);

	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			style={styles.screen}
		>
			<Stack.Title>{camera?.name ?? "Camera"}</Stack.Title>

			{cameraQuery.isPending ? (
				<StatePanel detail="Loading camera…" loading title="Connecting" />
			) : cameraQuery.isError ? (
				<StatePanel
					actionLabel="Try again"
					detail={cameraQuery.error.message}
					onAction={() => cameraQuery.refetch()}
					title="Camera unavailable"
				/>
			) : camera ? (
				<View style={styles.camera}>
					<View style={styles.stage}>
						{!camera.enabled ? (
							<View style={styles.offline}>
								<Text style={styles.offlineText}>Disabled</Text>
							</View>
						) : camera.status !== "online" ? (
							<View style={styles.offline}>
								<Text style={styles.offlineText}>Offline</Text>
							</View>
						) : isFocused && media.kind === "ready" ? (
							<LiveStreamPlayer hlsUri={media.hlsUrl} nativeControls whepUri={media.whepUrl} />
						) : streamQuery.isError || media.kind === "error" ? (
							<StatePanel
								actionLabel="Retry"
								detail={
									media.kind === "error"
										? media.error.message
										: (streamQuery.error?.message ?? "The stream is unavailable")
								}
								onAction={() => streamQuery.refetch()}
								title="Stream unavailable"
							/>
						) : (
							<StatePanel detail="Starting stream…" loading title="Live" />
						)}

						<View style={styles.nameBar} pointerEvents="none">
							<Text numberOfLines={1} style={styles.name}>
								{camera.name}
							</Text>
						</View>
					</View>

					<View style={styles.details}>
						<View style={styles.detailRow}>
							<Text style={styles.detailLabel}>Status</Text>
							<StatusDot
								label={camera.enabled ? camera.status : "disabled"}
								tone={camera.status === "unknown" || !camera.enabled ? "neutral" : camera.status}
							/>
						</View>
						<View style={styles.detailRow}>
							<Text style={styles.detailLabel}>Address</Text>
							<Text selectable numberOfLines={1} style={styles.detailValue}>
								{camera.host}
							</Text>
						</View>
						<View style={styles.detailRow}>
							<Text style={styles.detailLabel}>Streams</Text>
							<Text selectable style={styles.detailValue}>
								{camera.streamProfiles.length}
							</Text>
						</View>
					</View>
				</View>
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
	camera: {
		gap: 18,
	},
	stage: {
		aspectRatio: 16 / 9,
		backgroundColor: swatches.preview,
		borderCurve: "continuous",
		borderRadius: 22,
		overflow: "hidden",
		width: "100%",
	},
	details: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 18,
		gap: 1,
		paddingHorizontal: 16,
		width: "100%",
	},
	detailRow: {
		alignItems: "center",
		borderBottomColor: colors.separator,
		borderBottomWidth: StyleSheet.hairlineWidth,
		flexDirection: "row",
		gap: 16,
		justifyContent: "space-between",
		minHeight: 50,
	},
	detailLabel: {
		color: colors.label,
		fontSize: 15,
	},
	detailValue: {
		color: colors.secondaryLabel,
		flexShrink: 1,
		fontSize: 13,
	},
	offline: {
		...StyleSheet.absoluteFill,
		alignItems: "center",
		backgroundColor: swatches.preview,
		justifyContent: "center",
	},
	offlineText: {
		color: "rgba(255,255,255,0.55)",
		fontSize: 15,
		fontWeight: "600",
	},
	nameBar: {
		backgroundColor: "rgba(0, 0, 0, 0.45)",
		left: 0,
		paddingHorizontal: 12,
		paddingVertical: 8,
		position: "absolute",
		right: 0,
		top: 0,
	},
	name: {
		color: swatches.white,
		fontSize: 14,
		fontWeight: "700",
	},
});
