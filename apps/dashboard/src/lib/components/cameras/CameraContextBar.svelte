<script lang="ts">
	import type { Snippet } from "svelte";
	import { ArrowLeft } from "lucide-svelte";
	import type { Camera } from "$lib/cameras";
	import { primaryCodec, primaryResolution } from "$lib/cameras";
	import PageActions from "$lib/components/PageActions.svelte";

	type Props = {
		camera: Camera;
		actions?: Snippet;
	};

	let { camera, actions }: Props = $props();

	const codec = $derived(primaryCodec(camera.streamProfiles));
	const resolution = $derived(primaryResolution(camera.streamProfiles));
</script>

<PageActions side="start">
	<a
		href="/cameras"
		class="inline-flex items-center gap-1 rounded-md px-1.5 py-1.5 text-zinc-400 no-underline transition-colors hover:bg-zinc-900 hover:text-zinc-100"
		aria-label="Back to cameras"
		title="Back to cameras"
	>
		<ArrowLeft class="size-4" />
		<span class="hidden text-sm sm:inline">Back</span>
	</a>
	<span class="min-w-0 max-w-[34vw] truncate text-sm font-semibold text-zinc-100">
		{camera.name}
	</span>
	{#if codec}
		<span class="hidden rounded-md border border-zinc-700/80 bg-zinc-900 px-2 py-0.5 font-mono text-[11px] text-zinc-300 sm:inline">
			{codec}
		</span>
	{/if}
	{#if resolution}
		<span class="hidden rounded-md border border-zinc-700/80 bg-zinc-900 px-2 py-0.5 font-mono text-[11px] text-zinc-400 sm:inline">
			{resolution}
		</span>
	{/if}
</PageActions>

{#if actions}
	<PageActions>
		{@render actions()}
	</PageActions>
{/if}
