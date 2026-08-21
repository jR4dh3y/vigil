<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import type {
		EditCameraFormValues,
		ProbeCameraRequest,
		ProbeResult,
		UpdateCameraRequest,
	} from "$lib/cameras";
	import {
		CameraApiError,
		cameraKeys,
		deleteCamera,
		formValuesFromCamera,
		getCamera,
		probeCamera,
		toUpdateCameraRequest,
		updateCamera,
	} from "$lib/cameras";
	import CameraContextBar from "$lib/components/cameras/CameraContextBar.svelte";
	import CameraSnapshot from "$lib/components/cameras/CameraSnapshot.svelte";
	import CameraSettingsModal from "$lib/components/cameras/CameraSettingsModal.svelte";
	import CameraStatusBadge from "$lib/components/cameras/CameraStatusBadge.svelte";
	import { Settings2 } from "lucide-svelte";
	import Spinner from "$lib/components/Spinner.svelte";

	const queryClient = useQueryClient();

	const cameraId = $derived(page.params.id ?? "");

	const cameraQuery = createQuery(() => ({
		queryKey: cameraKeys.detail(cameraId),
		queryFn: () => getCamera(cameraId),
		enabled: Boolean(cameraId),
	}));

	let serverError = $state<string | null>(null);
	let probeError = $state<string | null>(null);
	let probeResult = $state<ProbeResult | null>(null);
	let settingsOpen = $state(false);

	const saveMutation = createMutation(() => ({
		mutationFn: ({ id, body }: { id: string; body: UpdateCameraRequest }) =>
			updateCamera(id, body),
		onSuccess: async (camera) => {
			serverError = null;
			await queryClient.invalidateQueries({ queryKey: cameraKeys.all });
			queryClient.setQueryData(cameraKeys.detail(camera.id), camera);
			settingsOpen = false;
		},
		onError: (error: unknown) => {
			if (error instanceof CameraApiError) {
				serverError = error.message;
				return;
			}
			serverError = error instanceof Error ? error.message : "Failed to update camera";
		},
	}));

	const removeMutation = createMutation(() => ({
		mutationFn: (id: string) => deleteCamera(id),
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: cameraKeys.all });
			await goto("/cameras");
		},
		onError: (error: unknown) => {
			if (error instanceof CameraApiError) {
				serverError = error.message;
				return;
			}
			serverError = error instanceof Error ? error.message : "Failed to delete camera";
		},
	}));

	const probeMutation = createMutation(() => ({
		mutationFn: (body: ProbeCameraRequest) => probeCamera(body),
		onSuccess: (result) => {
			probeError = null;
			probeResult = result;
		},
		onError: (error: unknown) => {
			probeResult = null;
			if (error instanceof CameraApiError) {
				probeError = error.message;
				return;
			}
			probeError = error instanceof Error ? error.message : "Probe failed";
		},
	}));

	const camera = $derived(cameraQuery.data);
	const initial = $derived(camera ? formValuesFromCamera(camera) : undefined);

	async function handleSubmit(values: EditCameraFormValues) {
		if (!cameraId) {
			return;
		}
		serverError = null;
		await saveMutation.mutateAsync({
			id: cameraId,
			body: toUpdateCameraRequest(values),
		});
	}

	async function handleProbe(input: ProbeCameraRequest) {
		probeError = null;
		await probeMutation.mutateAsync(input);
	}

	async function handleDelete() {
		if (!cameraId) {
			return;
		}
		const confirmed = window.confirm(
			`Delete camera “${camera?.name ?? "this camera"}”? This cannot be undone.`,
		);
		if (!confirmed) {
			return;
		}
		serverError = null;
		await removeMutation.mutateAsync(cameraId);
	}
</script>

<svelte:head>
	<title>{camera?.name ?? "Camera"} · Vigil</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-5xl flex-col gap-6">
	{#if cameraQuery.isPending}
		<div class="flex min-h-[280px] items-center justify-center">
			<Spinner label="Loading camera" />
		</div>
	{:else if cameraQuery.isError}
		<div
			class="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-12 text-center"
		>
			<p class="text-sm font-medium text-red-200">Could not load camera</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{cameraQuery.error instanceof Error
					? cameraQuery.error.message
					: "Unknown error while loading camera."}
			</p>
			<div class="flex gap-2">
				<a
					href="/cameras"
					class="rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm text-zinc-200 no-underline hover:bg-zinc-800"
				>
					Back to list
				</a>
				<button
					type="button"
					class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
					onclick={() => cameraQuery.refetch()}
				>
					Retry
				</button>
			</div>
		</div>
	{:else if camera && initial}
		<CameraContextBar {camera}>
			{#snippet actions()}
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100"
				onclick={() => {
					serverError = null;
					probeResult = null;
					probeError = null;
					settingsOpen = true;
				}}
				aria-label="Open camera settings"
			>
				<Settings2 class="size-3.5" />
				<span class="hidden sm:inline">Settings</span>
			</button>
			<a
				href="/cameras/{camera.id}/timeline"
				class="inline-flex items-center rounded-md border border-emerald-500/25 bg-emerald-500/10 px-2.5 py-1.5 text-sm text-emerald-300 no-underline transition-colors hover:border-emerald-500/40 hover:bg-emerald-500/15"
			>
				Timeline
			</a>
			{/snippet}
		</CameraContextBar>

		<div class="flex flex-wrap items-center gap-2 text-xs">
			<CameraStatusBadge status={camera.status} />
			<span
				class="rounded-md border px-2 py-0.5
					{camera.enabled
					? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
					: 'border-zinc-700 bg-zinc-800 text-zinc-500'}"
			>
				{camera.enabled ? "Enabled" : "Disabled"}
			</span>
			<span class="font-mono text-zinc-500">{camera.host}</span>
		</div>

		<CameraSnapshot cameraId={camera.id} cameraName={camera.name} />

		<CameraSettingsModal
			open={settingsOpen}
			cameraName={camera.name}
			{initial}
			submitting={saveMutation.isPending}
			probing={probeMutation.isPending}
			deleting={removeMutation.isPending}
			{serverError}
			{probeResult}
			{probeError}
			onClose={() => (settingsOpen = false)}
			onSubmit={handleSubmit}
			onProbe={handleProbe}
			onDelete={handleDelete}
		/>
	{/if}
</section>
