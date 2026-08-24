import { Redirect } from "expo-router";
import { NativeTabs } from "expo-router/unstable-native-tabs";
import { ActivityIndicator, StyleSheet, View } from "react-native";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { useEventAlerts } from "@/features/notifications/use-event-alerts";
import { useApiConfiguration } from "@/features/server/use-api-base-url";
import { colors } from "@/theme/colors";

export default function TabsLayout() {
	const configuration = useApiConfiguration();
	const configured = configuration.kind === "configured";
	const auth = useAuthStatus(configured);
	useEventAlerts(Boolean(auth.data?.user));

	if (!configured) {
		return <Redirect href="/server" />;
	}

	if (auth.isPending) {
		return (
			<View style={styles.boot}>
				<ActivityIndicator color={colors.accent} size="large" />
			</View>
		);
	}

	if (auth.isError) {
		return <Redirect href="/server" />;
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
			<NativeTabs.Trigger name="(history)">
				<NativeTabs.Trigger.Icon sf="clock.arrow.circlepath" md="history" />
				<NativeTabs.Trigger.Label>History</NativeTabs.Trigger.Label>
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
