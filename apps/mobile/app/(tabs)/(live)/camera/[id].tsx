import { useQuery } from "@tanstack/react-query";
import { Stack, useLocalSearchParams } from "expo-router";
import { StyleSheet, Text, View } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { cameraKeys, getCamera, getLiveStream } from "@/features/cameras/api";
import { LivePlayer } from "@/features/cameras/components/live-player";
import { resolveMediaUrl } from "@/lib/api/config";
import { colors, swatches } from "@/theme/colors";

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
		refetchInterval: 50_000,
	});
	const camera = cameraQuery.data;
	const streamUrl = streamQuery.data
		? resolveMediaUrl(streamQuery.data.hlsUrl, streamQuery.data.token)
		: undefined;

	return (
		<View style={styles.screen}>
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
				<View style={styles.stage}>
					{streamUrl ? (
						<LivePlayer nativeControls uri={streamUrl} />
					) : camera.status !== "online" ? (
						<View style={styles.offline}>
							<Text style={styles.offlineText}>Offline</Text>
						</View>
					) : streamQuery.isError ? (
						<StatePanel
							actionLabel="Retry"
							detail={streamQuery.error.message}
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
			) : null}
		</View>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
		padding: 16,
	},
	stage: {
		aspectRatio: 16 / 9,
		backgroundColor: swatches.preview,
		borderCurve: "continuous",
		borderRadius: 22,
		overflow: "hidden",
		width: "100%",
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
