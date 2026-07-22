<script lang="ts">
	import { ChevronRight, Film } from "lucide-svelte";
	import type { Camera } from "$lib/cameras";
	import { primaryCodec, primaryResolution } from "$lib/cameras";
	import CameraStatusBadge from "./CameraStatusBadge.svelte";

	type Props = {
		camera: Camera;
	};

	let { camera }: Props = $props();

	const codec = $derived(primaryCodec(camera.streamProfiles));
	const resolution = $derived(primaryResolution(camera.streamProfiles));
</script>

<div
	class="group flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 transition-colors hover:border-zinc-700 hover:bg-zinc-900"
>
	<a href="/cameras/{camera.id}" class="flex items-start justify-between gap-3 no-underline">
		<div class="min-w-0">
			<h2 class="truncate text-sm font-semibold text-zinc-100 group-hover:text-white">
				{camera.name}
			</h2>
			<p class="mt-0.5 truncate font-mono text-xs text-zinc-500">{camera.host}</p>
		</div>
		<div class="flex shrink-0 items-center gap-2">
			<CameraStatusBadge status={camera.status} />
			<ChevronRight
				class="size-4 text-zinc-600 transition-colors group-hover:text-zinc-400"
			/>
		</div>
	</a>

	<div class="flex flex-wrap items-center justify-between gap-2 text-xs">
		<div class="flex flex-wrap items-center gap-2">
			<span
				class="rounded-md border px-2 py-0.5
					{camera.enabled
					? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300'
					: 'border-zinc-700 bg-zinc-800 text-zinc-500'}"
			>
				{camera.enabled ? "Enabled" : "Disabled"}
			</span>
			{#if codec}
				<span
					class="rounded-md border border-zinc-700/80 bg-zinc-800/80 px-2 py-0.5 text-zinc-300"
				>
					{codec}
				</span>
			{/if}
			{#if resolution}
				<span
					class="rounded-md border border-zinc-700/80 bg-zinc-800/80 px-2 py-0.5 text-zinc-400"
				>
					{resolution}
				</span>
			{/if}
		</div>
		<a
			href="/cameras/{camera.id}/timeline"
			class="inline-flex items-center gap-1 rounded-md border border-zinc-700/80 bg-zinc-800/60 px-2 py-0.5 text-zinc-300 no-underline transition-colors hover:border-emerald-500/30 hover:bg-emerald-500/10 hover:text-emerald-300"
		>
			<Film class="size-3" />
			Timeline
		</a>
	</div>
</div>
