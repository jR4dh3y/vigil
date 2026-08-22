<script lang="ts">
	import { resolve } from "$app/paths";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";
	import { Film, LayoutGrid, Video, X } from "lucide-svelte";
	import { cameraKeys, listCameras } from "$lib/cameras";
	import PageActions from "$lib/components/PageActions.svelte";
	import PlaybackGrid from "$lib/components/live/PlaybackGrid.svelte";
	import LiveGrid from "$lib/components/live/LiveGrid.svelte";
	import PlaybackDock from "$lib/components/timeline/PlaybackDock.svelte";
	import Spinner from "$lib/components/Spinner.svelte";
	import type { LiveCamera } from "$lib/live";
	import {
		calendarGridRange,
		calendarMonthForDate,
		calendarMonthForValue,
		currentLocalDayRange,
		isFutureTime,
		listRecordingDays,
		listRecordings,
		localDateValue,
		rangeForLocalDate,
		recordingKeys,
		requestPlayback,
		toIso,
	} from "$lib/recordings";
	import type {
		CalendarMonth,
		PlaybackSession,
		RecordingDaySource,
		RecordingList,
	} from "$lib/recordings";

	const camerasQuery = createQuery(() => ({
		queryKey: cameraKeys.list(),
		queryFn: listCameras,
	}));

	let playbackEnabled = $state(false);
	let playbackStarted = $state(false);
	let range = $state(currentLocalDayRange());
	let archiveDate = $state(localDateValue(new Date()));
	let calendarMonth = $state(calendarMonthForDate(new Date()));
	let calendarTimeZone = $state("UTC");
	let selectedTime = $state<Date | null>(null);
	let sessions = new SvelteMap<string, PlaybackSession>();
	let playbackErrors = new SvelteMap<string, string>();
	let latestRequestId = 0;
	const advancingTokens = new SvelteSet<string>();

	const enabledCameras = $derived(
		(camerasQuery.data ?? [])
			.filter((camera) => camera.enabled)
			.map((camera): LiveCamera => ({ id: camera.id, name: camera.name })),
	);
	const cameraIds = $derived(enabledCameras.map((camera) => camera.id));
	const fromIso = $derived(toIso(range.from));
	const toIsoString = $derived(toIso(range.to));
	const maxArchiveDate = $derived(localDateValue(new Date()));
	const calendarRange = $derived(calendarGridRange(calendarMonth));
	const calendarFromIso = $derived(toIso(calendarRange.from));
	const calendarToIso = $derived(toIso(calendarRange.to));

	async function loadRecordings(): Promise<SvelteMap<string, RecordingList>> {
		const entries = await Promise.all(
			enabledCameras.map(async (camera) => {
				const recordings = await listRecordings(camera.id, fromIso, toIsoString);
				return [camera.id, recordings] as const;
			}),
		);
		return new SvelteMap(entries);
	}

	const recordingsQuery = createQuery(() => ({
		queryKey: recordingKeys.listMany(cameraIds, fromIso, toIsoString),
		queryFn: loadRecordings,
		enabled: playbackEnabled && cameraIds.length > 0,
	}));

	const recordingDaysQuery = createQuery(() => ({
		queryKey: recordingKeys.days(
			cameraIds,
			calendarFromIso,
			calendarToIso,
			calendarTimeZone,
		),
		queryFn: () =>
			listRecordingDays(calendarFromIso, calendarToIso, calendarTimeZone),
		enabled: playbackEnabled && cameraIds.length > 0,
	}));

	const playbackMutation = createMutation(() => ({
		mutationFn: async ({
			start,
			requestId,
		}: {
			start: string;
			requestId: number;
		}) => {
			const results = await Promise.all(
				enabledCameras.map(async (camera) => {
					try {
						const session = await requestPlayback(camera.id, start);
						return { cameraId: camera.id, session };
					} catch (error: unknown) {
						return {
							cameraId: camera.id,
							error: error instanceof Error ? error.message : "Playback failed",
						};
					}
				}),
			);

			const nextSessions = new SvelteMap<string, PlaybackSession>();
			const nextErrors = new SvelteMap<string, string>();
			for (const result of results) {
				if ("error" in result) {
					nextErrors.set(result.cameraId, result.error ?? "Playback failed");
				} else {
					nextSessions.set(result.cameraId, result.session);
				}
			}

			return { requestId, sessions: nextSessions, errors: nextErrors };
		},
		onSuccess: (result) => {
			if (!playbackEnabled || result.requestId !== latestRequestId) {
				return;
			}
			replaceMap(sessions, result.sessions);
			replaceMap(playbackErrors, result.errors);
		},
	}));

	const recordings = $derived(
		recordingsQuery.data ?? new SvelteMap<string, RecordingList>(),
	);
	const coverageTracks = $derived(
		enabledCameras.map((camera, index) => ({
			id: camera.id,
			channel: index + 1,
			name: camera.name,
			coverage: recordings.get(camera.id)?.coverage ?? [],
		})),
	);
	const recordingsError = $derived(
		recordingsQuery.isError
			? recordingsQuery.error instanceof Error
				? recordingsQuery.error.message
				: "Could not load recording coverage"
			: null,
	);
	const playbackLoading = $derived(playbackMutation.isPending);
	const playbackDisabled = $derived(recordingsQuery.isPending || playbackMutation.isPending);
	const calendarAvailability = $derived(
		new SvelteMap<string, RecordingDaySource>(
			(recordingDaysQuery.data?.days ?? []).map((day) => [day.date, day.source]),
		),
	);

	function replaceMap<T>(target: SvelteMap<string, T>, source: ReadonlyMap<string, T>) {
		target.clear();
		for (const [key, value] of source) {
			target.set(key, value);
		}
	}

	function resetPlaybackView() {
		latestRequestId += 1;
		playbackStarted = false;
		selectedTime = null;
		sessions.clear();
		playbackErrors.clear();
		advancingTokens.clear();
	}

	function togglePlayback() {
		const enabled = !playbackEnabled;
		if (enabled) {
			calendarTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
		}
		playbackEnabled = enabled;
		resetPlaybackView();
	}

	function setArchiveDate(value: string) {
		const selectedRange = rangeForLocalDate(value);
		if (!selectedRange) {
			return;
		}
		archiveDate = value;
		calendarMonth = calendarMonthForValue(value) ?? calendarMonth;
		range = selectedRange;
		resetPlaybackView();
	}

	function setCalendarMonth(month: CalendarMonth) {
		calendarMonth = month;
	}

	function handleSeek(time: Date) {
		if (!playbackEnabled || enabledCameras.length === 0 || isFutureTime(time)) {
			return;
		}
		selectedTime = time;
		playbackStarted = true;
		playbackErrors.clear();
		const requestId = latestRequestId + 1;
		latestRequestId = requestId;
		void playbackMutation.mutateAsync({ start: toIso(time), requestId });
	}

	async function handlePlaybackEnded(endedSession: PlaybackSession) {
		const nextStart = endedSession.nextRecordingStart;
		if (
			!playbackEnabled ||
			endedSession.source !== "gdrive" ||
			!nextStart ||
			advancingTokens.has(endedSession.token) ||
			sessions.get(endedSession.cameraId)?.token !== endedSession.token
		) {
			return;
		}

		advancingTokens.add(endedSession.token);
		playbackErrors.delete(endedSession.cameraId);
		try {
			const nextSession = await requestPlayback(endedSession.cameraId, nextStart);
			if (
				playbackEnabled &&
				sessions.get(endedSession.cameraId)?.token === endedSession.token
			) {
				sessions.set(endedSession.cameraId, nextSession);
			}
		} catch (error: unknown) {
			if (sessions.get(endedSession.cameraId)?.token === endedSession.token) {
				playbackErrors.set(
					endedSession.cameraId,
					error instanceof Error ? error.message : "Playback failed",
				);
			}
		} finally {
			advancingTokens.delete(endedSession.token);
		}
	}
</script>

<svelte:head>
	<title>{playbackEnabled ? "Playback" : "Live"} · Vigil</title>
</svelte:head>

<div class="h-full w-full">
	{#if camerasQuery.isSuccess && enabledCameras.length > 0}
		<PageActions>
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-sm transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-500/70
					{playbackEnabled
					? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300 hover:bg-emerald-500/15'
					: 'border-zinc-800 bg-zinc-900 text-zinc-300 hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100'}"
				aria-pressed={playbackEnabled}
				onclick={togglePlayback}
			>
				{#if playbackEnabled}
					<X class="size-3.5" />
					<span>Close playback</span>
				{:else}
					<Film class="size-3.5" />
					<span>Playback</span>
				{/if}
			</button>
		</PageActions>
	{/if}

	{#if camerasQuery.isPending}
		<div class="flex h-full items-center justify-center">
			<Spinner label="Loading cameras" />
		</div>
	{:else if camerasQuery.isError}
		<div class="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
			<p class="text-sm font-medium text-red-200">Could not load cameras</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{camerasQuery.error instanceof Error
					? camerasQuery.error.message
					: "Unknown error while loading cameras."}
			</p>
			<button
				type="button"
				class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
				onclick={() => camerasQuery.refetch()}
			>
				Retry
			</button>
		</div>
	{:else if enabledCameras.length === 0}
		<div class="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
			<span
				class="flex size-12 items-center justify-center rounded-xl bg-zinc-800/80 text-zinc-400"
			>
				{#if (camerasQuery.data ?? []).length === 0}
					<LayoutGrid class="size-6" />
				{:else}
					<Video class="size-6" />
				{/if}
			</span>
			{#if (camerasQuery.data ?? []).length === 0}
				<p class="text-sm font-medium text-zinc-300">No cameras yet</p>
				<p class="max-w-sm text-sm text-zinc-500">
					Add cameras from the
					<a
						href={resolve("/cameras")}
						class="text-emerald-400 no-underline hover:text-emerald-300">Cameras</a
					>
					page, then enable them for live view.
				</p>
			{:else}
				<p class="text-sm font-medium text-zinc-300">No enabled cameras</p>
				<p class="max-w-sm text-sm text-zinc-500">
					You have cameras configured, but none are enabled. Turn one on from
					<a
						href={resolve("/cameras")}
						class="text-emerald-400 no-underline hover:text-emerald-300">Cameras</a
					>
					to see the live grid.
				</p>
			{/if}
		</div>
	{:else}
		<div class="flex h-full min-h-0 flex-col">
			<div class="min-h-0 flex-1">
				{#if playbackEnabled && playbackStarted}
					<PlaybackGrid
						cameras={enabledCameras}
						{sessions}
						loading={playbackLoading}
						errors={playbackErrors}
						onEnded={handlePlaybackEnded}
					/>
				{:else}
					<LiveGrid cameras={enabledCameras} />
				{/if}
			</div>

			{#if playbackEnabled}
				<PlaybackDock
					tracks={coverageTracks}
					from={range.from}
					to={range.to}
					{selectedTime}
					{archiveDate}
					{maxArchiveDate}
					{calendarMonth}
					{calendarAvailability}
					calendarLoading={recordingDaysQuery.isPending}
					calendarError={recordingDaysQuery.isError}
					loading={recordingsQuery.isPending}
					error={recordingsError}
					disabled={playbackDisabled}
					onArchiveDateChange={setArchiveDate}
					onCalendarMonthChange={setCalendarMonth}
					onSeek={handleSeek}
					onRetry={() => recordingsQuery.refetch()}
				/>
			{/if}
		</div>
	{/if}
</div>
