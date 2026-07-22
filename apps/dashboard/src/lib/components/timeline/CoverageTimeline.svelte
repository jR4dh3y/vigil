<script lang="ts">
	import type { CoverageBar } from "$lib/recordings";
	import {
		formatTimelineLabel,
		fractionAtTime,
		timeAtFraction,
	} from "$lib/recordings";

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
		/** Called while dragging / previewing a scrub position. */
		onPreview?: (time: Date) => void;
		/** Called when the user commits a seek (click or drag end). */
		onSeek: (time: Date) => void;
		class?: string;
	};

	let {
		coverage,
		from,
		to,
		selectedTime,
		// Reserved for future event markers on the timeline.
		events: _events = [],
		disabled = false,
		onPreview,
		onSeek,
		class: className = "",
	}: Props = $props();

	let trackEl = $state<HTMLDivElement | null>(null);
	let dragging = $state(false);

	const rangeMs = $derived(Math.max(1, to.getTime() - from.getTime()));
	const selectedFraction = $derived(
		selectedTime ? fractionAtTime(from, to, selectedTime) : null,
	);

	const bars = $derived(
		coverage
			.map((bar) => {
				const start = Date.parse(bar.start);
				const end = Date.parse(bar.end);
				if (Number.isNaN(start) || Number.isNaN(end) || end <= start) {
					return null;
				}
				const left = fractionAtTime(from, to, new Date(start));
				const right = fractionAtTime(from, to, new Date(end));
				const width = Math.max(right - left, 0.001);
				return { left, width, start: bar.start, end: bar.end };
			})
			.filter((b): b is { left: number; width: number; start: string; end: string } => b !== null),
	);

	const tickCount = $derived(
		rangeMs <= 2 * 60 * 60 * 1000 ? 5 : rangeMs <= 2 * 24 * 60 * 60 * 1000 ? 6 : 7,
	);

	const ticks = $derived(
		Array.from({ length: tickCount }, (_, i) => {
			const fraction = i / (tickCount - 1);
			const time = timeAtFraction(from, to, fraction);
			return { fraction, label: formatTimelineLabel(time, rangeMs) };
		}),
	);

	function timeFromPointer(clientX: number): Date | null {
		const el = trackEl;
		if (!el) {
			return null;
		}
		const rect = el.getBoundingClientRect();
		if (rect.width <= 0) {
			return null;
		}
		const fraction = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
		return timeAtFraction(from, to, fraction);
	}

	function previewFromPointer(clientX: number): Date | null {
		if (disabled) {
			return null;
		}
		const time = timeFromPointer(clientX);
		if (time) {
			onPreview?.(time);
		}
		return time;
	}

	function onPointerDown(event: PointerEvent) {
		if (disabled || event.button !== 0) {
			return;
		}
		dragging = true;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
		previewFromPointer(event.clientX);
	}

	function onPointerMove(event: PointerEvent) {
		if (!dragging || disabled) {
			return;
		}
		previewFromPointer(event.clientX);
	}

	function onPointerUp(event: PointerEvent) {
		if (!dragging) {
			return;
		}
		dragging = false;
		const time = previewFromPointer(event.clientX);
		if (time) {
			onSeek(time);
		}
		try {
			(event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId);
		} catch {
			// already released
		}
	}

	function onKeyDown(event: KeyboardEvent) {
		if (disabled) {
			return;
		}
		const step = rangeMs / 100;
		const base = selectedTime ?? to;
		if (event.key === "ArrowLeft") {
			event.preventDefault();
			onSeek(new Date(Math.max(from.getTime(), base.getTime() - step)));
		} else if (event.key === "ArrowRight") {
			event.preventDefault();
			onSeek(new Date(Math.min(to.getTime(), base.getTime() + step)));
		} else if (event.key === "Home") {
			event.preventDefault();
			onSeek(from);
		} else if (event.key === "End") {
			event.preventDefault();
			onSeek(to);
		}
	}
</script>

<div class="flex flex-col gap-2 {className}">
	<div
		bind:this={trackEl}
		class="relative h-12 w-full touch-none select-none rounded-lg border border-zinc-800 bg-zinc-950 outline-none
			{disabled ? 'cursor-not-allowed opacity-60' : 'cursor-crosshair'}"
		role="slider"
		tabindex={disabled ? -1 : 0}
		aria-label="Recording timeline"
		aria-valuemin={from.getTime()}
		aria-valuemax={to.getTime()}
		aria-valuenow={selectedTime?.getTime() ?? undefined}
		aria-disabled={disabled}
		onpointerdown={onPointerDown}
		onpointermove={onPointerMove}
		onpointerup={onPointerUp}
		onpointercancel={onPointerUp}
		onkeydown={onKeyDown}
	>
		<!-- Track fill -->
		<div class="absolute inset-1 overflow-hidden rounded-md bg-zinc-900/90">
			{#each bars as bar (bar.start + bar.end)}
				<div
					class="absolute top-0 bottom-0 rounded-sm bg-emerald-500/70"
					style:left="{bar.left * 100}%"
					style:width="{bar.width * 100}%"
					title="Coverage"
				></div>
			{/each}
		</div>

		{#if selectedFraction !== null}
			<div
				class="pointer-events-none absolute top-0 bottom-0 z-10 w-0.5 -translate-x-1/2 bg-white shadow-[0_0_6px_rgba(255,255,255,0.5)]"
				style:left="{selectedFraction * 100}%"
			>
				<div
					class="absolute top-1/2 left-1/2 size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white bg-emerald-400"
				></div>
			</div>
		{/if}
	</div>

	<div class="relative h-4 px-1">
		{#each ticks as tick (tick.fraction)}
			<span
				class="absolute -translate-x-1/2 text-[10px] text-zinc-500 tabular-nums
					{tick.fraction === 0 ? 'translate-x-0' : ''}
					{tick.fraction === 1 ? '-translate-x-full' : ''}"
				style:left="{tick.fraction * 100}%"
			>
				{tick.label}
			</span>
		{/each}
	</div>
</div>
