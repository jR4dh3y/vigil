import { Pressable, StyleSheet, Text, View } from "react-native";
import { colors } from "@/theme/colors";

type EventFilterProps = {
	unacknowledgedOnly: boolean;
	onChange: (value: boolean) => void;
};

type FilterButtonProps = {
	label: string;
	selected: boolean;
	onPress: () => void;
};

function FilterButton({ label, selected, onPress }: FilterButtonProps) {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityState={{ selected }}
			onPress={onPress}
			style={({ pressed }) => [
				styles.button,
				selected && styles.buttonSelected,
				pressed && styles.buttonPressed,
			]}
		>
			<Text style={[styles.label, selected && styles.labelSelected]}>{label}</Text>
		</Pressable>
	);
}

export function EventFilter({ unacknowledgedOnly, onChange }: EventFilterProps) {
	return (
		<View style={styles.container}>
			<FilterButton
				label="All activity"
				onPress={() => onChange(false)}
				selected={!unacknowledgedOnly}
			/>
			<FilterButton
				label="Needs review"
				onPress={() => onChange(true)}
				selected={unacknowledgedOnly}
			/>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 14,
		flexDirection: "row",
		gap: 4,
		padding: 4,
	},
	button: {
		alignItems: "center",
		borderCurve: "continuous",
		borderRadius: 11,
		flex: 1,
		paddingHorizontal: 12,
		paddingVertical: 9,
	},
	buttonSelected: {
		backgroundColor: colors.label,
	},
	buttonPressed: {
		opacity: 0.7,
	},
	label: {
		color: colors.secondaryLabel,
		fontSize: 13,
		fontWeight: "700",
	},
	labelSelected: {
		color: colors.background,
	},
});
