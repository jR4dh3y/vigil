import { StyleSheet, Text, View } from "react-native";
import { colors, swatches } from "@/theme/colors";

type StatusTone = "online" | "warning" | "offline" | "neutral";

type StatusDotProps = {
	label: string;
	tone: StatusTone;
	inverted?: boolean;
};

// Bright landing hues for the dot fill; labels stay neutral for contrast.
const toneColor = {
	online: colors.lime,
	warning: colors.pop,
	offline: colors.coral,
	neutral: colors.secondaryLabel,
} satisfies Record<StatusTone, string | object>;

export function StatusDot({ label, tone, inverted = false }: StatusDotProps) {
	return (
		<View style={styles.container}>
			<View style={[styles.dot, { backgroundColor: toneColor[tone] }]} />
			<Text style={[styles.label, inverted && styles.labelInverted]}>{label}</Text>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		flexDirection: "row",
		gap: 6,
	},
	dot: {
		borderRadius: 99,
		height: 7,
		width: 7,
	},
	label: {
		color: colors.secondaryLabel,
		fontSize: 12,
		fontWeight: "600",
	},
	labelInverted: {
		color: swatches.white,
	},
});
