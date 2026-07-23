import type { Camera } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { cameraKeys, getLiveStream } from "@/features/cameras/api";
import { LivePlayer } from "@/features/cameras/components/live-player";
import { resolveMediaUrl } from "@/lib/api/config";
import { colors, swatches } from "@/theme/colors";

type CameraCardProps = {
	camera: Camera;
	onPress: () => void;
};

export function CameraCard({ camera, onPress }: CameraCardProps) {
	const streamQuery = useQuery({
		queryKey: cameraKeys.live(camera.id),
		queryFn: () => getLiveStream(camera.id),
		enabled: camera.status === "online" && camera.enabled,
		staleTime: 45_000,
		refetchInterval: 50_000,
	});

	const streamUrl = streamQuery.data
		? resolveMediaUrl(streamQuery.data.hlsUrl, streamQuery.data.token)
		: undefined;

	return (
		<Pressable
			accessibilityHint="Opens the live camera"
			accessibilityLabel={camera.name}
			accessibilityRole="button"
			onPress={onPress}
			style={({ pressed }) => [styles.container, pressed && styles.pressed]}
		>
			<View style={styles.preview}>
				{streamUrl ? (
					<LivePlayer uri={streamUrl} />
				) : camera.status === "online" && streamQuery.isPending ? (
					<View style={styles.loading}>
						<ActivityIndicator color={swatches.white} />
					</View>
				) : (
					<View style={styles.empty} />
				)}

				<View style={styles.nameBar} pointerEvents="none">
					<Text numberOfLines={1} style={styles.name}>
						{camera.name}
					</Text>
				</View>
			</View>
		</Pressable>
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
		backgroundColor: swatches.preview,
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
