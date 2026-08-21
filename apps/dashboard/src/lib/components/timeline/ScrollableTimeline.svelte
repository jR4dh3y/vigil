<script lang="ts">
	import { tick } from "svelte";
	import type { CoverageBar } from "$lib/recordings";
	import {
		buildTimelineTicks,
		coverageBarRects,
		formatSelectedTime,
		fractionAtTime,
		timeAtFraction,
		timelineTickInterval,
		timelineTrackWidth,
	} from "$lib/recordings";

	type Props = {
		coverage: CoverageBar[];
		from: Date;
		to: Date;
		selectedTime: Date | null;
		disabled?: boolean;
		tone?: "emerald" | "violet";
		ariaLabel: string;
		onPreview?: (time: Date) => void;
		onSeek: (time: Date) => void;
		class?: string;
	};

	let {
		coverage,
		from,
		to,
		selectedTime,
		disabled = false,
		tone = "emerald",
		ariaLabel,
		onPreview,
		onSeek,
		class: className = "",
	}: Props = $props();

	let scrollerEl = $state<HTMLDivElement | null>(null);
	let trackEl = $state<HTMLDivElement | null>(null);
	let viewportWidth = $state(0);
	let zoom = $state(1);
	let dragging = $state(false);
	let pointerTime = $state<Date | null>(null);
	let touchStart: { clientX: number; scrollLeft: number } | null = null;

	const rangeMs = $derived(Math.max(1, to.getTime() - from.getTime()));
	const trackWidth = $derived(timelineTrackWidth(viewportWidth, zoom));
	const selectedFraction = $derived(
		selectedTime ? fractionAtTime(from, to, selectedTime) : null,
	);
	const exactTime = $derived(pointerTime ?? selectedTime);
	const bars = $derived(coverageBarRects(coverage, from, to));
	const ticks = $derived(buildTimelineTicks(from, to, trackWidth));
	const keyboardStepMs = $derived(timelineTickInterval(rangeMs, trackWidth));

	const MIN_ZOOM = 1;
	const MAX_ZOOM = 64;

	function timeFromPointer(clientX: number): Date | null {
		const track = trackEl;
		if (!track) {
			return null;
		}
		const rect = track.getBoundingClientRect();
		if (rect.width <= 0) {
			return null;
		}
		return timeAtFraction(from, to, (clientX - rect.left) / rect.width);
	}

	function updatePointerTime(clientX: number, preview: boolean): Date | null {
		const time = timeFromPointer(clientX);
		if (time) {
			pointerTime = time;
			if (preview) {
				onPreview?.(time);
			}
		}
		return time;
	}

	function capturePointer(target: EventTarget | null, pointerId: number) {
		if (target instanceof HTMLElement) {
			target.setPointerCapture(pointerId);
		}
	}

	function releasePointer(target: EventTarget | null, pointerId: number) {
		if (!(target instanceof HTMLElement) || !target.hasPointerCapture(pointerId)) {
			return;
		}
		target.releasePointerCapture(pointerId);
	}

	function onPointerDown(event: PointerEvent) {
		if (disabled || event.button !== 0) {
			return;
		}
		if (event.pointerType === "touch") {
			touchStart = {
				clientX: event.clientX,
				scrollLeft: scrollerEl?.scrollLeft ?? 0,
			};
			return;
		}
		dragging = true;
		capturePointer(event.currentTarget, event.pointerId);
		updatePointerTime(event.clientX, true);
	}

	function onPointerMove(event: PointerEvent) {
		if (disabled || event.pointerType === "touch") {
			return;
		}
		updatePointerTime(event.clientX, dragging);
	}

	function onPointerUp(event: PointerEvent) {
		if (event.pointerType === "touch") {
			const start = touchStart;
			touchStart = null;
			const scrollDistance = Math.abs((scrollerEl?.scrollLeft ?? 0) - (start?.scrollLeft ?? 0));
			if (start && Math.abs(event.clientX - start.clientX) < 6 && scrollDistance < 6) {
				const time = updatePointerTime(event.clientX, false);
				if (time) {
					onSeek(time);
				}
			}
			return;
		}
		if (!dragging) {
			return;
		}
		const time = updatePointerTime(event.clientX, false);
		dragging = false;
		if (time) {
			onSeek(time);
		}
		releasePointer(event.currentTarget, event.pointerId);
	}

	function onPointerCancel(event: PointerEvent) {
		touchStart = null;
		dragging = false;
		pointerTime = null;
		releasePointer(event.currentTarget, event.pointerId);
	}

	function onPointerLeave() {
		if (!dragging) {
			pointerTime = null;
		}
	}

	function onKeyDown(event: KeyboardEvent) {
		if (disabled) {
			return;
		}
		const base = selectedTime ?? to;
		if (event.key === "ArrowLeft") {
			event.preventDefault();
			onSeek(new Date(Math.max(from.getTime(), base.getTime() - keyboardStepMs)));
		} else if (event.key === "ArrowRight") {
			event.preventDefault();
			onSeek(new Date(Math.min(to.getTime(), base.getTime() + keyboardStepMs)));
		} else if (event.key === "Home") {
			event.preventDefault();
			onSeek(from);
		} else if (event.key === "End") {
			event.preventDefault();
			onSeek(to);
		}
	}

	async function setZoom(nextZoom: number, anchorX: number) {
		const scroller = scrollerEl;
		if (!scroller) {
			return;
		}
		const clampedZoom = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, nextZoom));
		if (clampedZoom === zoom) {
			return;
		}
		const timeFraction = (scroller.scrollLeft + anchorX) / Math.max(1, scroller.scrollWidth);
		zoom = clampedZoom;
		await tick();
		scroller.scrollLeft = timeFraction * scroller.scrollWidth - anchorX;
	}

	function zoomFromButton(direction: "in" | "out") {
		const scroller = scrollerEl;
		if (disabled || !scroller) {
			return;
		}
		void setZoom(zoom * (direction === "in" ? 1.5 : 1 / 1.5), scroller.clientWidth / 2);
	}

	function onWheel(event: WheelEvent) {
		const scroller = scrollerEl;
		if (disabled || !scroller) {
			return;
		}
		if (Math.abs(event.deltaX) > Math.abs(event.deltaY)) {
			if (zoom > MIN_ZOOM) {
				event.preventDefault();
				scroller.scrollLeft += event.deltaX;
			}
			return;
		}
		if (event.deltaY === 0) {
			return;
		}
		const deltaScale =
			event.deltaMode === WheelEvent.DOM_DELTA_LINE
				? 16
				: event.deltaMode === WheelEvent.DOM_DELTA_PAGE
					? scroller.clientHeight
					: 1;
		const normalizedDelta = Math.min(100, Math.max(-100, event.deltaY * deltaScale));
		const nextZoom = zoom * Math.exp(-normalizedDelta * 0.0025);
		const clampedZoom = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, nextZoom));
		if (clampedZoom === zoom) {
			return;
		}
		event.preventDefault();
		const rect = scroller.getBoundingClientRect();
		const anchorX = Math.min(scroller.clientWidth, Math.max(0, event.clientX - rect.left));
		void setZoom(clampedZoom, anchorX);
	}
</script>

<div class="flex min-w-0 flex-col gap-1.5 {className}">
	<div
		class="flex min-h-4 flex-wrap items-center justify-between gap-x-3 gap-y-1 px-1 text-[10px] text-zinc-500"
	>
		<span>Scroll to zoom · swipe to pan</span>
		<div class="flex items-center gap-2">
			<output class="font-medium text-zinc-300 tabular-nums">
				{exactTime ? formatSelectedTime(exactTime) : "Point to a timestamp"}
			</output>
			<div class="inline-flex items-center rounded border border-zinc-800 bg-zinc-900/80">
				<button
					type="button"
					class="px-1.5 py-0.5 text-zinc-400 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-40"
					disabled={disabled || zoom <= MIN_ZOOM}
					aria-label="Zoom timeline out"
					onclick={() => zoomFromButton("out")}
				>−</button
				>
				<output class="min-w-10 border-x border-zinc-800 px-1 text-center text-zinc-300 tabular-nums">
					{Math.round(zoom * 100)}%
				</output>
				<button
					type="button"
					class="px-1.5 py-0.5 text-zinc-400 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-40"
					disabled={disabled || zoom >= MAX_ZOOM}
					aria-label="Zoom timeline in"
					onclick={() => zoomFromButton("in")}
				>+</button
				>
			</div>
		</div>
	</div>

	<div
		bind:this={scrollerEl}
		bind:clientWidth={viewportWidth}
		class="min-w-0 overflow-x-auto overscroll-x-contain pb-1.5"
		onwheel={onWheel}
	>
		<div class="flex flex-col gap-1.5" style:width="{trackWidth}px">
			<div
				bind:this={trackEl}
				class="relative h-10 touch-pan-x select-none rounded-md border border-zinc-800 bg-zinc-950 outline-none
					focus-visible:ring-2 focus-visible:ring-emerald-500
					{disabled ? 'cursor-not-allowed opacity-60' : 'cursor-crosshair'}"
				role="slider"
				tabindex={disabled ? -1 : 0}
				aria-label={ariaLabel}
				aria-valuemin={from.getTime()}
				aria-valuemax={to.getTime()}
				aria-valuenow={selectedTime?.getTime() ?? undefined}
				aria-valuetext={selectedTime ? formatSelectedTime(selectedTime) : undefined}
				aria-disabled={disabled}
				onpointerdown={onPointerDown}
				onpointermove={onPointerMove}
				onpointerup={onPointerUp}
				onpointercancel={onPointerCancel}
				onpointerleave={onPointerLeave}
				onkeydown={onKeyDown}
			>
				<div class="absolute inset-1 overflow-hidden rounded-sm bg-zinc-900/90">
					{#each bars as bar (bar.start + bar.end)}
						<div
							class="absolute top-0 bottom-0 {tone === 'violet'
								? 'bg-violet-400/70'
								: 'bg-emerald-500/70'}"
							style:left="{bar.left * 100}%"
							style:width="{bar.width * 100}%"
							title="{formatSelectedTime(new Date(bar.start))} to {formatSelectedTime(
								new Date(bar.end),
							)}"
						></div>
					{/each}
				</div>

				{#if selectedFraction !== null}
					<div
						class="pointer-events-none absolute top-0 bottom-0 z-10 w-0.5 -translate-x-1/2 bg-white shadow-[0_0_6px_rgba(255,255,255,0.5)]"
						style:left="{selectedFraction * 100}%"
					>
						<div
							class="absolute top-1/2 left-1/2 size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white {tone ===
							'violet'
								? 'bg-amber-300'
								: 'bg-emerald-400'}"
						></div>
					</div>
				{/if}
			</div>

			<div class="relative h-4 px-1">
				{#each ticks as tick (tick.at.getTime())}
					<span
						class="absolute whitespace-nowrap -translate-x-1/2 text-[10px] text-zinc-500 tabular-nums"
						style:left="{tick.fraction * 100}%"
					>
						{tick.label}
					</span>
				{/each}
			</div>
		</div>
	</div>
</div>
