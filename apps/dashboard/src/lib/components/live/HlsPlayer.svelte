<script lang="ts">
	import { untrack } from "svelte";
	import Hls from "hls.js";
	import { streamEndpoint, withStreamToken } from "$lib/live";

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
	const endpoint = $derived(streamEndpoint(hlsUrl));

	$effect(() => {
		const el = videoEl;
		if (!el) {
			return;
		}
		const hlsSupported = Hls.isSupported();
		const url = hlsSupported
			? withStreamToken(endpoint, untrack(() => token))
			: withStreamToken(hlsUrl, token);

		let cancelled = false;
		let hls: Hls | null = null;

		function fail(message: string) {
			if (!cancelled) {
				onError?.(new Error(message));
			}
		}

		if (hlsSupported) {
			hls = new Hls({
				enableWorker: true,
				// MediaMTX low-latency HLS; slightly more tolerant for DVR jitter.
				lowLatencyMode: true,
				// First frame sooner when source just came on-demand.
				maxBufferLength: 12,
				manifestLoadingMaxRetry: 4,
				levelLoadingMaxRetry: 4,
				fragLoadingMaxRetry: 4,
				// Refreshing a short-lived token must not recreate the player. Apply
				// the latest token to every playlist and segment request instead.
				xhrSetup: (xhr, requestUrl) => {
					xhr.open("GET", withStreamToken(requestUrl, token), true);
				},
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
