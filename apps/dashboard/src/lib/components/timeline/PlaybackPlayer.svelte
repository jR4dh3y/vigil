<script lang="ts">
	import { AlertCircle } from "lucide-svelte";
	import type { PlaybackSession } from "$lib/recordings";
	import Spinner from "$lib/components/Spinner.svelte";
	import HlsPlayer from "$lib/components/live/HlsPlayer.svelte";

	type Props = {
		session: PlaybackSession | null;
		loading?: boolean;
		error?: string | null;
		class?: string;
		videoClass?: string;
		onError?: (error: Error) => void;
		onPlaying?: () => void;
	};

	let {
		session,
		loading = false,
		error = null,
		class: className = "",
		videoClass = "object-contain",
		onError,
		onPlaying,
	}: Props = $props();

	let playing = $state(false);

	$effect(() => {
		void session?.token;
		void session?.playbackUrl;
		playing = false;
	});

	function handlePlaying() {
		playing = true;
		onPlaying?.();
	}

	function handleError(err: Error) {
		playing = false;
		onError?.(err);
	}
</script>

<div
	class="relative aspect-video w-full overflow-hidden rounded-xl border border-zinc-800 bg-black {className}"
>
	{#if loading}
		<div class="absolute inset-0 flex items-center justify-center bg-zinc-950">
			<Spinner label="Starting playback" />
		</div>
	{:else if error}
		<div
			class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-zinc-950 px-4 text-center"
		>
			<AlertCircle class="size-6 text-red-400" />
			<p class="text-sm font-medium text-red-200">Playback failed</p>
			<p class="max-w-sm text-xs text-red-300/80">{error}</p>
		</div>
	{:else if session}
		{#key `${session.cameraId}-${session.token}-${session.playbackUrl}`}
			<HlsPlayer
				hlsUrl={session.playbackUrl}
				token={session.token}
				class={videoClass}
				controls
				onError={handleError}
				onPlaying={handlePlaying}
			/>
		{/key}
		{#if !playing}
			<div
				class="pointer-events-none absolute inset-0 flex items-center justify-center bg-zinc-950/40"
			>
				<Spinner label="Loading video" />
			</div>
		{/if}
	{:else}
		<div
			class="absolute inset-0 flex flex-col items-center justify-center gap-1 bg-zinc-950 px-4 text-center"
		>
			<p class="text-sm font-medium text-zinc-400">No playback selected</p>
			<p class="max-w-xs text-xs text-zinc-600">
				Click the timeline to seek and start recorded playback.
			</p>
		</div>
	{/if}
</div>
