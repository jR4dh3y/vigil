<script lang="ts">
	import { onMount } from "svelte";
	import { resolve } from "$app/paths";
	import { goto } from "$app/navigation";
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import type {
		CreateCameraRequest,
		DiscoveredCamera,
		ProbeCameraRequest,
		ProbeResult,
	} from "$lib/cameras";
	import {
		CameraApiError,
		cameraKeys,
		createCamera,
		discoverCameras,
		type CreateCameraFormValues,
		probeCamera,
		toCreateCameraRequest,
	} from "$lib/cameras";
	import CameraDiscoveryPanel from "$lib/components/cameras/CameraDiscoveryPanel.svelte";
	import CameraForm from "$lib/components/cameras/CameraForm.svelte";

	const queryClient = useQueryClient();

	let serverError = $state<string | null>(null);
	let probeError = $state<string | null>(null);
	let probeResult = $state<ProbeResult | null>(null);
	let discoveryError = $state<string | null>(null);
	let discoveredCameras = $state<DiscoveredCamera[]>([]);
	let selectedCamera = $state<DiscoveredCamera | null>(null);
	let configureCamera = $state(false);

	const createCameraMutation = createMutation(() => ({
		mutationFn: (body: CreateCameraRequest) => createCamera(body),
		onSuccess: async (camera) => {
			serverError = null;
			await queryClient.invalidateQueries({ queryKey: cameraKeys.all });
			await goto(resolve("/cameras/[id]", { id: camera.id }));
		},
		onError: (error: unknown) => {
			if (error instanceof CameraApiError) {
				serverError = error.message;
				return;
			}
			serverError = error instanceof Error ? error.message : "Failed to create camera";
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

	const discoveryMutation = createMutation(() => ({
		mutationFn: discoverCameras,
	}));

	onMount(() => {
		void handleDiscover();
	});

	async function handleDiscover() {
		discoveryError = null;
		try {
			discoveredCameras = await discoveryMutation.mutateAsync();
		} catch (error: unknown) {
			discoveredCameras = [];
			discoveryError =
				error instanceof CameraApiError
					? error.message
					: error instanceof Error
						? error.message
						: "Camera discovery failed";
		}
	}

	function handleSelect(camera: DiscoveredCamera) {
		selectedCamera = camera;
		configureCamera = true;
		serverError = null;
		probeError = null;
		probeResult = null;
	}

	function handleManual() {
		selectedCamera = null;
		configureCamera = true;
		serverError = null;
		probeError = null;
		probeResult = null;
	}

	function handleBackToDiscovery() {
		configureCamera = false;
		selectedCamera = null;
		serverError = null;
		probeError = null;
		probeResult = null;
	}

	function handleSubmit(values: CreateCameraFormValues) {
		serverError = null;
		createCameraMutation.mutate(toCreateCameraRequest(values));
	}

	function handleProbe(input: ProbeCameraRequest) {
		probeError = null;
		probeMutation.mutate(input);
	}
</script>

<svelte:head>
	<title>Add camera · NVR</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-2xl flex-col gap-6">
	<div class="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5 sm:p-6">
		{#if !configureCamera}
			<CameraDiscoveryPanel
				cameras={discoveredCameras}
				scanning={discoveryMutation.isPending}
				error={discoveryError}
				onScan={handleDiscover}
				onSelect={handleSelect}
				onManual={handleManual}
			/>
		{:else}
			<div class="mb-5 flex flex-wrap items-start justify-between gap-3 border-b border-zinc-800 pb-4">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-400">
						{selectedCamera ? "Camera discovered" : "Manual setup"}
					</p>
					<h1 class="mt-1 text-xl font-semibold text-zinc-100">
						{selectedCamera ? selectedCamera.name : "Add a camera"}
					</h1>
					{#if selectedCamera}
						<p class="mt-1 font-mono text-xs text-zinc-500">{selectedCamera.host}</p>
					{:else}
						<p class="mt-1 text-sm text-zinc-400">Enter the camera network details below.</p>
					{/if}
				</div>
				<button
					type="button"
					disabled={createCameraMutation.isPending || probeMutation.isPending}
					onclick={handleBackToDiscovery}
					class="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-600 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-50"
				>
					Back to discovery
				</button>
			</div>

			<CameraForm
				mode="create"
				initial={selectedCamera ? { name: selectedCamera.name, host: selectedCamera.host } : {}}
				submitting={createCameraMutation.isPending}
				probing={probeMutation.isPending}
				{serverError}
				{probeResult}
				{probeError}
				onSubmit={handleSubmit}
				onProbe={handleProbe}
			/>
		{/if}
	</div>
</section>
