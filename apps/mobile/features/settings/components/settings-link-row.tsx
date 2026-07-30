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
				<Text selectable style={styles.label}>
					{label}
				</Text>
				<View style={styles.trailing}>
					{value ? (
						<Text selectable numberOfLines={1} style={styles.value}>
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
	pressed: {
		opacity: 0.65,
	},
	label: {
		color: colors.label,
		flex: 1,
		fontSize: 16,
	},
	trailing: {
		alignItems: "center",
		flexDirection: "row",
		gap: 8,
		maxWidth: "60%",
	},
	value: {
		color: colors.secondaryLabel,
		flexShrink: 1,
		fontSize: 14,
	},
	chevron: {
		color: colors.secondaryLabel,
		fontSize: 24,
		fontWeight: "300",
	},
});
