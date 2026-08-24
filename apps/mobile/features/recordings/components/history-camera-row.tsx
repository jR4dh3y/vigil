import { type Href, router } from "expo-router";
import { Pressable, StyleSheet, Text } from "react-native";
import { Status } from "@/components/status-without-dot";
import { colors } from "@/theme/colors";

type HistoryCameraRowProps = {
	id: string;
	name: string;
	enabled: boolean;
	status: "online" | "offline" | "unknown";
};

export function HistoryCameraRow({ id, name, enabled, status }: HistoryCameraRowProps) {
	return (
		<Pressable
			accessibilityHint="Browse retained video from this camera"
			accessibilityLabel={`${name} recording history`}
			accessibilityRole="button"
			onPress={() => router.push({ pathname: "/history/[id]", params: { id } } as Href)}
			style={({ pressed }) => [styles.container, pressed ? styles.pressed : null]}
		>
			<Text ellipsizeMode="tail" numberOfLines={1} style={styles.name}>
				{name}
			</Text>
			<Status
				label={enabled ? status : "disabled"}
				tone={status === "unknown" || !enabled ? "neutral" : status}
			/>
			<Text style={styles.chevron}>›</Text>
		</Pressable>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		alignSelf: "stretch",
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 14,
		flexDirection: "row",
		flexWrap: "nowrap",
		gap: 12,
		minHeight: 56,
		paddingLeft: 16,
		paddingRight: 12,
		paddingVertical: 12,
	},
	pressed: {
		opacity: 0.6,
	},
	name: {
		color: colors.label,
		flexBasis: 0,
		flexGrow: 1,
		flexShrink: 1,
		fontSize: 16,
		lineHeight: 20,
		minWidth: 0,
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
