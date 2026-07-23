import { useMutation, useQueryClient } from "@tanstack/react-query";
import { router } from "expo-router";
import { Alert, Pressable, ScrollView, StyleSheet, Switch, Text } from "react-native";
import { logout } from "@/features/auth/api";
import { authKeys } from "@/features/auth/keys";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { SettingsGroup } from "@/features/settings/components/settings-group";
import { SettingsRow } from "@/features/settings/components/settings-row";
import { apiBaseUrl } from "@/lib/api/config";
import { useAppStore } from "@/lib/store";
import { colors } from "@/theme/colors";

export default function SettingsScreen() {
	const queryClient = useQueryClient();
	const auth = useAuthStatus();
	const user = auth.data?.user;
	const armed = useAppStore((state) => state.armed);
	const setArmed = useAppStore((state) => state.setArmed);
	const notificationsEnabled = useAppStore((state) => state.notificationsEnabled);
	const setNotificationsEnabled = useAppStore((state) => state.setNotificationsEnabled);

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

			<SettingsGroup
				footer="Arming controls whether this device surfaces new activity alerts. Live video remains available."
				title="Monitoring"
			>
				<SettingsRow
					control={
						<Switch onValueChange={setArmed} trackColor={{ true: colors.green }} value={armed} />
					}
					detail={armed ? "Camera alerts are active" : "Alerts are paused"}
					label="Armed"
				/>
				<SettingsRow
					control={
						<Switch
							onValueChange={setNotificationsEnabled}
							trackColor={{ true: colors.green }}
							value={notificationsEnabled}
						/>
					}
					detail="Critical and warning events"
					label="Notifications"
					last
				/>
			</SettingsGroup>

			<SettingsGroup
				footer="Set EXPO_PUBLIC_API_URL to point this app at your recorder. The URL is bundled with the app and must not contain credentials."
				title="Recorder"
			>
				<SettingsRow label="Server" last value={apiBaseUrl} />
			</SettingsGroup>

			<SettingsGroup title="About">
				<SettingsRow label="App" value="NVR Mobile" />
				<SettingsRow label="Scope" last value="Live · Events · Alerts" />
			</SettingsGroup>

			<Pressable
				accessibilityRole="button"
				disabled={logoutMutation.isPending}
				onPress={() => logoutMutation.mutate()}
				style={({ pressed }) => [styles.signOut, pressed && styles.signOutPressed]}
			>
				<Text style={styles.signOutLabel}>
					{logoutMutation.isPending ? "Signing out…" : "Sign out"}
				</Text>
			</Pressable>
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
	signOut: {
		alignItems: "center",
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 14,
		justifyContent: "center",
		minHeight: 50,
		paddingHorizontal: 16,
	},
	signOutPressed: {
		opacity: 0.7,
	},
	signOutLabel: {
		color: colors.red,
		fontSize: 16,
		fontWeight: "700",
	},
});
