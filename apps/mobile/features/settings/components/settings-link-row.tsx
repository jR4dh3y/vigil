import { type Href, Link } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { colors } from "@/theme/colors";

type SettingsLinkRowProps = {
	href: Href;
	label: string;
	value?: string;
	last?: boolean;
};

export function SettingsLinkRow({ href, label, value, last = false }: SettingsLinkRowProps) {
	return (
		<Link href={href} asChild>
			<Pressable
				accessibilityRole="button"
				style={({ pressed }) => [
					styles.container,
					!last ? styles.divider : null,
					pressed ? styles.pressed : null,
				]}
			>
				<Text numberOfLines={1} style={styles.label}>
					{label}
				</Text>
				<View style={styles.trailing}>
					{value ? (
						<Text ellipsizeMode="tail" numberOfLines={1} style={styles.value}>
							{value}
						</Text>
					) : null}
					<Text style={styles.chevron}>›</Text>
				</View>
			</Pressable>
		</Link>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		flexDirection: "row",
		gap: 12,
		justifyContent: "space-between",
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
		flexShrink: 1,
		fontSize: 16,
		lineHeight: 20,
		minWidth: 0,
	},
	trailing: {
		alignItems: "center",
		flexDirection: "row",
		flexShrink: 1,
		gap: 4,
		justifyContent: "flex-end",
		maxWidth: "68%",
		minWidth: 0,
	},
	value: {
		color: colors.secondaryLabel,
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
