import { Redirect, router } from "expo-router";
import { useEffect } from "react";
import { ActivityIndicator, StyleSheet, View } from "react-native";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import {
	takePendingEventRoute,
	usePendingEventRoute,
} from "@/features/notifications/pending-event-route";
import { useApiConfiguration } from "@/features/server/use-api-base-url";
import { colors } from "@/theme/colors";

/** Entry redirect so deep links and cold start land on the right surface. */
export default function Index() {
	const configuration = useApiConfiguration();
	const configured = configuration.kind === "configured";
	const auth = useAuthStatus(configured);
	const pendingEventId = usePendingEventRoute();

	useEffect(() => {
		if (!auth.data?.user || !pendingEventId) {
			return;
		}
		const eventId = takePendingEventRoute();
		if (eventId) {
			router.replace({ pathname: "/event/[id]", params: { id: eventId } });
		}
	}, [auth.data?.user, pendingEventId]);

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

	if (auth.isError || !auth.data) {
		return <Redirect href="/server" />;
	}

	if (auth.data.setupRequired) {
		return <Redirect href="/setup" />;
	}

	if (!auth.data.user) {
		return <Redirect href="/login" />;
	}
	if (pendingEventId) {
		return (
			<View style={styles.boot}>
				<ActivityIndicator color={colors.accent} size="large" />
			</View>
		);
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
