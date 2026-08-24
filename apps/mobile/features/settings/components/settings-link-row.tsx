import { type Href, router } from "expo-router";
import { Pressable, StyleSheet, Text } from "react-native";
import { colors } from "@/theme/colors";

type SettingsLinkRowProps = {
	href: Href;
	label: string;
	value?: string;
	last?: boolean;
};

export function SettingsLinkRow({ href, label, value, last = false }: SettingsLinkRowProps) {
	return (
		<Pressable
			accessibilityRole="button"
			onPress={() => router.push(href)}
			style={({ pressed }) => [
				styles.container,
				!last ? styles.divider : null,
				pressed ? styles.pressed : null,
			]}
		>
			<Text numberOfLines={1} style={styles.label}>
				{label}
			</Text>
			{value ? (
				<Text ellipsizeMode="tail" numberOfLines={1} style={styles.value}>
					{value}
				</Text>
			) : null}
			<Text style={styles.chevron}>›</Text>
		</Pressable>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		alignSelf: "stretch",
		flexDirection: "row",
		flexWrap: "nowrap",
		gap: 12,
		marginLeft: 16,
		minHeight: 56,
		paddingRight: 12,
		paddingVertical: 12,
	},
	divider: {
		borderBottomColor: colors.separator,
		borderBottomWidth: StyleSheet.hairlineWidth,
	},
	pressed: {
		opacity: 0.6,
	},
	label: {
		color: colors.label,
		flexShrink: 0,
		fontSize: 16,
		lineHeight: 20,
	},
	value: {
		color: colors.secondaryLabel,
		flexBasis: 0,
		flexGrow: 1,
		flexShrink: 1,
		fontSize: 14,
		includeFontPadding: false,
		lineHeight: 18,
		minWidth: 0,
		textAlign: "right",
	},
	chevron: {
		color: colors.secondaryLabel,
		flexShrink: 0,
		fontSize: 20,
		includeFontPadding: false,
		lineHeight: 20,
		textAlign: "center",
		width: 12,
	},
});
