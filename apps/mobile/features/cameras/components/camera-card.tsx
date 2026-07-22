import type { Camera } from "@nvr/api-client";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { StatusDot } from "@/components/status-dot";
import { colors, swatches } from "@/theme/colors";

type CameraCardProps = {
	camera: Camera;
	onPress: () => void;
};

const statusTone = {
	online: "online",
	offline: "offline",
	unknown: "neutral",
} as const;

export function CameraCard({ camera, onPress }: CameraCardProps) {
	const statusLabel = camera.status.charAt(0).toUpperCase() + camera.status.slice(1);

	return (
		<Pressable
			accessibilityHint="Opens the live camera"
			accessibilityLabel={`${camera.name}, ${statusLabel}`}
			accessibilityRole="button"
			onPress={onPress}
			style={({ pressed }) => [styles.container, pressed && styles.pressed]}
		>
			<View style={styles.preview}>
				<View style={styles.previewGlow} />
				<View style={styles.frame}>
					<View style={styles.frameCornerTop} />
					<View style={styles.frameCornerBottom} />
				</View>
				<View style={styles.livePill}>
					<StatusDot
						inverted
						label={camera.status === "online" ? "LIVE" : statusLabel.toUpperCase()}
						tone={statusTone[camera.status]}
					/>
				</View>
				<Text style={styles.monogram}>{camera.name.slice(0, 2).toUpperCase()}</Text>
			</View>
			<View style={styles.footer}>
				<View style={styles.copy}>
					<Text numberOfLines={1} selectable style={styles.name}>
						{camera.name}
					</Text>
					<Text numberOfLines={1} selectable style={styles.host}>
						{camera.host}
					</Text>
				</View>
				<Text style={styles.chevron}>›</Text>
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
		opacity: 0.78,
		transform: [{ scale: 0.99 }],
	},
	preview: {
		alignItems: "center",
		aspectRatio: 16 / 9,
		backgroundColor: swatches.preview,
		justifyContent: "center",
		overflow: "hidden",
	},
	previewGlow: {
		backgroundColor: "rgba(108, 168, 132, 0.10)",
		borderRadius: 999,
		height: 210,
		position: "absolute",
		right: -64,
		top: -120,
		width: 210,
	},
	frame: {
		borderColor: "rgba(255, 255, 255, 0.06)",
		borderRadius: 18,
		borderWidth: 1,
		height: "58%",
		position: "absolute",
		transform: [{ rotate: "-5deg" }],
		width: "56%",
	},
	frameCornerTop: {
		borderLeftColor: "rgba(255, 255, 255, 0.17)",
		borderLeftWidth: 2,
		borderTopColor: "rgba(255, 255, 255, 0.17)",
		borderTopLeftRadius: 7,
		borderTopWidth: 2,
		height: 20,
		left: 8,
		position: "absolute",
		top: 8,
		width: 20,
	},
	frameCornerBottom: {
		borderBottomColor: "rgba(255, 255, 255, 0.17)",
		borderBottomRightRadius: 7,
		borderBottomWidth: 2,
		borderRightColor: "rgba(255, 255, 255, 0.17)",
		borderRightWidth: 2,
		bottom: 8,
		height: 20,
		position: "absolute",
		right: 8,
		width: 20,
	},
	livePill: {
		backgroundColor: "rgba(0, 0, 0, 0.5)",
		borderRadius: 99,
		left: 12,
		paddingHorizontal: 10,
		paddingVertical: 6,
		position: "absolute",
		top: 12,
	},
	monogram: {
		color: "rgba(255, 255, 255, 0.38)",
		fontSize: 27,
		fontWeight: "300",
		letterSpacing: 4,
	},
	footer: {
		alignItems: "center",
		flexDirection: "row",
		gap: 12,
		padding: 16,
	},
	copy: {
		flex: 1,
		gap: 3,
	},
	name: {
		color: colors.label,
		fontSize: 16,
		fontWeight: "700",
	},
	host: {
		color: colors.secondaryLabel,
		fontSize: 12,
	},
	chevron: {
		color: colors.secondaryLabel,
		fontSize: 26,
		fontWeight: "300",
	},
});
