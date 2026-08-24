import type { SystemStatus } from "@nvr/api-client";
import { StyleSheet, Text, View } from "react-native";
import { SettingsGroup } from "@/features/settings/components/settings-group";
import { SettingsRow } from "@/features/settings/components/settings-row";
import { formatDiskUsage } from "@/features/system/format";
import { colors, swatches } from "@/theme/colors";

type SystemSummaryProps = {
	status: SystemStatus;
};

export function SystemSummary({ status }: SystemSummaryProps) {
	const degraded = status.health.status === "degraded";

	return (
		<SettingsGroup title="Recorder health">
			<SettingsRow
				control={
				<Text style={[styles.badgeLabel, degraded ? styles.warningLabel : styles.healthyLabel]}>
					{degraded ? "Degraded" : "Healthy"}
				</Text>
				}
				label="Status"
			/>
			<SettingsRow
				detail={`${status.disk.usedPercent.toFixed(0)}% used`}
				label="Storage"
				value={formatDiskUsage(status.disk.usedBytes, status.disk.totalBytes)}
			/>
			<SettingsRow
				detail={`${status.cameras.enabled} enabled`}
				label="Cameras"
				value={`${status.cameras.online} of ${status.cameras.total} online`}
			/>
			<SettingsRow label="Retention" value={`${status.retentionDays} days`} />
			<SettingsRow label="Recorder version" last value={status.version} />
		</SettingsGroup>
	);
}

const styles = StyleSheet.create({
	badge: {
		borderCurve: "continuous",
		borderRadius: 99,
		paddingHorizontal: 10,
		paddingVertical: 5,
	},
	healthy: {
		backgroundColor: swatches.greenSoft,
	},
	warning: {
		backgroundColor: swatches.orangeSoft,
	},
	badgeLabel: {
		fontSize: 12,
		fontWeight: "700",
	},
	healthyLabel: {
		color: colors.green,
	},
	warningLabel: {
		color: colors.orange,
	},
});
