<script lang="ts">
	import { calculateCameraGridLayout, type LiveCamera } from "$lib/live";
	import type { PlaybackSession } from "$lib/recordings";
	import PlaybackPlayer from "$lib/components/timeline/PlaybackPlayer.svelte";

	const GRID_SIZE = 9;

	type Props = {
		cameras: LiveCamera[];
		/** Sessions keyed by camera id; a missing entry renders the no-session state. */
		sessions: ReadonlyMap<string, PlaybackSession>;
		/** Global loading flag; only shown for cameras without a session yet. */
		loading?: boolean;
		/** Per-camera errors keyed by camera id. */
		errors?: ReadonlyMap<string, string>;
		onEnded?: (session: PlaybackSession) => void;
	};

	let {
		cameras,
		sessions,
		loading = false,
		errors,
		onEnded,
	}: Props = $props();

	type Slot = {
		channel: number;
		camera: LiveCamera | null;
	};

	let gridWidth = $state(0);
	let gridHeight = $state(0);

	const visibleCameraCount = $derived(Math.min(cameras.length, GRID_SIZE));
	const slots = $derived(
		Array.from({ length: visibleCameraCount }, (_, index): Slot => ({
			channel: index + 1,
			camera: cameras[index] ?? null,
		})),
	);
	const gridLayout = $derived(
		calculateCameraGridLayout(visibleCameraCount, gridWidth, gridHeight),
	);
	const gridColumns = $derived(
		`repeat(${gridLayout.columns}, minmax(0, 1fr))`,
	);
	const gridRows = $derived(`repeat(${gridLayout.rows}, minmax(0, 1fr))`);
</script>

<div
	bind:clientWidth={gridWidth}
	bind:clientHeight={gridHeight}
	class="grid h-full min-h-0 w-full min-w-0 gap-px overflow-hidden bg-zinc-900"
	style:grid-template-columns={gridColumns}
	style:grid-template-rows={gridRows}
>
	{#each slots as slot (slot.channel)}
		{#if slot.camera}
			{@const camera = slot.camera}
			{@const session = sessions.get(camera.id) ?? null}
			{@const error = errors?.get(camera.id) ?? null}

			<div class="group relative h-full min-h-0 w-full min-w-0 overflow-hidden bg-black">
				<div class="absolute inset-0">
					<PlaybackPlayer
						session={session}
						error={error}
						onEnded={onEnded}
						loading={loading && !session}
						class="h-full !aspect-auto rounded-none border-0"
						videoClass="object-cover"
					/>
				</div>
				<!-- Channel label: always visible -->
				<div
					class="pointer-events-none absolute inset-x-0 bottom-0 z-10 bg-gradient-to-t from-black/70 to-transparent px-2.5 py-2"
				>
					<p class="truncate text-xs font-medium text-white drop-shadow">
						CH{slot.channel} · {camera.name}
					</p>
				</div>
			</div>
		{:else}
			<div class="relative min-h-0 min-w-0 bg-black" aria-hidden="true"></div>
		{/if}
	{/each}
</div>
