<script lang="ts">
	import { AlertCircle } from "lucide-svelte";
	import type { PlaybackSession } from "$lib/recordings";
	import Spinner from "$lib/components/Spinner.svelte";

	type Props = {
		session: PlaybackSession | null;
		loading?: boolean;
		error?: string | null;
		class?: string;
		videoClass?: string;
		onError?: (error: Error) => void;
		onPlaying?: () => void;
		onEnded?: (session: PlaybackSession) => void;
	};

	let {
		session,
		loading = false,
		error = null,
		class: className = "",
		videoClass = "object-contain",
		onError,
		onPlaying,
		onEnded,
	}: Props = $props();

	let playing = $state(false);

	function handleLoadStart() {
		playing = false;
	}

	function handlePlaying() {
		playing = true;
		onPlaying?.();
	}

	function handleError(err: Error) {
		playing = false;
		onError?.(err);
	}

	function handleVideoError(event: Event) {
		if (!(event.currentTarget instanceof HTMLVideoElement)) {
			return;
		}
		const video = event.currentTarget;
		const detail = video.error?.message?.trim();
		handleError(new Error(detail || "Recorded MP4 playback failed"));
	}

	function handleLoadedMetadata(event: Event) {
		if (!(event.currentTarget instanceof HTMLVideoElement)) {
			return;
		}
		const video = event.currentTarget;
		const offset = session?.source === "gdrive" ? session.startOffsetSec : 0;
		if (offset > 0 && Number.isFinite(offset)) {
			video.currentTime = Math.min(offset, Math.max(0, video.duration || offset));
		}
	}

	function handleEnded() {
		playing = false;
		if (session) {
			onEnded?.(session);
		}
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
			<video
				src={session.playbackUrl}
				class="h-full w-full bg-black {videoClass}"
				controls
				autoplay
				playsinline
				muted
				preload="auto"
				onplaying={handlePlaying}
				onloadstart={handleLoadStart}
				onloadedmetadata={handleLoadedMetadata}
				onended={handleEnded}
				onerror={handleVideoError}
			></video>
		{/key}
		<span
			class="pointer-events-none absolute top-2 right-2 rounded-md border border-zinc-700/80 bg-zinc-950/80 px-2 py-1 text-[10px] font-medium tracking-wide text-zinc-300 uppercase"
		>
			{session.source === "gdrive" ? "Google Drive" : "Local recording"}
		</span>
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
