import { ScrollView, StyleSheet, Switch } from "react-native";
import { SettingsGroup } from "@/features/settings/components/settings-group";
import { SettingsRow } from "@/features/settings/components/settings-row";
import { apiBaseUrl } from "@/lib/api/config";
import { useAppStore } from "@/lib/store";
import { colors } from "@/theme/colors";

export default function SettingsScreen() {
	const armed = useAppStore((state) => state.armed);
	const setArmed = useAppStore((state) => state.setArmed);
	const notificationsEnabled = useAppStore((state) => state.notificationsEnabled);
	const setNotificationsEnabled = useAppStore((state) => state.setNotificationsEnabled);

	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			style={styles.screen}
		>
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
