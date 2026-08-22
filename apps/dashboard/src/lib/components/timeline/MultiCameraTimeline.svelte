<script lang="ts" module>
	import type { CoverageBar } from "$lib/recordings";

	/** Camera coverage sources that are merged into one shared scrub bar. */
	export type CameraCoverageTrack = {
		id: string;
		channel: number;
		name: string;
		coverage: CoverageBar[];
	};
</script>

<script lang="ts">
	import type { Snippet } from "svelte";
	import { mergeCoverageBars } from "$lib/recordings";
	import ScrollableTimeline from "./ScrollableTimeline.svelte";

	type Props = {
		tracks: CameraCoverageTrack[];
		from: Date;
		to: Date;
		selectedTime: Date | null;
		disabled?: boolean;
		controls?: Snippet;
		onSeek: (time: Date) => void;
		class?: string;
	};

	let {
		tracks,
		from,
		to,
		selectedTime,
		disabled = false,
		controls,
		onSeek,
		class: className = "",
	}: Props = $props();

	const coverage = $derived(mergeCoverageBars(tracks.flatMap((track) => track.coverage)));
</script>

<ScrollableTimeline
	{coverage}
	{from}
	{to}
	{selectedTime}
	{disabled}
	tone="violet"
	ariaLabel="Recording scrub bar for all cameras"
	{controls}
	{onSeek}
	class={className}
/>
