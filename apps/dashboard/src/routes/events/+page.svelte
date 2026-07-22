<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import { RefreshCw } from "lucide-svelte";
	import {
		acknowledgeEvent,
		EventApiError,
		eventKeys,
		listEvents,
		type ListEventsParams,
	} from "$lib/events";
	import EventRow from "$lib/components/events/EventRow.svelte";
	import EventsEmptyState from "$lib/components/events/EventsEmptyState.svelte";
	import Spinner from "$lib/components/Spinner.svelte";

	const queryClient = useQueryClient();

	let unacknowledgedOnly = $state(false);
	let acknowledgingId = $state<string | null>(null);
	let actionError = $state<string | null>(null);

	const listParams = $derived<ListEventsParams>({
		unacknowledgedOnly: unacknowledgedOnly || undefined,
		limit: 100,
	});

	const eventsQuery = createQuery(() => ({
		queryKey: eventKeys.list(listParams),
		queryFn: () => listEvents(listParams),
		refetchInterval: 15_000,
	}));

	const ackMutation = createMutation(() => ({
		mutationFn: (id: string) => acknowledgeEvent(id),
		onMutate: (id) => {
			acknowledgingId = id;
			actionError = null;
		},
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: eventKeys.all });
		},
		onError: (error: unknown) => {
			if (error instanceof EventApiError) {
				actionError = error.message;
				return;
			}
			actionError = error instanceof Error ? error.message : "Failed to acknowledge event";
		},
		onSettled: () => {
			acknowledgingId = null;
		},
	}));

	const events = $derived(eventsQuery.data ?? []);

	async function handleAcknowledge(id: string) {
		await ackMutation.mutateAsync(id);
	}
</script>

<svelte:head>
	<title>Events · NVR</title>
</svelte:head>

<section class="flex flex-col gap-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex flex-col gap-1">
			<h1 class="text-2xl font-semibold tracking-tight text-zinc-50">Events</h1>
			<p class="text-sm text-zinc-400">
				System and camera alerts. Auto-refreshes every 15 seconds.
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<label
				class="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-700"
			>
				<input
					type="checkbox"
					class="rounded border-zinc-600 bg-zinc-950 text-emerald-500 focus:ring-emerald-500/40"
					bind:checked={unacknowledgedOnly}
				/>
				Unacknowledged only
			</label>
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100 disabled:opacity-50"
				disabled={eventsQuery.isFetching}
				onclick={() => eventsQuery.refetch()}
			>
				<RefreshCw class="size-3.5 {eventsQuery.isFetching ? 'animate-spin' : ''}" />
				Refresh
			</button>
		</div>
	</div>

	{#if actionError}
		<p
			class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300"
			role="alert"
		>
			{actionError}
		</p>
	{/if}

	{#if eventsQuery.isPending}
		<div class="flex min-h-[280px] items-center justify-center">
			<Spinner label="Loading events" />
		</div>
	{:else if eventsQuery.isError}
		<div
			class="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-12 text-center"
		>
			<p class="text-sm font-medium text-red-200">Could not load events</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{eventsQuery.error instanceof Error
					? eventsQuery.error.message
					: "Unknown error while loading events."}
			</p>
			<button
				type="button"
				class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
				onclick={() => eventsQuery.refetch()}
			>
				Retry
			</button>
		</div>
	{:else if events.length === 0}
		<EventsEmptyState {unacknowledgedOnly} />
	{:else}
		<div class="flex flex-col gap-3">
			{#each events as event (event.id)}
				<EventRow
					{event}
					acknowledging={acknowledgingId === event.id}
					onAcknowledge={handleAcknowledge}
				/>
			{/each}
		</div>
	{/if}
</section>
