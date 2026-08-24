import { Link, Stack } from "expo-router";
import { ScrollView, StyleSheet, Text } from "react-native";
import { colors } from "@/theme/colors";

export default function NotFoundScreen() {
	return (
		<ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={styles.content}>
			<Stack.Title>Not found</Stack.Title>
			<Text selectable style={styles.title}>
				This screen is not available.
			</Text>
			<Link href="/" style={styles.link}>
				Return to live view
			</Link>
		</ScrollView>
	);
}

const styles = StyleSheet.create({
	content: {
		backgroundColor: colors.background,
		flexGrow: 1,
		gap: 14,
		justifyContent: "center",
		padding: 24,
	},
	title: {
		color: colors.label,
		fontSize: 20,
		fontWeight: "700",
	},
	link: {
		color: colors.deepPurple,
		fontSize: 16,
		fontWeight: "700",
	},
});
