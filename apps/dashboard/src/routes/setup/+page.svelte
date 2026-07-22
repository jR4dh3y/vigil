<script lang="ts">
	import { goto } from "$app/navigation";
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import {
		AuthApiError,
		authKeys,
		setup as setupRequest,
		type SetupFormValues,
	} from "$lib/auth";
	import AuthCard from "$lib/components/AuthCard.svelte";
	import SetupForm from "$lib/components/forms/SetupForm.svelte";

	const queryClient = useQueryClient();

	let serverError = $state<string | null>(null);

	const setupMutation = createMutation(() => ({
		mutationFn: (values: SetupFormValues) => setupRequest(values),
		onSuccess: async () => {
			serverError = null;
			await queryClient.invalidateQueries({ queryKey: authKeys.all });
			await goto("/");
		},
		onError: (error: unknown) => {
			if (error instanceof AuthApiError) {
				serverError = error.message;
				return;
			}
			serverError = error instanceof Error ? error.message : "Setup failed";
		},
	}));

	async function handleSubmit(values: SetupFormValues) {
		serverError = null;
		await setupMutation.mutateAsync(values);
	}
</script>

<svelte:head>
	<title>Initial setup · NVR</title>
</svelte:head>

<AuthCard
	title="Initial setup"
	description="Create the first admin account for this NVR instance. This is only available when no users exist yet."
>
	<SetupForm
		submitting={setupMutation.isPending}
		serverError={serverError}
		onSubmit={handleSubmit}
	/>
</AuthCard>
