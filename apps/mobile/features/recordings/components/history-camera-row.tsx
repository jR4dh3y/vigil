import { Link } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { StatusDot } from "@/components/status-dot";
import { colors } from "@/theme/colors";

type HistoryCameraRowProps = {
	id: string;
	name: string;
	host: string;
	enabled: boolean;
	status: "online" | "offline" | "unknown";
};

export function HistoryCameraRow({ id, name, host, enabled, status }: HistoryCameraRowProps) {
	return (
		<Link href={{ pathname: "/history/[id]", params: { id } }} asChild>
			<Pressable
				accessibilityHint="Browse retained video from this camera"
				accessibilityLabel={`${name} recording history`}
				accessibilityRole="button"
				style={({ pressed }) => [styles.container, pressed && styles.pressed]}
			>
				<View style={styles.copy}>
					<Text selectable numberOfLines={1} style={styles.name}>
						{name}
					</Text>
					<Text selectable numberOfLines={1} style={styles.host}>
						{host}
					</Text>
				</View>
				<View style={styles.trailing}>
					<StatusDot
						label={enabled ? status : "disabled"}
						tone={status === "unknown" || !enabled ? "neutral" : status}
					/>
					<Text style={styles.chevron}>›</Text>
				</View>
			</Pressable>
		</Link>
	);
}

const styles = StyleSheet.create({
	container: {
		alignItems: "center",
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 18,
		flexDirection: "row",
		gap: 16,
		minHeight: 72,
		paddingHorizontal: 16,
		paddingVertical: 12,
	},
	pressed: {
		opacity: 0.65,
	},
	copy: {
		flex: 1,
		gap: 4,
	},
	name: {
		color: colors.label,
		fontSize: 16,
		fontWeight: "700",
	},
	host: {
		color: colors.secondaryLabel,
		fontSize: 12,
	},
	trailing: {
		alignItems: "center",
		flexDirection: "row",
		gap: 10,
	},
	chevron: {
		color: colors.secondaryLabel,
		fontSize: 24,
		fontWeight: "300",
	},
});
