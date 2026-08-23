import {
	ActivityIndicator,
	KeyboardAvoidingView,
	Platform,
	Pressable,
	ScrollView,
	StyleSheet,
	Text,
	TextInput,
	View,
} from "react-native";
import { colors, swatches } from "@/theme/colors";

type ServerConnectionScreenProps = {
	value: string;
	error: string | null;
	connecting: boolean;
	onChange: (value: string) => void;
	onConnect: () => void;
};

export function ServerConnectionScreen({
	value,
	error,
	connecting,
	onChange,
	onConnect,
}: ServerConnectionScreenProps) {
	const canConnect = value.trim().length > 0 && !connecting;

	return (
		<KeyboardAvoidingView
			behavior={Platform.OS === "ios" ? "padding" : undefined}
			style={styles.screen}
		>
			<ScrollView
				automaticallyAdjustKeyboardInsets
				contentInsetAdjustmentBehavior="automatic"
				contentContainerStyle={styles.content}
				keyboardShouldPersistTaps="handled"
				style={styles.screen}
			>
			<View style={styles.hero}>
				<Text selectable style={styles.title}>
					Connect to your recorder
				</Text>
				<Text selectable style={styles.description}>
					Enter the LAN or HTTPS address you use to reach Vigil. The API path is added
					automatically.
				</Text>
			</View>

			<View style={styles.card}>
				<View style={styles.field}>
					<Text style={styles.label}>Recorder address</Text>
					<TextInput
						autoCapitalize="none"
						autoCorrect={false}
						editable={!connecting}
						keyboardType="url"
						onChangeText={onChange}
						onSubmitEditing={canConnect ? onConnect : undefined}
						placeholder="https://vigil.example.com"
						placeholderTextColor={colors.secondaryLabel}
						returnKeyType="go"
						style={styles.input}
						value={value}
					/>
				</View>
				<Text selectable style={styles.hint}>
					Use HTTPS outside your trusted local network. Credentials never belong in this URL.
				</Text>
				{error ? (
					<Text selectable style={styles.error}>
						{error}
					</Text>
				) : null}
				<Pressable
					accessibilityRole="button"
					disabled={!canConnect}
					onPress={onConnect}
					style={({ pressed }) => [
						styles.connect,
						!canConnect || pressed ? styles.connectDisabled : null,
					]}
				>
					{connecting ? (
						<ActivityIndicator color={swatches.white} />
					) : (
						<Text style={styles.connectLabel}>Test and connect</Text>
					)}
				</Pressable>
			</View>
		</ScrollView>
		</KeyboardAvoidingView>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		gap: 24,
		padding: 20,
		paddingBottom: 32,
	},
	hero: {
		gap: 8,
		paddingHorizontal: 4,
	},
	title: {
		color: colors.label,
		fontSize: 28,
		fontWeight: "700",
		letterSpacing: -0.6,
	},
	description: {
		color: colors.secondaryLabel,
		fontSize: 15,
		lineHeight: 22,
	},
	card: {
		backgroundColor: colors.surface,
		borderCurve: "continuous",
		borderRadius: 22,
		gap: 16,
		padding: 18,
	},
	field: {
		gap: 8,
	},
	label: {
		color: colors.secondaryLabel,
		fontSize: 13,
		fontWeight: "600",
	},
	input: {
		backgroundColor: colors.background,
		borderColor: colors.separator,
		borderCurve: "continuous",
		borderRadius: 14,
		borderWidth: StyleSheet.hairlineWidth,
		color: colors.label,
		fontSize: 16,
		paddingHorizontal: 14,
		paddingVertical: 12,
	},
	hint: {
		color: colors.secondaryLabel,
		fontSize: 12,
		lineHeight: 17,
	},
	error: {
		color: colors.red,
		fontSize: 14,
		lineHeight: 20,
	},
	connect: {
		alignItems: "center",
		backgroundColor: colors.label,
		borderCurve: "continuous",
		borderRadius: 14,
		justifyContent: "center",
		minHeight: 48,
		paddingHorizontal: 16,
	},
	connectDisabled: {
		opacity: 0.55,
	},
	connectLabel: {
		color: swatches.white,
		fontSize: 16,
		fontWeight: "700",
	},
});
