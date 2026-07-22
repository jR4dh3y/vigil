<script lang="ts">
	import type { LiveCamera } from "$lib/live";
	import LiveTile from "./LiveTile.svelte";

	type Props = {
		cameras: LiveCamera[];
	};

	let { cameras }: Props = $props();
</script>

<div
	class="grid gap-3
		{cameras.length === 1
		? 'grid-cols-1'
		: cameras.length === 2
			? 'grid-cols-1 sm:grid-cols-2'
			: cameras.length <= 4
				? 'grid-cols-1 sm:grid-cols-2'
				: 'grid-cols-1 sm:grid-cols-2 xl:grid-cols-3'}"
>
	{#each cameras as camera (camera.id)}
		<LiveTile {camera} />
	{/each}
</div>
