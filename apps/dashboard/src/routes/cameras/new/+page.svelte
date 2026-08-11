<script lang="ts">
	import { onMount } from "svelte";
	import { resolve } from "$app/paths";
	import { goto } from "$app/navigation";
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import type {
		CreateCameraRequest,
		DiscoverCameraStreamsRequest,
		DiscoverCameraStreamsResult,
		DiscoveredCamera,
		ProbeCameraRequest,
		ProbeResult,
	} from "$lib/cameras";
	import {
		CameraApiError,
		cameraKeys,
		createCamera,
		discoverCameraStreams,
		discoverCameras,
		type CreateCameraFormValues,
		probeCamera,
		toCreateCameraRequest,
	} from "$lib/cameras";
	import CameraCredentialsPanel from "$lib/components/cameras/CameraCredentialsPanel.svelte";
	import CameraDiscoveryPanel from "$lib/components/cameras/CameraDiscoveryPanel.svelte";
	import CameraForm from "$lib/components/cameras/CameraForm.svelte";

	const queryClient = useQueryClient();

	let serverError = $state<string | null>(null);
	let probeError = $state<string | null>(null);
	let probeResult = $state<ProbeResult | null>(null);
	let discoveryError = $state<string | null>(null);
	let streamDiscoveryError = $state<string | null>(null);
	let discoveredCameras = $state<DiscoveredCamera[]>([]);
	let selectedCamera = $state<DiscoveredCamera | null>(null);
	let streamResult = $state<DiscoverCameraStreamsResult | null>(null);
	let credentialUsername = $state("");
	let credentialPassword = $state("");
	let configureCamera = $state(false);
	let credentialsReady = $state(false);

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

	const streamDiscoveryMutation = createMutation(() => ({
		mutationFn: (body: DiscoverCameraStreamsRequest) => discoverCameraStreams(body),
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
		credentialsReady = false;
		streamResult = null;
		streamDiscoveryError = null;
		credentialUsername = "";
		credentialPassword = "";
		serverError = null;
		probeError = null;
		probeResult = null;
	}

	function handleManual() {
		selectedCamera = null;
		configureCamera = true;
		credentialsReady = true;
		streamResult = null;
		credentialUsername = "";
		credentialPassword = "";
		serverError = null;
		probeError = null;
		probeResult = null;
	}

	async function handleDiscoverStreams(username: string, password: string) {
		if (!selectedCamera) {
			return;
		}
		credentialUsername = username;
		credentialPassword = password;
		streamDiscoveryError = null;
		try {
			streamResult = await streamDiscoveryMutation.mutateAsync({
				xaddr: selectedCamera.xaddr,
				username,
				password,
			});
			credentialsReady = true;
		} catch (error: unknown) {
			streamResult = null;
			streamDiscoveryError =
				error instanceof CameraApiError
					? error.message
					: error instanceof Error
						? error.message
						: "RTSP stream discovery failed";
		}
	}

	function handleContinueWithoutDiscovery(username: string, password: string) {
		credentialUsername = username;
		credentialPassword = password;
		streamResult = null;
		streamDiscoveryError = null;
		credentialsReady = true;
	}

	function handleBackToDiscovery() {
		configureCamera = false;
		credentialsReady = false;
		selectedCamera = null;
		streamResult = null;
		credentialUsername = "";
		credentialPassword = "";
		streamDiscoveryError = null;
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

			{#if selectedCamera && !credentialsReady}
				<CameraCredentialsPanel
					camera={selectedCamera}
					detecting={streamDiscoveryMutation.isPending}
					error={streamDiscoveryError}
					onDetect={handleDiscoverStreams}
					onContinue={handleContinueWithoutDiscovery}
				/>
			{:else}
				<CameraForm
					mode="create"
					initial={
						selectedCamera
							? {
									name: selectedCamera.name,
									host: selectedCamera.host,
									username: credentialUsername,
									password: credentialPassword,
									liveRtspUrl: streamResult?.liveRtspUrl ?? "",
									recordRtspUrl: streamResult?.recordRtspUrl ?? "",
								}
							: {}
					}
					submitting={createCameraMutation.isPending}
					probing={probeMutation.isPending}
					{serverError}
					{probeResult}
					{probeError}
					onSubmit={handleSubmit}
					onProbe={handleProbe}
				/>
			{/if}
		{/if}
	</div>
</section>
