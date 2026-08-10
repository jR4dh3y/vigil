<script lang="ts">
	import { ChevronLeft, ChevronRight, LayoutGrid } from "lucide-svelte";
	import { calculateCameraGridLayout, type LiveCamera } from "$lib/live";
	import PageActions from "$lib/components/PageActions.svelte";
	import LiveTile from "./LiveTile.svelte";

	const GRID_SIZE = 9;

	type Props = {
		cameras: LiveCamera[];
	};

	let { cameras }: Props = $props();

	/** Index into `cameras` when focused; null shows the grid. */
	let focusedIndex = $state<number | null>(null);
	let gridWidth = $state(0);
	let gridHeight = $state(0);

	const visibleCameraCount = $derived(Math.min(cameras.length, GRID_SIZE));
	const slots = $derived(
		Array.from({ length: visibleCameraCount }, (_, index) => ({
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

	const focused = $derived.by(() => {
		if (focusedIndex === null) {
			return null;
		}
		const camera = cameras[focusedIndex];
		if (!camera) {
			return null;
		}
		return {
			camera,
			channel: focusedIndex + 1,
			index: focusedIndex,
		};
	});

	const canBrowse = $derived(cameras.length > 1);

	// Keep focus valid if the camera list changes.
	$effect(() => {
		if (focusedIndex === null) {
			return;
		}
		if (cameras.length === 0) {
			focusedIndex = null;
			return;
		}
		if (focusedIndex >= cameras.length) {
			focusedIndex = cameras.length - 1;
		}
	});

	function openCamera(index: number) {
		focusedIndex = index;
	}

	function closeFocused() {
		focusedIndex = null;
	}

	function goPrev() {
		if (focusedIndex === null || cameras.length === 0) {
			return;
		}
		focusedIndex = (focusedIndex - 1 + cameras.length) % cameras.length;
	}

	function goNext() {
		if (focusedIndex === null || cameras.length === 0) {
			return;
		}
		focusedIndex = (focusedIndex + 1) % cameras.length;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (focusedIndex === null) {
			return;
		}
		const target = event.target;
		if (
			target instanceof HTMLElement &&
			(target.isContentEditable ||
				target.tagName === "INPUT" ||
				target.tagName === "TEXTAREA" ||
				target.tagName === "SELECT")
		) {
			return;
		}

		if (event.key === "Escape") {
			event.preventDefault();
			closeFocused();
			return;
		}
		if (event.key === "ArrowLeft") {
			event.preventDefault();
			goPrev();
			return;
		}
		if (event.key === "ArrowRight") {
			event.preventDefault();
			goNext();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if focused}
	<PageActions side="start">
		<span class="min-w-0 max-w-[50vw] truncate text-sm text-zinc-400">
			CH{focused.channel} · {focused.camera.name}
		</span>
	</PageActions>

	<PageActions>
		<button
			type="button"
			class="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100"
			onclick={closeFocused}
		>
			<LayoutGrid class="size-3.5" />
			<span class="hidden sm:inline">Back</span>
		</button>
		<button
			type="button"
			class="inline-flex items-center gap-1 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-40"
			disabled={!canBrowse}
			onclick={goPrev}
			aria-label="Previous camera"
			title="Previous camera (←)"
		>
			<ChevronLeft class="size-3.5" />
			<span class="hidden sm:inline">Prev</span>
		</button>
		<button
			type="button"
			class="inline-flex items-center gap-1 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-40"
			disabled={!canBrowse}
			onclick={goNext}
			aria-label="Next camera"
			title="Next camera (→)"
		>
			<span class="hidden sm:inline">Next</span>
			<ChevronRight class="size-3.5" />
		</button>
	</PageActions>

	<div class="h-full min-h-0 w-full bg-black">
		{#key focused.camera.id}
			<LiveTile camera={focused.camera} channel={focused.channel} focused />
		{/key}
	</div>
{:else}
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
				{@const index = slot.channel - 1}
				<LiveTile
					{camera}
					channel={slot.channel}
					onSelect={() => openCamera(index)}
				/>
			{:else}
				<div class="relative min-h-0 min-w-0 bg-black" aria-hidden="true"></div>
			{/if}
		{/each}
	</div>
{/if}
