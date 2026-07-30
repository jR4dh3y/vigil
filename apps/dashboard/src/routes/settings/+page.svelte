<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { RefreshCw } from "lucide-svelte";
	import { onMount } from "svelte";
	import { authKeys, loadStatus } from "$lib/auth";
	import PageActions from "$lib/components/PageActions.svelte";
	import GoogleDriveCard from "$lib/components/settings/GoogleDriveCard.svelte";
	import SettingsForm from "$lib/components/settings/SettingsForm.svelte";
	import SystemStatusCard from "$lib/components/settings/SystemStatusCard.svelte";
	import Spinner from "$lib/components/Spinner.svelte";
	import {
		deleteGDriveDisconnect,
		formatGDriveArchiveResult,
		getGDriveStatus,
		postGDriveArchive,
		postGDriveConnect,
		readGDriveCallback,
		StorageApiError,
		storageKeys,
		stripGDriveCallback,
	} from "$lib/storage";
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

	const gdriveQuery = createQuery(() => ({
		queryKey: storageKeys.gdriveStatus(),
		queryFn: getGDriveStatus,
	}));

	const isAdmin = $derived(statusQuery.data?.user?.role === "admin");

	let serverError = $state<string | null>(null);
	let successMessage = $state<string | null>(null);
	let driveServerError = $state<string | null>(null);
	let driveFlashSuccess = $state<string | null>(null);

	// OAuth performs a full-page redirect, so consume its one-time notice once
	// when this settings route mounts.
	onMount(() => {
		const notice = readGDriveCallback(page.url.searchParams);
		if (!notice) {
			return;
		}
		if (notice.kind === "connected") {
			driveFlashSuccess = notice.message;
			driveServerError = null;
			void queryClient.invalidateQueries({ queryKey: storageKeys.gdriveStatus() });
		} else {
			driveServerError = notice.message;
			driveFlashSuccess = null;
		}
		void goto(stripGDriveCallback(page.url), {
			replaceState: true,
			noScroll: true,
			keepFocus: true,
		});
	});

	const saveMutation = createMutation(() => ({
		mutationFn: (values: SettingsFormValues) =>
			patchSettings({
				siteName: values.siteName,
				retentionDays: values.retentionDays,
				recordingsDir: values.recordingsDir,
				recordingEnabled: values.recordingEnabled,
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

	const connectMutation = createMutation(() => ({
		mutationFn: postGDriveConnect,
		onSuccess: (result) => {
			driveServerError = null;
			driveFlashSuccess = null;
			window.location.assign(result.authorizationUrl);
		},
		onError: (error: unknown) => {
			driveFlashSuccess = null;
			if (error instanceof StorageApiError) {
				driveServerError = error.message;
				return;
			}
			driveServerError =
				error instanceof Error ? error.message : "Failed to start Google Drive connection";
		},
	}));

	const disconnectMutation = createMutation(() => ({
		mutationFn: deleteGDriveDisconnect,
		onSuccess: async () => {
			driveServerError = null;
			driveFlashSuccess = "Google Drive disconnected.";
			await queryClient.invalidateQueries({ queryKey: storageKeys.gdriveStatus() });
		},
		onError: (error: unknown) => {
			driveFlashSuccess = null;
			if (error instanceof StorageApiError) {
				driveServerError = error.message;
				return;
			}
			driveServerError =
				error instanceof Error ? error.message : "Failed to disconnect Google Drive";
		},
	}));

	const archiveMutation = createMutation(() => ({
		mutationFn: () => postGDriveArchive(),
		onSuccess: (result) => {
			driveServerError = null;
			driveFlashSuccess = formatGDriveArchiveResult(result);
		},
		onError: (error: unknown) => {
			driveFlashSuccess = null;
			if (error instanceof StorageApiError) {
				driveServerError = error.message;
				return;
			}
			driveServerError =
				error instanceof Error ? error.message : "Failed to run Google Drive archive";
		},
	}));

	const formInitial = $derived(
		settingsQuery.data
			? {
					siteName: settingsQuery.data.siteName,
					retentionDays: settingsQuery.data.retentionDays,
					recordingsDir: settingsQuery.data.recordingsDir,
					recordingEnabled: settingsQuery.data.recordingEnabled,
				}
			: null,
	);

	const isRefreshing = $derived(
		systemQuery.isFetching || settingsQuery.isFetching || gdriveQuery.isFetching,
	);

	async function handleSave(values: SettingsFormValues) {
		serverError = null;
		successMessage = null;
		await saveMutation.mutateAsync(values);
	}

	async function handleConnect() {
		if (!isAdmin) {
			return;
		}
		driveServerError = null;
		driveFlashSuccess = null;
		await connectMutation.mutateAsync();
	}

	async function handleDisconnect() {
		if (!isAdmin) {
			return;
		}
		const confirmed = window.confirm(
			"Disconnect Google Drive? The NVR will stop using this account until you connect again.",
		);
		if (!confirmed) {
			return;
		}
		driveServerError = null;
		driveFlashSuccess = null;
		await disconnectMutation.mutateAsync();
	}

	async function handleArchive() {
		if (!isAdmin) {
			return;
		}
		driveServerError = null;
		driveFlashSuccess = null;
		await archiveMutation.mutateAsync();
	}

	function refreshAll() {
		void systemQuery.refetch();
		void settingsQuery.refetch();
		void gdriveQuery.refetch();
	}
</script>

<svelte:head>
	<title>Settings · NVR</title>
</svelte:head>

<PageActions>
	<button
		type="button"
		class="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:opacity-50"
		disabled={isRefreshing}
		onclick={refreshAll}
	>
		<RefreshCw class="size-3.5 {isRefreshing ? 'animate-spin' : ''}" />
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
					Site name, recording location on this server, and retention.
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
			{#key `${formInitial.siteName}:${formInitial.retentionDays}:${formInitial.recordingsDir}:${formInitial.recordingEnabled}`}
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

	{#if gdriveQuery.isPending}
		<div class="flex min-h-[120px] items-center justify-center rounded-xl border border-zinc-800 bg-zinc-900/50">
			<Spinner label="Loading Google Drive status" />
		</div>
	{:else if gdriveQuery.isError}
		<div
			class="flex flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-10 text-center"
		>
			<p class="text-sm font-medium text-red-200">Could not load Google Drive status</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{gdriveQuery.error instanceof Error
					? gdriveQuery.error.message
					: "Unknown error while loading Drive status."}
			</p>
			<button
				type="button"
				class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
				onclick={() => gdriveQuery.refetch()}
			>
				Retry
			</button>
		</div>
	{:else if gdriveQuery.data}
		{#if driveFlashSuccess}
			<p
				class="rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-300"
				role="status"
			>
				{driveFlashSuccess}
			</p>
		{/if}
		<GoogleDriveCard
			status={gdriveQuery.data}
			readonly={!isAdmin}
			connecting={connectMutation.isPending}
			disconnecting={disconnectMutation.isPending}
			archiving={archiveMutation.isPending}
			serverError={driveServerError}
			onConnect={handleConnect}
			onDisconnect={handleDisconnect}
			onArchive={handleArchive}
		/>
	{/if}
</section>
