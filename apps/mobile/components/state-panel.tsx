import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { colors, swatches } from "@/theme/colors";

type StatePanelProps = {
	title: string;
	detail: string;
	actionLabel?: string;
	onAction?: () => void;
	loading?: boolean;
};

export function StatePanel({
	title,
	detail,
	actionLabel,
	onAction,
	loading = false,
}: StatePanelProps) {
	return (
		<View style={styles.container}>
			{loading ? <ActivityIndicator color={colors.accent} /> : <View style={styles.indicator} />}
			<View style={styles.copy}>
				<Text selectable style={styles.title}>
					{title}
				</Text>
				<Text selectable style={styles.detail}>
					{detail}
				</Text>
			</View>
			{actionLabel && onAction ? (
				<Pressable
					accessibilityRole="button"
					onPress={onAction}
					style={({ pressed }) => [styles.action, pressed && styles.actionPressed]}
				>
					<Text style={styles.actionLabel}>{actionLabel}</Text>
				</Pressable>
			) : null}
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 22,
		gap: 16,
		paddingHorizontal: 24,
		paddingVertical: 34,
	},
	indicator: {
		backgroundColor: swatches.orangeSoft,
		borderColor: colors.accent,
		borderRadius: 12,
		borderWidth: 5,
		height: 24,
		width: 24,
	},
	copy: {
		alignItems: "center",
		gap: 6,
	},
	title: {
		color: colors.label,
		fontSize: 17,
		fontWeight: "700",
	},
	detail: {
		color: colors.secondaryLabel,
		fontSize: 14,
		lineHeight: 20,
		maxWidth: 320,
		textAlign: "center",
	},
	action: {
		backgroundColor: colors.label,
		borderCurve: "continuous",
		borderRadius: 12,
		paddingHorizontal: 18,
		paddingVertical: 10,
	},
	actionPressed: {
		opacity: 0.65,
	},
	actionLabel: {
		color: colors.background,
		fontSize: 14,
		fontWeight: "700",
	},
});
