import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Redirect, router } from "expo-router";
import { useState } from "react";
import { StyleSheet, Text } from "react-native";
import { login } from "@/features/auth/api";
import { AuthScreen } from "@/features/auth/components/auth-screen";
import { authKeys } from "@/features/auth/keys";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { apiBaseUrl } from "@/lib/api/config";
import { ApiError } from "@/lib/api/error";
import { colors } from "@/theme/colors";

export default function LoginScreen() {
	const queryClient = useQueryClient();
	const auth = useAuthStatus();
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState<string | null>(null);

	const loginMutation = useMutation({
		mutationFn: () => login({ username: username.trim(), password }),
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
			setError(err instanceof Error ? err.message : "Login failed");
		},
	});

	if (auth.isSuccess && auth.data.setupRequired) {
		return <Redirect href="/setup" />;
	}
	if (auth.isSuccess && auth.data.user) {
		return <Redirect href="/(tabs)/(live)" />;
	}

	return (
		<AuthScreen
			description="Sign in with the same account you use on the web dashboard."
			error={error}
			footer={<Text style={styles.server}>Server · {apiBaseUrl}</Text>}
			onPasswordChange={setPassword}
			onSubmit={() => {
				setError(null);
				loginMutation.mutate();
			}}
			onUsernameChange={setUsername}
			password={password}
			submitLabel="Sign in"
			submitting={loginMutation.isPending}
			title="Sign in"
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
