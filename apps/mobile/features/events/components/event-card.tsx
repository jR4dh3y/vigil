import type { Event, EventSeverity } from "@nvr/api-client";
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { eventEyebrow, formatEventTime, severityLabel } from "@/features/events/format";
import { colors, swatches } from "@/theme/colors";

type EventCardProps = {
	event: Event;
	acknowledging: boolean;
	onAcknowledge: (id: string) => void;
};

const severityStyle = {
	info: { color: colors.green, backgroundColor: swatches.greenSoft },
	warning: { color: colors.orange, backgroundColor: swatches.orangeSoft },
	critical: { color: colors.red, backgroundColor: swatches.redSoft },
} satisfies Record<EventSeverity, { color: string | object; backgroundColor: string }>;

export function EventCard({ event, acknowledging, onAcknowledge }: EventCardProps) {
	const severity = severityStyle[event.severity];

	return (
		<View style={[styles.container, !event.acknowledged && styles.unread]}>
			<View style={styles.topRow}>
				<View style={[styles.badge, { backgroundColor: severity.backgroundColor }]}>
					<View style={[styles.badgeDot, { backgroundColor: severity.color }]} />
					<Text style={[styles.badgeLabel, { color: severity.color }]}>
						{severityLabel(event.severity)}
					</Text>
				</View>
				<Text selectable style={styles.time}>
					{formatEventTime(event.startedAt)}
				</Text>
			</View>

			<View style={styles.copy}>
				<Text selectable style={styles.eyebrow}>
					{eventEyebrow(event)}
				</Text>
				<Text selectable style={styles.title}>
					{event.title}
				</Text>
				<Text selectable style={styles.message}>
					{event.message}
				</Text>
			</View>

			{event.acknowledged ? (
				<Text style={styles.reviewed}>Reviewed</Text>
			) : (
				<Pressable
					accessibilityLabel={`Mark ${event.title} as reviewed`}
					accessibilityRole="button"
					disabled={acknowledging}
					onPress={() => onAcknowledge(event.id)}
					style={({ pressed }) => [styles.action, pressed && styles.actionPressed]}
				>
					{acknowledging ? (
						<ActivityIndicator color={colors.label} size="small" />
					) : (
						<Text style={styles.actionLabel}>Mark as reviewed</Text>
					)}
				</Pressable>
			)}
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		backgroundColor: colors.surface,
		borderColor: "transparent",
		borderCurve: "continuous",
		borderRadius: 22,
		borderWidth: 1,
		gap: 16,
		padding: 18,
	},
	unread: {
		borderColor: "rgba(255, 149, 0, 0.24)",
	},
	topRow: {
		alignItems: "center",
		flexDirection: "row",
		justifyContent: "space-between",
	},
	badge: {
		alignItems: "center",
		borderRadius: 99,
		flexDirection: "row",
		gap: 6,
		paddingHorizontal: 9,
		paddingVertical: 6,
	},
	badgeDot: {
		borderRadius: 99,
		height: 6,
		width: 6,
	},
	badgeLabel: {
		fontSize: 11,
		fontWeight: "800",
	},
	time: {
		color: colors.secondaryLabel,
		fontSize: 12,
		fontVariant: ["tabular-nums"],
	},
	copy: {
		gap: 5,
	},
	eyebrow: {
		color: colors.secondaryLabel,
		fontSize: 10,
		fontWeight: "800",
		letterSpacing: 1.1,
	},
	title: {
		color: colors.label,
		fontSize: 18,
		fontWeight: "700",
		letterSpacing: -0.3,
	},
	message: {
		color: colors.secondaryLabel,
		fontSize: 14,
		lineHeight: 20,
	},
	action: {
		alignItems: "center",
		backgroundColor: swatches.neutralSoft,
		borderCurve: "continuous",
		borderRadius: 12,
		minHeight: 40,
		justifyContent: "center",
		paddingHorizontal: 14,
		paddingVertical: 10,
	},
	actionPressed: {
		opacity: 0.6,
	},
	actionLabel: {
		color: colors.label,
		fontSize: 13,
		fontWeight: "700",
	},
	reviewed: {
		color: colors.secondaryLabel,
		fontSize: 12,
		fontWeight: "600",
	},
});
