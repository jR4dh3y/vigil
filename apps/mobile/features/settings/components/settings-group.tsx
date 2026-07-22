import type { ReactNode } from "react";
import { StyleSheet, Text, View } from "react-native";
import { colors } from "@/theme/colors";

type SettingsGroupProps = {
	title: string;
	children: ReactNode;
	footer?: string;
};

export function SettingsGroup({ title, children, footer }: SettingsGroupProps) {
	return (
		<View style={styles.wrapper}>
			<Text selectable style={styles.title}>
				{title.toUpperCase()}
			</Text>
			<View style={styles.container}>{children}</View>
			{footer ? (
				<Text selectable style={styles.footer}>
					{footer}
				</Text>
			) : null}
		</View>
	);
}

const styles = StyleSheet.create({
	wrapper: {
		gap: 8,
	},
	title: {
		color: colors.secondaryLabel,
		fontSize: 11,
		fontWeight: "700",
		letterSpacing: 0.8,
		paddingHorizontal: 6,
	},
	container: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 18,
		overflow: "hidden",
	},
	footer: {
		color: colors.secondaryLabel,
		fontSize: 12,
		lineHeight: 17,
		paddingHorizontal: 6,
	},
});
