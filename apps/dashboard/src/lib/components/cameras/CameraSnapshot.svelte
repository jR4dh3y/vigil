<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { ImageOff, RefreshCw } from "lucide-svelte";
	import { cameraKeys, getCameraSnapshot } from "$lib/cameras";
	import Spinner from "$lib/components/Spinner.svelte";

	type Props = {
		cameraId: string;
		cameraName?: string;
	};

	let { cameraId, cameraName }: Props = $props();

	const snapshotQuery = createQuery(() => ({
		queryKey: cameraKeys.snapshot(cameraId),
		queryFn: () => getCameraSnapshot(cameraId),
		enabled: Boolean(cameraId),
		staleTime: 10_000,
		retry: false,
	}));

	let objectUrl = $state<string | null>(null);

	$effect(() => {
		const blob = snapshotQuery.data;
		if (!blob) {
			return;
		}
		const url = URL.createObjectURL(blob);
		objectUrl = url;
		return () => {
			URL.revokeObjectURL(url);
			if (objectUrl === url) {
				objectUrl = null;
			}
		};
	});
</script>

<div class="overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900/40">
	<div class="flex items-center justify-between gap-3 border-b border-zinc-800 px-4 py-3">
		<div class="flex flex-col gap-0.5">
			<h2 class="text-sm font-semibold text-zinc-100">Snapshot</h2>
			<p class="text-xs text-zinc-500">Latest JPEG frame from the camera.</p>
		</div>
		<button
			type="button"
			class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-xs text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:opacity-50"
			disabled={snapshotQuery.isFetching}
			onclick={() => snapshotQuery.refetch()}
		>
			<RefreshCw class="size-3.5 {snapshotQuery.isFetching ? 'animate-spin' : ''}" />
			Refresh
		</button>
	</div>

	<div
		class="relative flex min-h-[220px] items-center justify-center bg-zinc-950/60"
	>
		{#if snapshotQuery.isPending && !objectUrl}
			<Spinner label="Loading snapshot" />
		{:else if objectUrl && !snapshotQuery.isError}
			<img
				src={objectUrl}
				alt={cameraName ? `Snapshot of ${cameraName}` : "Camera snapshot"}
				class="max-h-[420px] w-full object-contain"
			/>
		{:else}
			<div class="flex flex-col items-center gap-2 px-6 py-10 text-center">
				<span
					class="flex size-10 items-center justify-center rounded-lg bg-zinc-800/80 text-zinc-500"
				>
					<ImageOff class="size-5" />
				</span>
				<p class="text-sm font-medium text-zinc-400">Snapshot unavailable</p>
				<p class="max-w-xs text-xs text-zinc-600">
					{snapshotQuery.error instanceof Error
						? snapshotQuery.error.message
						: "The camera may be offline or not providing frames yet."}
				</p>
			</div>
		{/if}
	</div>
</div>
