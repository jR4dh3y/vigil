<script lang="ts">
	import type { CoverageBar } from "$lib/recordings";
	import ScrollableTimeline from "./ScrollableTimeline.svelte";

	type TimelineEvent = {
		at: string;
		label?: string;
	};

	type Props = {
		coverage: CoverageBar[];
		from: Date;
		to: Date;
		selectedTime: Date | null;
		/** Reserved for future event markers; currently unused. */
		events?: TimelineEvent[];
		disabled?: boolean;
		onPreview?: (time: Date) => void;
		onSeek: (time: Date) => void;
		class?: string;
	};

	let {
		coverage,
		from,
		to,
		selectedTime,
		events: _events = [],
		disabled = false,
		onPreview,
		onSeek,
		class: className = "",
	}: Props = $props();
</script>

{#key `${from.getTime()}-${to.getTime()}`}
	<ScrollableTimeline
		{coverage}
		{from}
		{to}
		{selectedTime}
		{disabled}
		ariaLabel="Recording timeline"
		{onPreview}
		{onSeek}
		class={className}
	/>
{/key}
