<script lang="ts">
	import type { LiveCamera } from "$lib/live";
	import LiveTile from "./LiveTile.svelte";

	const GRID_SIZE = 9;

	type Props = {
		cameras: LiveCamera[];
	};

	let { cameras }: Props = $props();

	const slots = $derived(
		Array.from({ length: GRID_SIZE }, (_, index) => ({
			channel: index + 1,
			camera: cameras[index] ?? null,
		})),
	);
</script>

<div class="grid h-full w-full grid-cols-3 grid-rows-3 gap-px bg-zinc-900">
	{#each slots as slot (slot.channel)}
		{#if slot.camera}
			<LiveTile camera={slot.camera} channel={slot.channel} />
		{:else}
			<div class="relative min-h-0 min-w-0 bg-black" aria-hidden="true"></div>
		{/if}
	{/each}
</div>
