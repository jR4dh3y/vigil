import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Redirect, router } from "expo-router";
import { useState } from "react";
import { login } from "@/features/auth/api";
import { AuthScreen } from "@/features/auth/components/auth-screen";
import { authKeys } from "@/features/auth/keys";
import { useAuthStatus } from "@/features/auth/use-auth-status";
import { takePendingEventRoute } from "@/features/notifications/pending-event-route";
import { RecorderLink } from "@/features/server/components/recorder-link";
import { useApiConfiguration } from "@/features/server/use-api-base-url";
import { ApiError } from "@/lib/api/error";

export default function LoginScreen() {
	const queryClient = useQueryClient();
	const configuration = useApiConfiguration();
	const configured = configuration.kind === "configured";
	const auth = useAuthStatus(configured);
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState<string | null>(null);

	const loginMutation = useMutation({
		mutationFn: () => login({ username: username.trim(), password }),
		onSuccess: async () => {
			setError(null);
			await queryClient.invalidateQueries({ queryKey: authKeys.all });
			const eventId = takePendingEventRoute();
			router.replace(
				eventId ? { pathname: "/event/[id]", params: { id: eventId } } : "/(tabs)/(live)",
			);
		},
		onError: (err: unknown) => {
			if (err instanceof ApiError) {
				setError(err.message);
				return;
			}
			setError(err instanceof Error ? err.message : "Login failed");
		},
	});

	if (!configured) {
		return <Redirect href="/server" />;
	}

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
			footer={<RecorderLink />}
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
