import { StyleSheet, Text, View } from "react-native";
import { colors, swatches } from "@/theme/colors";

type StatusTone = "online" | "warning" | "offline" | "neutral";

type StatusProps = {
	label: string;
	tone: StatusTone;
	inverted?: boolean;
};

const toneColor = {
	online: colors.green,
	warning: colors.orange,
	offline: colors.red,
	neutral: colors.secondaryLabel,
} satisfies Record<StatusTone, string | object>;

export function Status({ label, tone, inverted = false }: StatusProps) {
	return (
		<View style={styles.container}>
			<Text style={[styles.label, { color: toneColor[tone] }, inverted && styles.labelInverted]}>
				{label}
			</Text>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		flexDirection: "row",
		gap: 6,
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
