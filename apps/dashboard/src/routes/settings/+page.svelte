<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { RefreshCw, ShieldCheck } from "lucide-svelte";
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
		putGDriveConfiguration,
		readGDriveCallback,
		StorageApiError,
		storageKeys,
		stripGDriveCallback,
		type GDriveConfigurationRequest,
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
			void queryClient.invalidateQueries({
				queryKey: storageKeys.gdriveStatus(),
			});
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

	const configureMutation = createMutation(() => ({
		mutationFn: putGDriveConfiguration,
		onSuccess: (status) => {
			driveServerError = null;
			queryClient.setQueryData(storageKeys.gdriveStatus(), status);
		},
		onError: (error: unknown) => {
			driveFlashSuccess = null;
			if (error instanceof StorageApiError) {
				driveServerError = error.message;
				return;
			}
			driveServerError =
				error instanceof Error ? error.message : "Failed to save Google Drive configuration";
		},
	}));

	const disconnectMutation = createMutation(() => ({
		mutationFn: deleteGDriveDisconnect,
		onSuccess: async () => {
			driveServerError = null;
			driveFlashSuccess = "Google Drive disconnected.";
			await queryClient.invalidateQueries({
				queryKey: storageKeys.gdriveStatus(),
			});
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

	async function handleConfigure(values: GDriveConfigurationRequest) {
		if (!isAdmin) {
			return;
		}
		driveServerError = null;
		driveFlashSuccess = null;
		await configureMutation.mutateAsync(values);
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

<section class="mx-auto flex w-full max-w-6xl flex-col gap-6 pb-8">
	<div
		class="flex flex-col gap-4 border-b border-zinc-800/80 pb-6 sm:flex-row sm:items-end sm:justify-between"
	>
		<div class="max-w-2xl">
			<p class="mb-2 text-xs font-medium tracking-[0.16em] text-emerald-400 uppercase">
				Recorder configuration
			</p>
			<h2 class="text-xl font-semibold tracking-tight text-zinc-50 sm:text-2xl">
				General settings
			</h2>
			<p class="mt-2 text-sm leading-6 text-zinc-400">
				Manage this recorder, local storage, retention, and off-site archives from one place.
			</p>
		</div>
		<div class="flex items-center gap-2 text-xs text-zinc-400">
			<span class="flex size-7 items-center justify-center rounded-full bg-zinc-900 text-zinc-400">
				<ShieldCheck class="size-3.5" />
			</span>
			<span>{isAdmin ? "Administrator access" : "View-only access"}</span>
		</div>
	</div>

	{#if systemQuery.isPending}
		<div
			class="flex min-h-[190px] items-center justify-center rounded-2xl border border-zinc-800 bg-zinc-900/30"
		>
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

	<div class="grid items-start gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(340px,0.8fr)]">
		<div class="overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900/35">
			<div class="border-b border-zinc-800 px-5 py-4 sm:px-6">
				<h2 class="text-sm font-semibold text-zinc-100">Recorder</h2>
				<p class="mt-1 text-xs leading-5 text-zinc-500">
					Identity, local recording, and retention policy.
				</p>
			</div>

			<div class="p-5 sm:p-6">
				{#if settingsQuery.isPending}
					<div class="flex min-h-[240px] items-center justify-center">
						<Spinner label="Loading settings" />
					</div>
				{:else if settingsQuery.isError}
					<div class="flex min-h-[240px] flex-col items-center justify-center gap-3 text-center">
						<p class="text-sm font-medium text-red-200">Could not load recorder settings</p>
						<p class="max-w-sm text-sm text-red-300/80">
							{settingsQuery.error instanceof Error
								? settingsQuery.error.message
								: "Failed to load settings."}
						</p>
						<button
							type="button"
							class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
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
		</div>

		<div class="flex flex-col gap-3">
			{#if driveFlashSuccess}
				<p
					class="rounded-xl border border-emerald-500/25 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300"
					role="status"
				>
					{driveFlashSuccess}
				</p>
			{/if}

			{#if gdriveQuery.isPending}
				<div
					class="flex min-h-[280px] items-center justify-center rounded-2xl border border-zinc-800 bg-zinc-900/35"
				>
					<Spinner label="Loading Google Drive status" />
				</div>
			{:else if gdriveQuery.isError}
				<div
					class="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-2xl border border-red-500/20 bg-red-500/5 px-6 py-10 text-center"
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
				<GoogleDriveCard
					status={gdriveQuery.data}
					readonly={!isAdmin}
					connecting={connectMutation.isPending}
					configuring={configureMutation.isPending}
					disconnecting={disconnectMutation.isPending}
					archiving={archiveMutation.isPending}
					serverError={driveServerError}
					onConnect={handleConnect}
					onDisconnect={handleDisconnect}
					onArchive={handleArchive}
					onConfigure={handleConfigure}
				/>
			{/if}
		</div>
	</div>
</section>
