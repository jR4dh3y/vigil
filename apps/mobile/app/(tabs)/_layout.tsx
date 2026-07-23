import { Redirect } from "expo-router";
import { NativeTabs } from "expo-router/unstable-native-tabs";
import { ActivityIndicator, StyleSheet, View } from "react-native";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { colors } from "@/theme/colors";

export default function TabsLayout() {
	const auth = useAuthStatus();

	if (auth.isPending) {
		return (
			<View style={styles.boot}>
				<ActivityIndicator color={colors.accent} size="large" />
			</View>
		);
	}

	if (auth.isError) {
		return <Redirect href="/login" />;
	}

	if (auth.data.setupRequired) {
		return <Redirect href="/setup" />;
	}

	if (!auth.data.user) {
		return <Redirect href="/login" />;
	}

	return (
		<NativeTabs>
			<NativeTabs.Trigger name="(live)">
				<NativeTabs.Trigger.Icon sf="rectangle.grid.2x2.fill" md="dashboard" />
				<NativeTabs.Trigger.Label>Live</NativeTabs.Trigger.Label>
			</NativeTabs.Trigger>
			<NativeTabs.Trigger name="(events)">
				<NativeTabs.Trigger.Icon sf="bell.fill" md="notifications" />
				<NativeTabs.Trigger.Label>Events</NativeTabs.Trigger.Label>
			</NativeTabs.Trigger>
			<NativeTabs.Trigger name="(settings)">
				<NativeTabs.Trigger.Icon sf="gearshape.fill" md="settings" />
				<NativeTabs.Trigger.Label>Settings</NativeTabs.Trigger.Label>
			</NativeTabs.Trigger>
		</NativeTabs>
	);
}

const styles = StyleSheet.create({
	boot: {
		alignItems: "center",
		backgroundColor: colors.background,
		flex: 1,
		justifyContent: "center",
	},
});
