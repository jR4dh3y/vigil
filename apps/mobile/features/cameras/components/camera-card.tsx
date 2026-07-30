import type { Camera } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { Link } from "expo-router";
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { StatusDot } from "@/components/status-dot";
import { cameraKeys, getLiveStream, liveRefetchInterval } from "@/features/cameras/api";
import { LiveStreamPlayer } from "@/features/cameras/components/live-stream-player";
import { resolveMediaUrl } from "@/lib/api/config";
import { colors, swatches } from "@/theme/colors";

type CameraCardProps = {
	camera: Camera;
	active: boolean;
};

export function CameraCard({ camera, active }: CameraCardProps) {
	const streamQuery = useQuery({
		queryKey: cameraKeys.live(camera.id),
		queryFn: () => getLiveStream(camera.id),
		enabled: active && camera.status === "online" && camera.enabled,
		staleTime: 45_000,
		refetchInterval: (query) => liveRefetchInterval(query.state.data?.expiresAt),
	});

	const hlsUrl = streamQuery.data
		? resolveMediaUrl(streamQuery.data.hlsUrl, streamQuery.data.token)
		: undefined;
	const whepUrl = streamQuery.data
		? resolveMediaUrl(streamQuery.data.whepUrl, streamQuery.data.token)
		: undefined;

	return (
		<Link href={{ pathname: "/camera/[id]", params: { id: camera.id } }} asChild>
			<Pressable
				accessibilityHint="Opens the live camera"
				accessibilityLabel={camera.name}
				accessibilityRole="button"
				style={({ pressed }) => [styles.container, pressed ? styles.pressed : null]}
			>
				<View style={styles.preview}>
					{active && hlsUrl && whepUrl ? (
						<LiveStreamPlayer hlsUri={hlsUrl} whepUri={whepUrl} />
					) : active && camera.status === "online" && camera.enabled && streamQuery.isPending ? (
						<View style={styles.loading}>
							<ActivityIndicator color={swatches.white} />
						</View>
					) : (
						<View style={styles.empty}>
							<Text style={styles.emptyLabel}>
								{active
									? camera.enabled
										? "Preview unavailable"
										: "Camera disabled"
									: "Preview paused"}
							</Text>
						</View>
					)}

					<View style={styles.nameBar} pointerEvents="none">
						<Text numberOfLines={1} style={styles.name}>
							{camera.name}
						</Text>
					</View>
				</View>
				<View style={styles.footer}>
					<StatusDot
						label={camera.enabled ? camera.status : "disabled"}
						tone={camera.status === "unknown" || !camera.enabled ? "neutral" : camera.status}
					/>
					<Text numberOfLines={1} style={styles.host}>
						{camera.host}
					</Text>
				</View>
			</Pressable>
		</Link>
	);
}

const styles = StyleSheet.create({
	container: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 22,
		overflow: "hidden",
	},
	pressed: {
		opacity: 0.9,
		transform: [{ scale: 0.995 }],
	},
	preview: {
		aspectRatio: 16 / 9,
		backgroundColor: swatches.preview,
		overflow: "hidden",
	},
	loading: {
		...StyleSheet.absoluteFill,
		alignItems: "center",
		backgroundColor: swatches.preview,
		justifyContent: "center",
	},
	empty: {
		...StyleSheet.absoluteFill,
		alignItems: "center",
		backgroundColor: swatches.preview,
		justifyContent: "center",
	},
	emptyLabel: {
		color: "rgba(255,255,255,0.5)",
		fontSize: 13,
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
	footer: {
		alignItems: "center",
		flexDirection: "row",
		gap: 12,
		justifyContent: "space-between",
		minHeight: 44,
		paddingHorizontal: 13,
		paddingVertical: 10,
	},
	host: {
		color: colors.secondaryLabel,
		flexShrink: 1,
		fontSize: 12,
	},
});
