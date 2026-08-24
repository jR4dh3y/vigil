import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Constants from "expo-constants";
import { router } from "expo-router";
import { Alert, ScrollView, StyleSheet, Switch } from "react-native";
import { ActionButton } from "@/components/action-button";
import { logout } from "@/features/auth/api";
import { authKeys } from "@/features/auth/keys";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { useNotificationPreference } from "@/features/notifications/use-notification-preference";
import { useApiBaseUrl } from "@/features/server/use-api-base-url";
import { SettingsGroup } from "@/features/settings/components/settings-group";
import { SettingsLinkRow } from "@/features/settings/components/settings-link-row";
import { SettingsRow } from "@/features/settings/components/settings-row";
import { getSystemStatus, systemKeys } from "@/features/system/api";
import { SystemSummary } from "@/features/system/components/system-summary";
import { colors } from "@/theme/colors";

export default function SettingsScreen() {
	const queryClient = useQueryClient();
	const auth = useAuthStatus();
	const user = auth.data?.user;
	const notificationPreference = useNotificationPreference();
	const apiBaseUrl = useApiBaseUrl();
	const statusQuery = useQuery({
		queryKey: systemKeys.status(apiBaseUrl),
		queryFn: getSystemStatus,
		refetchInterval: 30_000,
	});

	const logoutMutation = useMutation({
		mutationFn: logout,
		onSuccess: async () => {
			await queryClient.resetQueries({ queryKey: authKeys.all });
			queryClient.clear();
			router.replace("/login");
		},
		onError: (err: unknown) => {
			Alert.alert("Sign out failed", err instanceof Error ? err.message : "Try again");
		},
	});

	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			style={styles.screen}
		>
			<SettingsGroup title="Account">
				<SettingsRow label="Signed in as" value={user?.username ?? "—"} />
				<SettingsRow label="Role" last value={user?.role ?? "—"} />
			</SettingsGroup>

			<SettingsGroup title="Alerts">
				<SettingsRow
					control={
						<Switch
							accessibilityLabel="Alerts while app is open"
							disabled={!notificationPreference.available || notificationPreference.requesting}
							onValueChange={(enabled) => {
								void notificationPreference.update(enabled).then((granted) => {
									if (enabled && !granted) {
										Alert.alert(
											"Notifications are off",
											"Allow notifications in system settings, then try again.",
										);
									}
								});
							}}
							trackColor={{ true: colors.lime }}
							value={notificationPreference.enabled}
						/>
					}
					detail={
						notificationPreference.available
							? "Warning and critical system alerts"
							: "Unavailable in Expo Go and on web"
					}
					label="Alerts while open"
					last
				/>
			</SettingsGroup>

			{statusQuery.data ? <SystemSummary status={statusQuery.data} /> : null}

			<SettingsGroup
				footer={statusQuery.isError ? `Status unavailable: ${statusQuery.error.message}` : ""}
				title="Recorder"
			>
				<SettingsLinkRow href="/server" label="Server" last value={apiBaseUrl} />
			</SettingsGroup>

			<SettingsGroup title="About">
				<SettingsRow label="App" value="Vigil" />
				<SettingsRow label="Version" value={Constants.expoConfig?.version ?? "—"} />
			</SettingsGroup>

			<ActionButton
				danger
				label={logoutMutation.isPending ? "Signing out…" : "Sign out"}
				loading={logoutMutation.isPending}
				onPress={() => logoutMutation.mutate()}
				variant="secondary"
			/>
		</ScrollView>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		gap: 24,
		padding: 16,
		paddingBottom: 40,
	},
});
