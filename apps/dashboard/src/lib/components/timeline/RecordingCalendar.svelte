<script lang="ts">
	import { CalendarDays, ChevronLeft, ChevronRight } from "lucide-svelte";
	import type { CalendarDay, CalendarMonth, RecordingDaySource } from "$lib/recordings";
	import {
		calendarGridDays,
		calendarMonthForValue,
		rangeForLocalDate,
		shiftCalendarMonth,
	} from "$lib/recordings";

	type Props = {
		value: string;
		max: string;
		month: CalendarMonth;
		availability: ReadonlyMap<string, RecordingDaySource>;
		loading?: boolean;
		error?: boolean;
		disabled?: boolean;
		onChange: (value: string) => void;
		onMonthChange: (month: CalendarMonth) => void;
	};

	let {
		value,
		max,
		month,
		availability,
		loading = false,
		error = false,
		disabled = false,
		onChange,
		onMonthChange,
	}: Props = $props();

	let open = $state(false);
	const days = $derived(calendarGridDays(month, max));
	const maxMonth = $derived(calendarMonthForValue(max));
	const nextMonthDisabled = $derived(
		maxMonth === null ||
			month.year * 12 + month.month >= maxMonth.year * 12 + maxMonth.month,
	);
	const monthLabel = $derived(
		new Intl.DateTimeFormat(undefined, { month: "long", year: "numeric" }).format(
			new Date(month.year, month.month, 1),
		),
	);
	const selectedLabel = $derived.by(() => {
		const selected = rangeForLocalDate(value)?.from;
		return selected
			? new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric" }).format(selected)
			: "Choose date";
	});

	const weekdayLabels = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"];

	function sourceLabel(source: RecordingDaySource | undefined): string {
		switch (source) {
			case "local":
				return "local recording";
			case "gdrive":
				return "Google Drive recording";
			case "mixed":
				return "local and Google Drive recordings";
			default:
				return "no recording";
		}
	}

	function dayLabel(day: CalendarDay): string {
		const date = rangeForLocalDate(day.value)?.from;
		const label = date
			? new Intl.DateTimeFormat(undefined, {
					weekday: "long",
					month: "long",
					day: "numeric",
					year: "numeric",
				}).format(date)
			: day.value;
		return `${label}, ${sourceLabel(availability.get(day.value))}`;
	}

	function selectDay(day: CalendarDay) {
		if (day.isFuture) {
			return;
		}
		onChange(day.value);
		open = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Escape") {
			open = false;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="relative">
	<button
		type="button"
		class="inline-flex h-6 w-[7.25rem] items-center justify-center gap-1.5 rounded border border-zinc-800 bg-zinc-900/80 px-2 text-[10px] font-medium text-zinc-300 transition-colors hover:border-zinc-700 hover:text-zinc-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-purple-500/70 disabled:cursor-not-allowed disabled:opacity-50"
		disabled={disabled}
		aria-label="Browse recordings by date"
		aria-haspopup="dialog"
		aria-expanded={open}
		onclick={() => (open = !open)}
	>
		<CalendarDays class="size-3" aria-hidden="true" />
		<span>{selectedLabel}</span>
	</button>

	{#if open}
		<button
			type="button"
			class="fixed inset-0 z-40 cursor-default"
			aria-label="Close recording calendar"
			onclick={() => (open = false)}
		></button>
		<div
			class="absolute right-0 bottom-full z-50 mb-2 w-[min(17.5rem,calc(100vw-1rem))] rounded-xl border border-zinc-700/80 bg-zinc-950 p-3 shadow-[0_18px_50px_rgba(0,0,0,0.65)]"
			role="dialog"
			aria-label="Recording calendar"
		>
			<div class="mb-2 flex items-center justify-between">
				<button
					type="button"
					class="inline-flex size-7 items-center justify-center rounded-md text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-purple-500/70"
					aria-label="Previous month"
					onclick={() => onMonthChange(shiftCalendarMonth(month, -1))}
				>
					<ChevronLeft class="size-4" aria-hidden="true" />
				</button>
				<p class="text-xs font-semibold text-zinc-100">{monthLabel}</p>
				<button
					type="button"
					class="inline-flex size-7 items-center justify-center rounded-md text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-purple-500/70 disabled:cursor-not-allowed disabled:opacity-30"
					disabled={nextMonthDisabled}
					aria-label="Next month"
					onclick={() => onMonthChange(shiftCalendarMonth(month, 1))}
				>
					<ChevronRight class="size-4" aria-hidden="true" />
				</button>
			</div>

			<div class="grid grid-cols-7 gap-1" aria-hidden="true">
				{#each weekdayLabels as weekday (weekday)}
					<span class="pb-1 text-center text-[9px] font-medium text-zinc-600">{weekday}</span>
				{/each}
			</div>
			<div class="grid grid-cols-7 gap-1">
				{#each days as day (day.value)}
					{@const source = availability.get(day.value)}
					<button
						type="button"
						class="relative flex aspect-square items-center justify-center rounded-md text-[10px] font-medium transition-colors focus-visible:z-10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-zinc-100
							{source === 'local'
								? 'bg-purple-500/20 text-purple-100 hover:bg-purple-500/30'
								: source === 'gdrive'
									? 'bg-green-500/20 text-green-100 hover:bg-green-500/30'
									: source === 'mixed'
										? 'bg-gradient-to-br from-purple-500/25 to-green-500/25 text-zinc-100 hover:from-purple-500/35 hover:to-green-500/35'
										: 'text-zinc-400 hover:bg-zinc-800'}
							{day.inMonth ? '' : 'opacity-40'}
							{day.value === value ? 'ring-1 ring-inset ring-zinc-100' : ''}
							disabled:cursor-not-allowed disabled:opacity-20"
						disabled={day.isFuture}
						aria-label={dayLabel(day)}
						aria-current={day.value === value ? "date" : undefined}
						title={dayLabel(day)}
						onclick={() => selectDay(day)}
					>
						{day.day}
						{#if source}
							<span class="absolute bottom-1 flex gap-0.5" aria-hidden="true">
								{#if source === "local" || source === "mixed"}
									<span class="size-1 rounded-full bg-purple-400"></span>
								{/if}
								{#if source === "gdrive" || source === "mixed"}
									<span class="size-1 rounded-full bg-green-400"></span>
								{/if}
							</span>
						{/if}
					</button>
				{/each}
			</div>

			<div class="mt-3 flex min-h-4 items-center gap-3 border-t border-zinc-800 pt-2 text-[9px] text-zinc-400">
				<span class="inline-flex items-center gap-1">
					<span class="size-1.5 rounded-full bg-purple-400"></span>Local
				</span>
				<span class="inline-flex items-center gap-1">
					<span class="size-1.5 rounded-full bg-green-400"></span>Google Drive
				</span>
				{#if loading}
					<span class="ml-auto text-zinc-500" role="status">Loading…</span>
				{:else if error}
					<span class="ml-auto text-amber-400" role="status">Markers unavailable</span>
				{/if}
			</div>
		</div>
	{/if}
</div>
