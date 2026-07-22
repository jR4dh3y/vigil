<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { RefreshCw } from "lucide-svelte";
	import { authKeys, loadStatus } from "$lib/auth";
	import PageActions from "$lib/components/PageActions.svelte";
	import SettingsForm from "$lib/components/settings/SettingsForm.svelte";
	import SystemStatusCard from "$lib/components/settings/SystemStatusCard.svelte";
	import Spinner from "$lib/components/Spinner.svelte";
	import {
		getSettings,
		getSystemStatus,
		patchSettings,
		type SettingsFormValues,
		SystemApiError,
		systemKeys,
	} from "$lib/system";

	const queryClient = useQueryClient();

	const statusQuery = createQuery(() => ({
		queryKey: authKeys.status,
		queryFn: loadStatus,
	}));

	const systemQuery = createQuery(() => ({
		queryKey: systemKeys.status(),
		queryFn: getSystemStatus,
	}));

	const settingsQuery = createQuery(() => ({
		queryKey: systemKeys.settings(),
		queryFn: getSettings,
	}));

	const isAdmin = $derived(statusQuery.data?.user?.role === "admin");

	let serverError = $state<string | null>(null);
	let successMessage = $state<string | null>(null);

	const saveMutation = createMutation(() => ({
		mutationFn: (values: SettingsFormValues) =>
			patchSettings({
				siteName: values.siteName,
				retentionDays: values.retentionDays,
			}),
		onSuccess: async (settings) => {
			serverError = null;
			successMessage = "Settings saved.";
			queryClient.setQueryData(systemKeys.settings(), settings);
			await queryClient.invalidateQueries({ queryKey: systemKeys.status() });
		},
		onError: (error: unknown) => {
			successMessage = null;
			if (error instanceof SystemApiError) {
				serverError = error.message;
				return;
			}
			serverError = error instanceof Error ? error.message : "Failed to save settings";
		},
	}));

	const formInitial = $derived(
		settingsQuery.data
			? {
					siteName: settingsQuery.data.siteName,
					retentionDays: settingsQuery.data.retentionDays,
				}
			: null,
	);

	async function handleSave(values: SettingsFormValues) {
		serverError = null;
		successMessage = null;
		await saveMutation.mutateAsync(values);
	}

	function refreshAll() {
		void systemQuery.refetch();
		void settingsQuery.refetch();
	}
</script>

<svelte:head>
	<title>Settings · NVR</title>
</svelte:head>

<PageActions>
	<button
		type="button"
		class="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:opacity-50"
		disabled={systemQuery.isFetching || settingsQuery.isFetching}
		onclick={refreshAll}
	>
		<RefreshCw
			class="size-3.5 {systemQuery.isFetching || settingsQuery.isFetching
				? 'animate-spin'
				: ''}"
		/>
		<span class="hidden sm:inline">Refresh</span>
	</button>
</PageActions>

<section class="mx-auto flex w-full max-w-3xl flex-col gap-6">
	{#if systemQuery.isPending}
		<div class="flex min-h-[160px] items-center justify-center">
			<Spinner label="Loading system status" />
		</div>
	{:else if systemQuery.isError}
		<div
			class="flex flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-10 text-center"
		>
			<p class="text-sm font-medium text-red-200">Could not load system status</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{systemQuery.error instanceof Error
					? systemQuery.error.message
					: "Unknown error while loading status."}
			</p>
			<button
				type="button"
				class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
				onclick={() => systemQuery.refetch()}
			>
				Retry
			</button>
		</div>
	{:else if systemQuery.data}
		<SystemStatusCard status={systemQuery.data} />
	{/if}

	<div class="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5 sm:p-6">
		<div class="mb-5 flex flex-col gap-1">
			<h2 class="text-sm font-semibold text-zinc-100">Site settings</h2>
			<p class="text-xs text-zinc-500">
				{#if isAdmin}
					Update the site name and default recording retention.
				{:else}
					View-only. Ask an administrator to change these values.
				{/if}
			</p>
		</div>

		{#if settingsQuery.isPending}
			<div class="flex min-h-[120px] items-center justify-center">
				<Spinner label="Loading settings" />
			</div>
		{:else if settingsQuery.isError}
			<div class="flex flex-col gap-2">
				<p class="text-sm text-red-300">
					{settingsQuery.error instanceof Error
						? settingsQuery.error.message
						: "Failed to load settings."}
				</p>
				<button
					type="button"
					class="w-fit rounded-lg bg-zinc-800 px-3 py-1.5 text-sm text-zinc-100 hover:bg-zinc-700"
					onclick={() => settingsQuery.refetch()}
				>
					Retry
				</button>
			</div>
		{:else if formInitial}
			{#key `${formInitial.siteName}:${formInitial.retentionDays}`}
				<SettingsForm
					initial={formInitial}
					readonly={!isAdmin}
					submitting={saveMutation.isPending}
					{serverError}
					{successMessage}
					onSubmit={handleSave}
				/>
			{/key}
		{/if}
	</div>
</section>
