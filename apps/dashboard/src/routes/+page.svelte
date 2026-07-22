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

<div class="h-full w-full">
	{#if camerasQuery.isPending}
		<div class="flex h-full items-center justify-center">
			<Spinner label="Loading cameras" />
		</div>
	{:else if camerasQuery.isError}
		<div class="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
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
		<div class="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
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
</div>
