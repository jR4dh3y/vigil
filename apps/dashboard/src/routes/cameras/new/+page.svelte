<script lang="ts">
	import { goto } from "$app/navigation";
	import { createMutation, useQueryClient } from "@tanstack/svelte-query";
	import type { CreateCameraRequest, ProbeCameraRequest, ProbeResult } from "$lib/cameras";
	import {
		CameraApiError,
		cameraKeys,
		createCamera,
		type CreateCameraFormValues,
		probeCamera,
		toCreateCameraRequest,
	} from "$lib/cameras";
	import CameraForm from "$lib/components/cameras/CameraForm.svelte";

	const queryClient = useQueryClient();

	let serverError = $state<string | null>(null);
	let probeError = $state<string | null>(null);
	let probeResult = $state<ProbeResult | null>(null);

	const createCameraMutation = createMutation(() => ({
		mutationFn: (body: CreateCameraRequest) => createCamera(body),
		onSuccess: async (camera) => {
			serverError = null;
			await queryClient.invalidateQueries({ queryKey: cameraKeys.all });
			await goto(`/cameras/${camera.id}`);
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

	async function handleSubmit(values: CreateCameraFormValues) {
		serverError = null;
		await createCameraMutation.mutateAsync(toCreateCameraRequest(values));
	}

	async function handleProbe(input: ProbeCameraRequest) {
		probeError = null;
		await probeMutation.mutateAsync(input);
	}
</script>

<svelte:head>
	<title>Add camera · NVR</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-2xl flex-col gap-6">
	<div class="flex flex-col gap-1">
		<p class="text-xs font-medium tracking-wide text-zinc-500 uppercase">
			<a href="/cameras" class="text-zinc-500 no-underline hover:text-zinc-300">Cameras</a>
			<span class="mx-1.5 text-zinc-700">/</span>
			<span class="text-zinc-400">New</span>
		</p>
		<h1 class="text-2xl font-semibold tracking-tight text-zinc-50">Add camera</h1>
		<p class="text-sm text-zinc-400">
			Configure RTSP endpoints and optionally probe the live stream before saving.
		</p>
	</div>

	<div class="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5 sm:p-6">
		<CameraForm
			mode="create"
			submitting={createCameraMutation.isPending}
			probing={probeMutation.isPending}
			{serverError}
			{probeResult}
			{probeError}
			onSubmit={handleSubmit}
			onProbe={handleProbe}
		/>
	</div>
</section>
