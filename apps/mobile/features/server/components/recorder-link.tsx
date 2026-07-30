import { Link } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { useApiBaseUrl } from "@/features/server/use-api-base-url";
import { colors } from "@/theme/colors";

export function RecorderLink() {
	const baseUrl = useApiBaseUrl();

	return (
		<View style={styles.wrapper}>
			<Text selectable numberOfLines={1} style={styles.value}>
				{baseUrl}
			</Text>
			<Link href="/server" asChild>
				<Pressable
					accessibilityHint="Choose a different recorder"
					accessibilityRole="button"
					style={({ pressed }) => [styles.button, pressed ? styles.buttonPressed : null]}
				>
					<Text style={styles.buttonLabel}>Change recorder</Text>
				</Pressable>
			</Link>
		</View>
	);
}

const styles = StyleSheet.create({
	wrapper: {
		alignItems: "center",
		gap: 8,
	},
	value: {
		color: colors.secondaryLabel,
		fontSize: 12,
		maxWidth: "100%",
		paddingHorizontal: 8,
		textAlign: "center",
	},
	button: {
		borderCurve: "continuous",
		borderRadius: 10,
		paddingHorizontal: 12,
		paddingVertical: 7,
	},
	buttonPressed: {
		backgroundColor: colors.surface,
	},
	buttonLabel: {
		color: colors.accent,
		fontSize: 13,
		fontWeight: "700",
	},
});
