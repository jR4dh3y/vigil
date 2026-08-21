<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import {
		AuthApiError,
		authKeys,
		login as loginRequest,
		type LoginFormValues,
	} from "$lib/auth";
	import AuthCard from "$lib/components/AuthCard.svelte";
	import RemoteServerBadge from "$lib/components/RemoteServerBadge.svelte";
	import LoginForm from "$lib/components/forms/LoginForm.svelte";

	const queryClient = useQueryClient();

	let serverError = $state<string | null>(null);

	const loginMutation = createMutation(() => ({
		mutationFn: (values: LoginFormValues) => loginRequest(values),
		onSuccess: async () => {
			serverError = null;
			await queryClient.invalidateQueries({ queryKey: authKeys.all });
			await goto(resolve("/"));
		},
		onError: (error: unknown) => {
			if (error instanceof AuthApiError) {
				serverError = error.message;
				return;
			}
			serverError = error instanceof Error ? error.message : "Login failed";
		},
	}));

	async function handleSubmit(values: LoginFormValues) {
		serverError = null;
		await loginMutation.mutateAsync(values);
	}
</script>

<svelte:head>
	<title>Sign in · Vigil</title>
</svelte:head>

<AuthCard title="Sign in" description="Access your NVR dashboard with your account credentials.">
	<RemoteServerBadge />
	<LoginForm
		submitting={loginMutation.isPending}
		serverError={serverError}
		onSubmit={handleSubmit}
	/>
</AuthCard>
