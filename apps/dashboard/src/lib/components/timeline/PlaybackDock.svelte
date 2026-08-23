<script lang="ts">
	import { AlertCircle, RefreshCw } from "lucide-svelte";
	import type { CalendarMonth, RecordingDaySource } from "$lib/recordings";
	import MultiCameraTimeline, {
		type CameraCoverageTrack,
	} from "./MultiCameraTimeline.svelte";
	import RecordingCalendar from "./RecordingCalendar.svelte";

	type Props = {
		tracks: CameraCoverageTrack[];
		from: Date;
		to: Date;
		selectedTime: Date | null;
		archiveDate: string;
		maxArchiveDate: string;
		calendarMonth: CalendarMonth;
		calendarAvailability: ReadonlyMap<string, RecordingDaySource>;
		calendarLoading?: boolean;
		calendarError?: boolean;
		loading?: boolean;
		error?: string | null;
		disabled?: boolean;
		onArchiveDateChange: (value: string) => void;
		onCalendarMonthChange: (month: CalendarMonth) => void;
		onSeek: (time: Date) => void;
		onRetry?: () => void;
	};

	let {
		tracks,
		from,
		to,
		selectedTime,
		archiveDate,
		maxArchiveDate,
		calendarMonth,
		calendarAvailability,
		calendarLoading = false,
		calendarError = false,
		loading = false,
		error = null,
		disabled = false,
		onArchiveDateChange,
		onCalendarMonthChange,
		onSeek,
		onRetry,
	}: Props = $props();
</script>

<section
	class="shrink-0 border-t border-zinc-800 bg-zinc-950/95 shadow-[0_-14px_32px_rgba(0,0,0,0.35)]"
	aria-label="Recorded playback timeline"
>
	<div class="max-h-[min(44vh,28rem)] w-full overflow-visible p-2 sm:p-3">
		<div class="relative min-w-0">
			<MultiCameraTimeline
				class="h-full"
				{tracks}
				{from}
				{to}
				{selectedTime}
				disabled={disabled}
				onSeek={onSeek}
			>
				{#snippet controls()}
					<RecordingCalendar
						value={archiveDate}
						max={maxArchiveDate}
						month={calendarMonth}
						availability={calendarAvailability}
						loading={calendarLoading}
						error={calendarError}
						disabled={disabled}
						onChange={onArchiveDateChange}
						onMonthChange={onCalendarMonthChange}
					/>
				{/snippet}
			</MultiCameraTimeline>

			{#if loading}
				<div
					class="pointer-events-none absolute inset-x-0 top-5 bottom-0 z-20 flex items-center justify-center rounded-md border border-zinc-800 bg-zinc-950/80"
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
					class="absolute inset-x-0 top-5 bottom-0 z-20 flex items-center justify-center gap-2 rounded-md border border-red-500/20 bg-zinc-950/90"
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
			{/if}
		</div>
	</div>
</section>
