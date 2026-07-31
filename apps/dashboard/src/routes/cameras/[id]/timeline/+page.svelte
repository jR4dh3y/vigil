<script lang="ts">
	import { page } from "$app/state";
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import { Film } from "lucide-svelte";
	import { cameraKeys, getCamera } from "$lib/cameras";
	import CameraContextBar from "$lib/components/cameras/CameraContextBar.svelte";
	import CameraStatusBadge from "$lib/components/cameras/CameraStatusBadge.svelte";
	import Spinner from "$lib/components/Spinner.svelte";
	import CoverageTimeline from "$lib/components/timeline/CoverageTimeline.svelte";
	import PlaybackPlayer from "$lib/components/timeline/PlaybackPlayer.svelte";
	import RangePresetButtons from "$lib/components/timeline/RangePresetButtons.svelte";
	import type { PlaybackSession, RangePreset } from "$lib/recordings";
	import {
		defaultTimeRange,
		formatSelectedTime,
		listRecordings,
		rangeForPreset,
		RecordingApiError,
		recordingKeys,
		requestPlayback,
		toIso,
	} from "$lib/recordings";

	const cameraId = $derived(page.params.id ?? "");

	let preset = $state<RangePreset>("24h");
	let range = $state(defaultTimeRange());
	let selectedTime = $state<Date | null>(null);
	let session = $state<PlaybackSession | null>(null);
	let playerError = $state<string | null>(null);

	const fromIso = $derived(toIso(range.from));
	const toIsoStr = $derived(toIso(range.to));

	const cameraQuery = createQuery(() => ({
		queryKey: cameraKeys.detail(cameraId),
		queryFn: () => getCamera(cameraId),
		enabled: Boolean(cameraId),
	}));

	const recordingsQuery = createQuery(() => ({
		queryKey: recordingKeys.list(cameraId, fromIso, toIsoStr),
		queryFn: () => listRecordings(cameraId, fromIso, toIsoStr),
		enabled: Boolean(cameraId),
	}));

	const playbackMutation = createMutation(() => ({
		mutationFn: ({ id, start }: { id: string; start: string }) => requestPlayback(id, start),
		onSuccess: (data) => {
			session = data;
			playerError = null;
		},
		onError: (error: unknown) => {
			session = null;
			if (error instanceof RecordingApiError) {
				playerError = error.message;
				return;
			}
			playerError = error instanceof Error ? error.message : "Failed to start playback";
		},
	}));

	const camera = $derived(cameraQuery.data);
	const coverage = $derived(recordingsQuery.data?.coverage ?? []);
	const hasCoverage = $derived(coverage.length > 0);
	const recordingsLoading = $derived(recordingsQuery.isPending);
	const recordingsEmpty = $derived(
		recordingsQuery.isSuccess && !hasCoverage && (recordingsQuery.data?.recordings.length ?? 0) === 0,
	);

	function setPreset(next: RangePreset) {
		preset = next;
		range = rangeForPreset(next);
		selectedTime = null;
		session = null;
		playerError = null;
	}

	function handlePreview(time: Date) {
		selectedTime = time;
	}

	function handleSeek(time: Date) {
		if (!cameraId) {
			return;
		}
		selectedTime = time;
		playerError = null;
		void playbackMutation.mutateAsync({
			id: cameraId,
			start: toIso(time),
		});
	}

	function handlePlayerError(error: Error) {
		playerError = error.message;
	}
</script>

<svelte:head>
	<title>{camera ? `${camera.name} · Timeline` : "Timeline"} · NVR</title>
</svelte:head>

<section class="mx-auto flex w-full max-w-5xl flex-col gap-6">
	{#if cameraQuery.isPending}
		<div class="flex min-h-[280px] items-center justify-center">
			<Spinner label="Loading camera" />
		</div>
	{:else if cameraQuery.isError}
		<div
			class="flex min-h-[280px] flex-col items-center justify-center gap-3 rounded-xl border border-red-500/20 bg-red-500/5 px-6 py-12 text-center"
		>
			<p class="text-sm font-medium text-red-200">Could not load camera</p>
			<p class="max-w-sm text-sm text-red-300/80">
				{cameraQuery.error instanceof Error
					? cameraQuery.error.message
					: "Unknown error while loading camera."}
			</p>
			<div class="flex gap-2">
				<a
					href="/cameras"
					class="rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm text-zinc-200 no-underline hover:bg-zinc-800"
				>
					Back to list
				</a>
				<button
					type="button"
					class="rounded-lg bg-zinc-800 px-4 py-2 text-sm text-zinc-100 hover:bg-zinc-700"
					onclick={() => cameraQuery.refetch()}
				>
					Retry
				</button>
			</div>
		</div>
	{:else if camera}
		<CameraContextBar {camera}>
			{#snippet actions()}
			<a
				href="/cameras/{camera.id}"
				class="inline-flex items-center rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 no-underline transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100"
			>
				Settings
			</a>
			<a
				href="/"
				class="inline-flex items-center rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1.5 text-sm text-zinc-300 no-underline transition-colors hover:border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100"
			>
				Live
			</a>
			{/snippet}
		</CameraContextBar>

		<div class="flex flex-wrap items-center gap-2">
			<CameraStatusBadge status={camera.status} />
			<span class="text-sm text-zinc-500">Recorded playback</span>
		</div>

		<div class="flex flex-col gap-4 rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 sm:p-5">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<div class="flex flex-col gap-0.5">
					<span class="text-sm font-medium text-zinc-200">Timeline</span>
					<span class="text-xs text-zinc-500">
						{#if selectedTime}
							Seek: {formatSelectedTime(selectedTime)}
						{:else}
							Select a time on the coverage bar
						{/if}
					</span>
				</div>
				<RangePresetButtons value={preset} onChange={setPreset} />
			</div>

			{#if recordingsLoading}
				<div class="flex min-h-[72px] items-center justify-center py-4">
					<Spinner label="Loading recordings" />
				</div>
			{:else if recordingsQuery.isError}
				<div
					class="flex flex-col items-center justify-center gap-2 rounded-lg border border-red-500/20 bg-red-500/5 px-4 py-8 text-center"
				>
					<p class="text-sm font-medium text-red-200">Could not load recordings</p>
					<p class="max-w-sm text-xs text-red-300/80">
						{recordingsQuery.error instanceof Error
							? recordingsQuery.error.message
							: "Unknown error while loading recordings."}
					</p>
					<button
						type="button"
						class="mt-1 rounded-md bg-zinc-800 px-3 py-1.5 text-xs text-zinc-100 hover:bg-zinc-700"
						onclick={() => recordingsQuery.refetch()}
					>
						Retry
					</button>
				</div>
			{:else if recordingsEmpty}
				<div
					class="flex flex-col items-center justify-center gap-2 rounded-lg border border-dashed border-zinc-800 bg-zinc-950/40 px-4 py-10 text-center"
				>
					<span
						class="flex size-10 items-center justify-center rounded-lg bg-zinc-800/80 text-zinc-500"
					>
						<Film class="size-5" />
					</span>
					<p class="text-sm font-medium text-zinc-300">No recordings in range</p>
					<p class="max-w-sm text-xs text-zinc-500">
						Nothing was recorded for this camera in the selected window. Try a wider range.
					</p>
				</div>
			{:else}
				<CoverageTimeline
					{coverage}
					from={range.from}
					to={range.to}
					{selectedTime}
					disabled={playbackMutation.isPending}
					onPreview={handlePreview}
					onSeek={handleSeek}
				/>
				{#if !hasCoverage && (recordingsQuery.data?.recordings.length ?? 0) > 0}
					<p class="text-xs text-zinc-500">
						{recordingsQuery.data?.recordings.length} segment(s) listed; no coverage bars returned.
					</p>
				{/if}
			{/if}
		</div>

		<div class="flex flex-col gap-2">
			<div class="flex items-center justify-between gap-2">
				<h2 class="text-sm font-medium text-zinc-200">Player</h2>
				{#if playbackMutation.isPending}
					<span class="text-xs text-zinc-500">Requesting session…</span>
				{/if}
			</div>
			<PlaybackPlayer
				{session}
				loading={playbackMutation.isPending && !session}
				error={playerError}
				onError={handlePlayerError}
			/>
		</div>
	{/if}
</section>
