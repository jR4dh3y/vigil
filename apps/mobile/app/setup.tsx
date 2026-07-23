import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Redirect, router } from "expo-router";
import { useState } from "react";
import { StyleSheet, Text } from "react-native";
import { setup } from "@/features/auth/api";
import { AuthScreen } from "@/features/auth/components/auth-screen";
import { authKeys } from "@/features/auth/keys";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { apiBaseUrl } from "@/lib/api/config";
import { ApiError } from "@/lib/api/error";
import { colors } from "@/theme/colors";

export default function SetupScreen() {
	const queryClient = useQueryClient();
	const auth = useAuthStatus();
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState<string | null>(null);

	const setupMutation = useMutation({
		mutationFn: () => setup({ username: username.trim(), password }),
		onSuccess: async () => {
			setError(null);
			await queryClient.invalidateQueries({ queryKey: authKeys.all });
			router.replace("/(tabs)/(live)");
		},
		onError: (err: unknown) => {
			if (err instanceof ApiError) {
				setError(err.message);
				return;
			}
			setError(err instanceof Error ? err.message : "Setup failed");
		},
	});

	if (auth.isSuccess && !auth.data.setupRequired) {
		return <Redirect href={auth.data.user ? "/(tabs)/(live)" : "/login"} />;
	}

	return (
		<AuthScreen
			description="Create the first admin account for this recorder. Password must be at least 8 characters."
			error={error}
			footer={<Text style={styles.server}>Server · {apiBaseUrl}</Text>}
			onPasswordChange={setPassword}
			onSubmit={() => {
				setError(null);
				if (password.length < 8) {
					setError("Password must be at least 8 characters");
					return;
				}
				setupMutation.mutate();
			}}
			onUsernameChange={setUsername}
			password={password}
			passwordHint="At least 8 characters"
			submitLabel="Create admin"
			submitting={setupMutation.isPending}
			title="First-time setup"
			username={username}
		/>
	);
}

const styles = StyleSheet.create({
	server: {
		color: colors.secondaryLabel,
		fontSize: 12,
		paddingHorizontal: 8,
		textAlign: "center",
	},
});
