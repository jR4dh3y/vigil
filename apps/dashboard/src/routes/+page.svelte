<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { LayoutGrid, Video } from "lucide-svelte";
	import { cameraKeys, listCameras } from "$lib/cameras";
	import LiveGrid from "$lib/components/live/LiveGrid.svelte";
	import Spinner from "$lib/components/Spinner.svelte";

	const camerasQuery = createQuery(() => ({
		queryKey: cameraKeys.list(),
		queryFn: listCameras,
	}));

	const enabledCameras = $derived(
		(camerasQuery.data ?? [])
			.filter((c) => c.enabled)
			.map((c) => ({ id: c.id, name: c.name })),
	);
</script>

<svelte:head>
	<title>Live · NVR</title>
</svelte:head>

<section class="flex flex-col gap-6">
	<div class="flex flex-col gap-1">
		<h1 class="text-2xl font-semibold tracking-tight text-zinc-50">Live</h1>
		<p class="text-sm text-zinc-400">
			{#if camerasQuery.isSuccess}
				{enabledCameras.length === 0
					? "Enable a camera to start live viewing."
					: `${enabledCameras.length} camera${enabledCameras.length === 1 ? "" : "s"} live`}
			{:else}
				Real-time camera mosaic (WebRTC with HLS fallback).
			{/if}
		</p>
	</div>

	{#if camerasQuery.isPending}
		<div class="flex min-h-[420px] items-center justify-center">
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
	{:else if enabledCameras.length === 0}
		<div
			class="flex min-h-[420px] flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-zinc-800 bg-zinc-900/40 px-6 py-16 text-center"
		>
			<span
				class="flex size-12 items-center justify-center rounded-xl bg-zinc-800/80 text-zinc-400"
			>
				{#if (camerasQuery.data ?? []).length === 0}
					<LayoutGrid class="size-6" />
				{:else}
					<Video class="size-6" />
				{/if}
			</span>
			{#if (camerasQuery.data ?? []).length === 0}
				<p class="text-sm font-medium text-zinc-300">No cameras yet</p>
				<p class="max-w-sm text-sm text-zinc-500">
					Add cameras from the
					<a
						href="/cameras"
						class="text-emerald-400 no-underline hover:text-emerald-300">Cameras</a
					>
					page, then enable them for live view.
				</p>
			{:else}
				<p class="text-sm font-medium text-zinc-300">No enabled cameras</p>
				<p class="max-w-sm text-sm text-zinc-500">
					You have cameras configured, but none are enabled. Turn one on from
					<a
						href="/cameras"
						class="text-emerald-400 no-underline hover:text-emerald-300">Cameras</a
					>
					to see the live grid.
				</p>
			{/if}
		</div>
	{:else}
		<LiveGrid cameras={enabledCameras} />
	{/if}
</section>
