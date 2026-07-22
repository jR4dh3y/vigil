<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { AlertCircle, Radio } from "lucide-svelte";
	import {
		type LiveCamera,
		liveKeys,
		liveRefetchInterval,
		requestLive,
	} from "$lib/live";
	import Spinner from "$lib/components/Spinner.svelte";
	import HlsPlayer from "./HlsPlayer.svelte";
	import WhepPlayer from "./WhepPlayer.svelte";

	// Prefer HLS first for home/LAN reliability (DVR + MediaMTX); fall back to WHEP.
	type PlaybackMode = "hls" | "whep" | "failed";

	type Props = {
		camera: LiveCamera;
	};

	let { camera }: Props = $props();

	let mode = $state<PlaybackMode>("hls");
	let playing = $state(false);
	let playerError = $state<string | null>(null);

	const liveQuery = createQuery(() => ({
		queryKey: liveKeys.stream(camera.id),
		queryFn: () => requestLive(camera.id),
		refetchInterval: (query) => liveRefetchInterval(query.state.data?.expiresAt),
		// Keep previous URLs briefly while a refresh is in flight.
		placeholderData: (prev) => prev,
	}));

	// Reset playback when the camera changes or a fresh session arrives.
	$effect(() => {
		void camera.id;
		void liveQuery.dataUpdatedAt;
		mode = "hls";
		playing = false;
		playerError = null;
	});

	const stream = $derived(liveQuery.data);
	const loading = $derived(liveQuery.isPending && !stream);
	const requestFailed = $derived(liveQuery.isError && !stream);

	function handleHlsError(error: Error) {
		playerError = error.message;
		playing = false;
		mode = "whep";
	}

	function handleWhepError(error: Error) {
		playerError = error.message;
		playing = false;
		mode = "failed";
	}

	function handlePlaying() {
		playing = true;
		playerError = null;
	}

	function retry() {
		mode = "hls";
		playing = false;
		playerError = null;
		void liveQuery.refetch();
	}
</script>

<article
	class="flex flex-col overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900/60 shadow-sm"
>
	<div class="relative aspect-video w-full bg-black">
		{#if loading}
			<div class="absolute inset-0 flex items-center justify-center bg-zinc-950">
				<Spinner label="Connecting" />
			</div>
		{:else if requestFailed}
			<div
				class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-zinc-950 px-4 text-center"
			>
				<AlertCircle class="size-6 text-red-400" />
				<p class="text-sm font-medium text-red-200">Could not start stream</p>
				<p class="max-w-xs text-xs text-red-300/80">
					{liveQuery.error instanceof Error
						? liveQuery.error.message
						: "Live session request failed."}
				</p>
				<button
					type="button"
					class="mt-1 rounded-md bg-zinc-800 px-3 py-1.5 text-xs text-zinc-100 hover:bg-zinc-700"
					onclick={retry}
				>
					Retry
				</button>
			</div>
		{:else if stream && mode === "failed"}
			<div
				class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-zinc-950 px-4 text-center"
			>
				<AlertCircle class="size-6 text-amber-400" />
				<p class="text-sm font-medium text-zinc-200">Playback unavailable</p>
				<p class="max-w-xs text-xs text-zinc-500">
					WebRTC (WHEP) and HLS both failed
					{#if playerError}
						— {playerError}
					{/if}
				</p>
				<button
					type="button"
					class="mt-1 rounded-md bg-zinc-800 px-3 py-1.5 text-xs text-zinc-100 hover:bg-zinc-700"
					onclick={retry}
				>
					Retry
				</button>
			</div>
		{:else if stream}
			{#key `${stream.cameraId}-${stream.token}-${mode}`}
				{#if mode === "hls"}
					<HlsPlayer
						hlsUrl={stream.hlsUrl}
						token={stream.token}
						onError={handleHlsError}
						onPlaying={handlePlaying}
					/>
				{:else if mode === "whep"}
					<WhepPlayer
						whepUrl={stream.whepUrl}
						token={stream.token}
						onError={handleWhepError}
						onPlaying={handlePlaying}
					/>
				{/if}
			{/key}

			{#if !playing && mode !== "failed"}
				<div
					class="pointer-events-none absolute inset-0 flex items-center justify-center bg-zinc-950/50"
				>
					<Spinner
						label={mode === "whep" ? "Trying WebRTC…" : "Starting live"}
					/>
				</div>
			{/if}
		{/if}
	</div>

	<div class="flex items-center justify-between gap-2 border-t border-zinc-800/80 px-3 py-2">
		<div class="min-w-0">
			<h2 class="truncate text-sm font-medium text-zinc-100">{camera.name}</h2>
			<p class="text-[11px] text-zinc-500">
				{#if mode === "hls" && playing}
					HLS
				{:else if mode === "whep" && playing}
					WebRTC
				{:else if mode === "failed"}
					Offline
				{:else}
					Connecting…
				{/if}
			</p>
		</div>
		<span
			class="inline-flex shrink-0 items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide
				{playing
				? 'bg-emerald-500/15 text-emerald-300'
				: mode === 'failed' || requestFailed
					? 'bg-red-500/15 text-red-300'
					: 'bg-zinc-800 text-zinc-500'}"
		>
			<Radio class="size-3 {playing ? 'animate-pulse' : ''}" />
			{playing ? "Live" : mode === "failed" || requestFailed ? "Error" : "…"}
		</span>
	</div>
</article>
