import { Link } from "expo-router";
import { StyleSheet, Text, View } from "react-native";
import { ActionButton } from "@/components/action-button";
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
				<ActionButton accessibilityHint="Choose a different recorder" label="Change recorder" variant="secondary" />
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
});
