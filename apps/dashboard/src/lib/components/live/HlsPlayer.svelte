<script lang="ts">
	import Hls from "hls.js";
	import { withStreamToken } from "$lib/live";

	type Props = {
		hlsUrl: string;
		token: string;
		class?: string;
		/** Native browser controls (useful for recording playback). Off for live mosaic. */
		controls?: boolean;
		onError?: (error: Error) => void;
		onPlaying?: () => void;
	};

	let {
		hlsUrl,
		token,
		class: className = "object-contain",
		controls = false,
		onError,
		onPlaying,
	}: Props = $props();

	let videoEl = $state<HTMLVideoElement | null>(null);

	$effect(() => {
		const url = withStreamToken(hlsUrl, token);
		const el = videoEl;
		if (!el) {
			return;
		}

		let cancelled = false;
		let hls: Hls | null = null;

		function fail(message: string) {
			if (!cancelled) {
				onError?.(new Error(message));
			}
		}

		if (Hls.isSupported()) {
			hls = new Hls({
				enableWorker: true,
				// MediaMTX low-latency HLS; slightly more tolerant for DVR jitter.
				lowLatencyMode: true,
				// First frame sooner when source just came on-demand.
				maxBufferLength: 12,
				manifestLoadingMaxRetry: 4,
				levelLoadingMaxRetry: 4,
				fragLoadingMaxRetry: 4,
			});
			hls.loadSource(url);
			hls.attachMedia(el);
			hls.on(Hls.Events.MANIFEST_PARSED, () => {
				void el.play().catch(() => {
					/* autoplay may be blocked until muted — video is muted */
				});
			});
			hls.on(Hls.Events.ERROR, (_event, data) => {
				if (!data.fatal || cancelled) {
					return;
				}
				if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
					hls?.startLoad();
					return;
				}
				if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
					hls?.recoverMediaError();
					return;
				}
				fail(data.error?.message ?? "HLS playback failed");
			});
		} else if (el.canPlayType("application/vnd.apple.mpegurl")) {
			el.src = url;
			const onNativeError = () => fail("Native HLS playback failed");
			el.addEventListener("error", onNativeError);
			return () => {
				cancelled = true;
				el.removeEventListener("error", onNativeError);
				el.removeAttribute("src");
				el.load();
			};
		} else {
			fail("HLS is not supported in this browser");
		}

		return () => {
			cancelled = true;
			if (hls) {
				hls.destroy();
				hls = null;
			}
			el.removeAttribute("src");
			el.load();
		};
	});
</script>

<video
	bind:this={videoEl}
	class="h-full w-full bg-black {className}"
	autoplay
	playsinline
	muted
	{controls}
	onplaying={() => onPlaying?.()}
></video>
