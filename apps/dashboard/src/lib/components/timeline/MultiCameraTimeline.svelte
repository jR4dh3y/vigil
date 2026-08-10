<script lang="ts">
	import {
		formatTimelineLabel,
		fractionAtTime,
		timeAtFraction,
	} from "$lib/recordings";
	import type { CoverageBar } from "$lib/recordings";

	/** Camera coverage sources that are merged into one shared scrub bar. */
	export type CameraCoverageTrack = {
		id: string;
		channel: number;
		name: string;
		coverage: CoverageBar[];
	};

	type Props = {
		tracks: CameraCoverageTrack[];
		from: Date;
		to: Date;
		selectedTime: Date | null;
		disabled?: boolean;
		/** Called when the user commits a seek (click or drag end). */
		onSeek: (time: Date) => void;
		class?: string;
	};

	let {
		tracks,
		from,
		to,
		selectedTime,
		disabled = false,
		onSeek,
		class: className = "",
	}: Props = $props();

	type CoverageRange = {
		startMs: number;
		endMs: number;
		start: string;
		end: string;
	};

	type BarRect = {
		left: number;
		width: number;
		start: string;
		end: string;
	};

	let trackEl = $state<HTMLDivElement | null>(null);
	let dragging = $state(false);
	let previewTime = $state<Date | null>(null);

	const rangeMs = $derived(Math.max(1, to.getTime() - from.getTime()));
	const displayTime = $derived(dragging ? previewTime : selectedTime);
	const selectedFraction = $derived(
		displayTime ? fractionAtTime(from, to, displayTime) : null,
	);

	/** Merge every camera's recording windows into one shared coverage track. */
	const bars = $derived.by(() => {
		const ranges = tracks
			.flatMap((track) => track.coverage)
			.map((bar): CoverageRange | null => {
				const startMs = Date.parse(bar.start);
				const endMs = Date.parse(bar.end);
				if (Number.isNaN(startMs) || Number.isNaN(endMs) || endMs <= startMs) {
					return null;
				}
				return { startMs, endMs, start: bar.start, end: bar.end };
			})
			.filter((range): range is CoverageRange => range !== null)
			.sort((left, right) => left.startMs - right.startMs);

		const merged: CoverageRange[] = [];
		for (const range of ranges) {
			const previous = merged[merged.length - 1];
			if (previous && range.startMs <= previous.endMs) {
				previous.endMs = Math.max(previous.endMs, range.endMs);
				previous.end = previous.endMs === range.endMs ? range.end : previous.end;
			} else {
				merged.push({ ...range });
			}
		}

		return merged.map((range): BarRect => {
			const left = fractionAtTime(from, to, new Date(range.startMs));
			const right = fractionAtTime(from, to, new Date(range.endMs));
			return {
				left,
				width: Math.max(right - left, 0.001),
				start: range.start,
				end: range.end,
			};
		});
	});

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
			previewTime = time;
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

	function releasePointer(event: PointerEvent) {
		try {
			(event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId);
		} catch {
			// Pointer capture may already be released by the browser.
		}
	}

	function onPointerUp(event: PointerEvent) {
		if (!dragging) {
			return;
		}
		const time = disabled ? null : timeFromPointer(event.clientX);
		dragging = false;
		previewTime = null;
		if (time) {
			onSeek(time);
		}
		releasePointer(event);
	}

	function onPointerCancel(event: PointerEvent) {
		dragging = false;
		previewTime = null;
		releasePointer(event);
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
		class="relative h-9 select-none touch-none rounded-md border border-zinc-800 bg-zinc-950 outline-none
			focus-visible:ring-2 focus-visible:ring-emerald-500
			{disabled ? 'cursor-not-allowed opacity-60' : 'cursor-crosshair'}"
		role="slider"
		tabindex={disabled ? -1 : 0}
		aria-label="Recording scrub bar for all cameras"
		aria-valuemin={from.getTime()}
		aria-valuemax={to.getTime()}
		aria-valuenow={displayTime?.getTime() ?? undefined}
		aria-disabled={disabled}
		onpointerdown={onPointerDown}
		onpointermove={onPointerMove}
		onpointerup={onPointerUp}
		onpointercancel={onPointerCancel}
		onkeydown={onKeyDown}
	>
		<div class="absolute inset-1 overflow-hidden rounded-sm bg-zinc-900/90">
			{#each bars as bar (bar.start + bar.end)}
				<div
					class="absolute top-0 bottom-0 bg-violet-400/70"
					style:left="{bar.left * 100}%"
					style:width="{bar.width * 100}%"
					title="Recording coverage"
				></div>
			{/each}
		</div>

		{#if selectedFraction !== null}
			<div
				class="pointer-events-none absolute top-0 bottom-0 z-10 w-0.5 -translate-x-1/2 bg-white shadow-[0_0_6px_rgba(255,255,255,0.5)]"
				style:left="{selectedFraction * 100}%"
			>
				<div
					class="absolute top-1/2 left-1/2 size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white bg-amber-300"
				></div>
			</div>
		{/if}
	</div>

	<div class="relative h-4 px-1">
		{#each ticks as tick (tick.fraction)}
			<span
				class="absolute whitespace-nowrap -translate-x-1/2 text-[10px] text-zinc-500 tabular-nums
					{tick.fraction === 0 ? 'translate-x-0' : ''}
					{tick.fraction === 1 ? '-translate-x-full' : ''}"
				style:left="{tick.fraction * 100}%"
			>
				{tick.label}
			</span>
		{/each}
	</div>
</div>