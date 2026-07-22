import type { ReactNode } from "react";
import { StyleSheet, Text, View } from "react-native";
import { colors } from "@/theme/colors";

type SettingsRowProps = {
	label: string;
	detail?: string;
	value?: string;
	control?: ReactNode;
	last?: boolean;
};

export function SettingsRow({ label, detail, value, control, last = false }: SettingsRowProps) {
	return (
		<View style={[styles.container, !last && styles.divider]}>
			<View style={styles.copy}>
				<Text selectable style={styles.label}>
					{label}
				</Text>
				{detail ? (
					<Text selectable style={styles.detail}>
						{detail}
					</Text>
				) : null}
			</View>
			{value ? (
				<Text selectable numberOfLines={1} style={styles.value}>
					{value}
				</Text>
			) : null}
			{control}
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		flexDirection: "row",
		gap: 16,
		marginLeft: 16,
		minHeight: 58,
		paddingRight: 16,
		paddingVertical: 11,
	},
	divider: {
		borderBottomColor: colors.separator,
		borderBottomWidth: StyleSheet.hairlineWidth,
	},
	copy: {
		flex: 1,
		gap: 3,
	},
	label: {
		color: colors.label,
		fontSize: 16,
	},
	detail: {
		color: colors.secondaryLabel,
		fontSize: 12,
	},
	value: {
		color: colors.secondaryLabel,
		fontSize: 14,
		maxWidth: "48%",
	},
});
