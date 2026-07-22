<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { AlertCircle } from "lucide-svelte";
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
		channel: number;
	};

	let { camera, channel }: Props = $props();

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

<article class="group relative min-h-0 min-w-0 bg-black">
	<div class="absolute inset-0">
		{#if loading}
			<div class="absolute inset-0 flex items-center justify-center bg-zinc-950">
				<Spinner label="Connecting" />
			</div>
		{:else if requestFailed}
			<div
				class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-zinc-950 px-3 text-center"
			>
				<AlertCircle class="size-5 text-red-400" />
				<p class="text-xs text-red-200">Stream failed</p>
				<button
					type="button"
					class="rounded-md bg-zinc-800 px-2.5 py-1 text-xs text-zinc-100 hover:bg-zinc-700"
					onclick={retry}
				>
					Retry
				</button>
			</div>
		{:else if stream && mode === "failed"}
			<div
				class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-zinc-950 px-3 text-center"
			>
				<AlertCircle class="size-5 text-amber-400" />
				<p class="text-xs text-zinc-300">Unavailable</p>
				{#if playerError}
					<p class="max-w-[90%] truncate text-[10px] text-zinc-500">{playerError}</p>
				{/if}
				<button
					type="button"
					class="rounded-md bg-zinc-800 px-2.5 py-1 text-xs text-zinc-100 hover:bg-zinc-700"
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
						class="object-cover"
						onError={handleHlsError}
						onPlaying={handlePlaying}
					/>
				{:else if mode === "whep"}
					<WhepPlayer
						whepUrl={stream.whepUrl}
						token={stream.token}
						class="object-cover"
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

	<!-- Channel label: hover only -->
	<div
		class="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent px-2.5 py-2 opacity-0 transition-opacity group-hover:opacity-100"
	>
		<p class="truncate text-xs font-medium text-white drop-shadow">
			CH{channel} · {camera.name}
		</p>
	</div>
</article>
