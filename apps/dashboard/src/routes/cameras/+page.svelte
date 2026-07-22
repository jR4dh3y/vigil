<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Plus, RefreshCw } from "lucide-svelte";
	import { cameraKeys, listCameras } from "$lib/cameras";
	import CameraCard from "$lib/components/cameras/CameraCard.svelte";
	import CamerasEmptyState from "$lib/components/cameras/CamerasEmptyState.svelte";
	import Spinner from "$lib/components/Spinner.svelte";

	const camerasQuery = createQuery(() => ({
		queryKey: cameraKeys.list(),
		queryFn: listCameras,
	}));

	const cameras = $derived(camerasQuery.data ?? []);
</script>

<svelte:head>
	<title>Cameras · NVR</title>
</svelte:head>

<section class="flex flex-col gap-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex flex-col gap-1">
			<h1 class="text-2xl font-semibold tracking-tight text-zinc-50">Cameras</h1>
			<p class="text-sm text-zinc-400">Manage RTSP cameras for live view and recording.</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:opacity-50"
				disabled={camerasQuery.isFetching}
				onclick={() => camerasQuery.refetch()}
			>
				<RefreshCw class="size-3.5 {camerasQuery.isFetching ? 'animate-spin' : ''}" />
				Refresh
			</button>
			<a
				href="/cameras/new"
				class="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-3.5 py-2 text-sm font-medium text-white no-underline transition-colors hover:bg-emerald-500"
			>
				<Plus class="size-4" />
				Add camera
			</a>
		</div>
	</div>

	{#if camerasQuery.isPending}
		<div class="flex min-h-[280px] items-center justify-center">
			<Spinner label="Loading cameras" />
		</div>
	{:else if camerasQuery.isError}
		<div
			class="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-12 text-center"
		>
			<p class="text-sm font-medium text-red-200">Could not load cameras</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{camerasQuery.error instanceof Error
					? camerasQuery.error.message
					: "Unknown error while loading cameras."}
			</p>
			<button
				type="button"
				class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
				onclick={() => camerasQuery.refetch()}
			>
				Retry
			</button>
		</div>
	{:else if cameras.length === 0}
		<CamerasEmptyState />
	{:else}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each cameras as camera (camera.id)}
				<CameraCard {camera} />
			{/each}
		</div>
	{/if}
</section>
