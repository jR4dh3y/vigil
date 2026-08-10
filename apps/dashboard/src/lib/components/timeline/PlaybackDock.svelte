<script lang="ts">
	import { AlertCircle, RefreshCw } from "lucide-svelte";
	import { type RangePreset } from "$lib/recordings";
	import MultiCameraTimeline, {
		type CameraCoverageTrack,
	} from "./MultiCameraTimeline.svelte";
	import RangePresetButtons from "./RangePresetButtons.svelte";

	type Props = {
		tracks: CameraCoverageTrack[];
		from: Date;
		to: Date;
		selectedTime: Date | null;
		preset: RangePreset;
		loading?: boolean;
		error?: string | null;
		disabled?: boolean;
		onPresetChange: (preset: RangePreset) => void;
		onSeek: (time: Date) => void;
		onRetry?: () => void;
	};

	let {
		tracks,
		from,
		to,
		selectedTime,
		preset,
		loading = false,
		error = null,
		disabled = false,
		onPresetChange,
		onSeek,
		onRetry,
	}: Props = $props();
</script>

<section
	class="shrink-0 border-t border-zinc-800 bg-zinc-950/95 shadow-[0_-14px_32px_rgba(0,0,0,0.35)]"
	aria-label="Recorded playback timeline"
>
	<div
		class="flex max-h-[min(44vh,28rem)] w-full items-stretch gap-2 overflow-y-auto p-2 sm:gap-3 sm:p-3"
	>
		<div class="min-w-0 flex-1">
			{#if loading}
				<div
					class="flex min-h-20 h-full items-center justify-center rounded-lg border border-zinc-800 bg-zinc-900/30"
					role="status"
					aria-label="Loading recording coverage"
				>
					<div
						class="size-4 animate-spin rounded-full border-2 border-zinc-700 border-t-emerald-400"
						aria-hidden="true"
					></div>
				</div>
			{:else if error}
				<div
					class="flex min-h-20 h-full items-center justify-center gap-2 rounded-lg border border-red-500/20 bg-red-500/5"
					aria-label={error}
				>
					<AlertCircle class="size-4 text-red-400" aria-hidden="true" />
					{#if onRetry}
						<button
							type="button"
							class="inline-flex size-7 items-center justify-center rounded-md bg-zinc-800 text-zinc-100 hover:bg-zinc-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-500/70"
							aria-label="Retry recording coverage"
							title="Retry recording coverage"
							onclick={onRetry}
						>
							<RefreshCw class="size-3.5" aria-hidden="true" />
						</button>
					{/if}
				</div>
			{:else}
				<MultiCameraTimeline
					class="h-full"
					{tracks}
					{from}
					{to}
					{selectedTime}
					disabled={disabled}
					onSeek={onSeek}
				/>
			{/if}
		</div>

		<div class="flex shrink-0 items-center border-l border-zinc-800 pl-2 sm:pl-3">
			<RangePresetButtons
				value={preset}
				orientation="vertical"
				disabled={loading || disabled}
				onChange={onPresetChange}
			/>
		</div>
	</div>
</section>

