import { useState } from "react";
import {
	ActivityIndicator,
	type GestureResponderEvent,
	Pressable,
	StyleSheet,
	Text,
	type ViewStyle,
} from "react-native";
import { colors } from "@/theme/colors";

type ActionButtonProps = {
	label: string;
	onPress?: (event: GestureResponderEvent) => void;
	/** Primary is the solid lavender landing CTA; secondary is the outlined chip. */
	variant?: "primary" | "secondary";
	/** Secondary only: coral border and label for destructive actions. */
	danger?: boolean;
	disabled?: boolean;
	loading?: boolean;
	accessibilityHint?: string;
};

/** Editorial button matching the landing CTA language: hard ink border,
 * uppercase label, inverts to ink-on-paper while pressed. */
export function ActionButton({
	label,
	onPress,
	variant = "primary",
	danger = false,
	disabled = false,
	loading = false,
	accessibilityHint,
}: ActionButtonProps) {
	const [pressed, setPressed] = useState(false);
	const interactive = !disabled && !loading;
	const inverted = pressed && interactive;
	const containerStyle: ViewStyle[] = [styles.base, styles[variant]];
	if (danger && variant === "secondary") {
		containerStyle.push(inverted ? styles.dangerInverted : styles.danger);
	}
	if (inverted) {
		containerStyle.push(styles.inverted);
	}
	if (!interactive) {
		containerStyle.push(styles.dimmed);
	}
	return (
		<Pressable
			accessibilityHint={accessibilityHint}
			accessibilityRole="button"
			accessibilityState={{ busy: loading, disabled: !interactive }}
			disabled={!interactive}
			onPress={onPress}
			onPressIn={() => setPressed(true)}
			onPressOut={() => setPressed(false)}
			style={containerStyle}
		>
			{loading ? (
				<ActivityIndicator
					color={inverted || variant === "secondary" ? colors.paper : colors.ink}
				/>
			) : (
				<Text
					style={[
						styles.label,
						danger && !inverted && styles.dangerLabel,
						inverted && styles.labelInverted,
					]}
				>
					{label}
				</Text>
			)}
		</Pressable>
	);
}

const styles = StyleSheet.create({
	base: {
		alignItems: "center",
		borderCurve: "continuous",
		borderRadius: 12,
		borderWidth: 2,
		justifyContent: "center",
		minHeight: 48,
		paddingHorizontal: 18,
	},
	primary: {
		backgroundColor: colors.accent,
		borderColor: colors.ink,
	},
	secondary: {
		backgroundColor: "transparent",
		borderColor: colors.ink,
		minHeight: 44,
	},
	inverted: {
		backgroundColor: colors.ink,
		borderColor: colors.ink,
	},
	danger: {
		borderColor: colors.coral,
	},
	dangerInverted: {
		borderColor: colors.ink,
	},
	dimmed: {
		opacity: 0.45,
	},
	label: {
		color: colors.ink,
		fontSize: 13,
		fontWeight: "800",
		letterSpacing: 1,
		textTransform: "uppercase",
	},
	dangerLabel: {
		color: colors.coral,
	},
	labelInverted: {
		color: colors.paper,
	},
});
