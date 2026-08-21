<script lang="ts">
	import { resolve } from "$app/paths";
	import { goto } from "$app/navigation";
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import type {
		CreateCameraRequest,
		DiscoverCameraStreamsRequest,
		DiscoverCameraStreamsResult,
		DiscoverCamerasRequest,
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
	let hasScanned = $state(false);
	let selectedCamera = $state<DiscoveredCamera | null>(null);
	let streamResult = $state<DiscoverCameraStreamsResult | null>(null);
	let credentialUsername = $state("");
	let credentialPassword = $state("");
	let configureCamera = $state(false);
	let credentialsReady = $state(false);
	let bulkAdding = $state(false);
	let bulkAddError = $state<string | null>(null);
	let bulkAddProgress = $state<{ completed: number; total: number } | null>(null);

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
		mutationFn: (body: DiscoverCamerasRequest) => discoverCameras(body),
	}));

	const streamDiscoveryMutation = createMutation(() => ({
		mutationFn: (body: DiscoverCameraStreamsRequest) => discoverCameraStreams(body),
	}));

	async function handleDiscover(username: string, password: string) {
		credentialUsername = username;
		credentialPassword = password;
		discoveryError = null;
		bulkAddError = null;
		try {
			discoveredCameras = await discoveryMutation.mutateAsync({ username, password });
		} catch (error: unknown) {
			discoveredCameras = [];
			discoveryError =
				error instanceof CameraApiError
					? error.message
					: error instanceof Error
						? error.message
						: "Camera discovery failed";
		} finally {
			hasScanned = true;
		}
	}

	async function handleAddSelected(cameras: DiscoveredCamera[]) {
		if (bulkAdding || cameras.length === 0) {
			return;
		}
		if (!credentialUsername.trim() || !credentialPassword) {
			const message = "Enter the NVR username and password before adding cameras.";
			bulkAddError = message;
			throw new Error(message);
		}

		bulkAdding = true;
		bulkAddError = null;
		bulkAddProgress = { completed: 0, total: cameras.length };
		let added = 0;
		const failures: string[] = [];

		try {
			for (const [index, camera] of cameras.entries()) {
				try {
					const streams = camera.liveRtspUrl
						? {
								liveRtspUrl: camera.liveRtspUrl,
								recordRtspUrl: camera.recordRtspUrl ?? camera.liveRtspUrl,
							}
						: await discoverCameraStreams({
								xaddr: camera.xaddr,
								username: credentialUsername,
								password: credentialPassword,
							});

					await createCamera({
						name: camera.name,
						host: camera.host,
						driver: "generic-rtsp",
						enabled: true,
						username: credentialUsername,
						password: credentialPassword,
						liveRtspUrl: streams.liveRtspUrl,
						recordRtspUrl: streams.recordRtspUrl,
					});
					added += 1;
				} catch (error: unknown) {
					const message =
						error instanceof CameraApiError
							? error.message
							: error instanceof Error
								? error.message
								: "failed to add camera";
					failures.push(`${camera.name}: ${message}`);
				}
				bulkAddProgress = { completed: index + 1, total: cameras.length };
			}
		} finally {
			bulkAdding = false;
		}

		if (failures.length > 0) {
			const message = `Added ${added} of ${cameras.length} cameras. ${failures.slice(0, 3).join("; ")}`;
			bulkAddError = message;
			throw new Error(message);
		}

		await queryClient.invalidateQueries({ queryKey: cameraKeys.all });
		await goto(resolve("/cameras"));
	}

	async function handleSelect(camera: DiscoveredCamera) {
		selectedCamera = camera;
		configureCamera = true;
		credentialsReady = false;
		streamResult = null;
		streamDiscoveryError = null;
		serverError = null;
		probeError = null;
		probeResult = null;
		if (camera.liveRtspUrl) {
			streamResult = {
				liveRtspUrl: camera.liveRtspUrl,
				recordRtspUrl: camera.recordRtspUrl ?? camera.liveRtspUrl,
			};
			credentialsReady = true;
			return;
		}
		await handleDiscoverStreams(credentialUsername, credentialPassword);
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
		bulkAddError = null;
		bulkAddProgress = null;
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
	<title>Add camera · Vigil</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-2xl flex-col gap-6">
	<div class="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5 sm:p-6">
		{#if !configureCamera}
			<CameraDiscoveryPanel
				cameras={discoveredCameras}
				scanning={discoveryMutation.isPending}
				{hasScanned}
				error={discoveryError ?? bulkAddError}
				onScan={handleDiscover}
				onSelect={handleSelect}
				onAddSelected={handleAddSelected}
				onManual={handleManual}
				addingSelected={bulkAdding}
				addProgress={bulkAddProgress}
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
					initialUsername={credentialUsername}
					initialPassword={credentialPassword}
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
