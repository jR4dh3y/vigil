import type { ReactNode } from "react";
import { KeyboardAvoidingView, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";
import { ActionButton } from "@/components/action-button";
import { colors } from "@/theme/colors";

type AuthScreenProps = {
	title: string;
	description: string;
	username: string;
	password: string;
	onUsernameChange: (value: string) => void;
	onPasswordChange: (value: string) => void;
	submitLabel: string;
	onSubmit: () => void;
	submitting?: boolean;
	error?: string | null;
	footer?: ReactNode;
	passwordHint?: string;
};

export function AuthScreen({
	title,
	description,
	username,
	password,
	onUsernameChange,
	onPasswordChange,
	submitLabel,
	onSubmit,
	submitting = false,
	error,
	footer,
	passwordHint,
}: AuthScreenProps) {
	const canSubmit = username.trim().length > 0 && password.length > 0 && !submitting;

	return (
		<KeyboardAvoidingView
			behavior={process.env.EXPO_OS === "ios" ? "padding" : undefined}
			style={styles.flex}
		>
			<ScrollView
				contentContainerStyle={styles.content}
				contentInsetAdjustmentBehavior="automatic"
				keyboardShouldPersistTaps="handled"
				style={styles.flex}
			>
				<View style={styles.hero}>
					<Text style={styles.kicker}>Vigil</Text>
					<Text style={styles.title}>{title}</Text>
					<Text style={styles.description}>{description}</Text>
				</View>

				<View style={styles.card}>
					<View style={styles.field}>
						<Text style={styles.label}>Username</Text>
						<TextInput
							autoCapitalize="none"
							autoCorrect={false}
							editable={!submitting}
							onChangeText={onUsernameChange}
							placeholder="admin"
							placeholderTextColor={colors.secondaryLabel}
							style={styles.input}
							textContentType="username"
							value={username}
						/>
					</View>

					<View style={styles.field}>
						<Text style={styles.label}>Password</Text>
						<TextInput
							autoCapitalize="none"
							editable={!submitting}
							onChangeText={onPasswordChange}
							placeholder={passwordHint ?? "••••••••"}
							placeholderTextColor={colors.secondaryLabel}
							secureTextEntry
							style={styles.input}
							textContentType="password"
							value={password}
						/>
					</View>

					{error ? (
						<Text selectable style={styles.error}>
							{error}
						</Text>
					) : null}

					<ActionButton
						disabled={!canSubmit}
						label={submitLabel}
						loading={submitting}
						onPress={onSubmit}
					/>
				</View>

				{footer}
			</ScrollView>
		</KeyboardAvoidingView>
	);
}

const styles = StyleSheet.create({
	flex: {
		flex: 1,
		backgroundColor: colors.background,
	},
	content: {
		flexGrow: 1,
		gap: 24,
		justifyContent: "center",
		padding: 20,
		paddingBottom: 40,
	},
	hero: {
		gap: 8,
		paddingHorizontal: 4,
	},
	kicker: {
		color: colors.deepPurple,
		fontSize: 13,
		fontWeight: "700",
		letterSpacing: 1.2,
		textTransform: "uppercase",
	},
	title: {
		color: colors.label,
		fontSize: 30,
		fontWeight: "700",
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
	error: {
		color: colors.red,
		fontSize: 14,
		lineHeight: 20,
	},
});
