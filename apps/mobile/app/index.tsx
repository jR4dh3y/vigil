import { StyleSheet, Text, View } from "react-native";

export default function HomeScreen() {
	return (
		<View style={styles.container}>
			<Text style={styles.title}>NVR Mobile</Text>
			<Text style={styles.sub}>Live view · events · notifications (not an admin surface)</Text>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		flex: 1,
		alignItems: "center",
		justifyContent: "center",
		padding: 24,
		gap: 8,
	},
	title: { fontSize: 22, fontWeight: "600" },
	sub: { fontSize: 14, opacity: 0.6, textAlign: "center" },
});
