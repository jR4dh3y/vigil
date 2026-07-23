import { Redirect } from "expo-router";
import { ActivityIndicator, StyleSheet, View } from "react-native";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { colors } from "@/theme/colors";

/** Entry redirect so deep links and cold start land on the right surface. */
export default function Index() {
	const auth = useAuthStatus();

	if (auth.isPending) {
		return (
			<View style={styles.boot}>
				<ActivityIndicator color={colors.accent} size="large" />
			</View>
		);
	}

	if (auth.isError || !auth.data) {
		return <Redirect href="/login" />;
	}

	if (auth.data.setupRequired) {
		return <Redirect href="/setup" />;
	}

	if (!auth.data.user) {
		return <Redirect href="/login" />;
	}

	return <Redirect href="/(tabs)/(live)" />;
}

const styles = StyleSheet.create({
	boot: {
		alignItems: "center",
		backgroundColor: colors.background,
		flex: 1,
		justifyContent: "center",
	},
});
