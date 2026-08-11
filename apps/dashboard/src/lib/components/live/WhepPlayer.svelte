<script lang="ts">
	import { untrack } from "svelte";
	import { streamEndpoint, withStreamToken } from "$lib/live";

	type Props = {
		whepUrl: string;
		token: string;
		class?: string;
		onError?: (error: Error) => void;
		onPlaying?: () => void;
	};

	let {
		whepUrl,
		token,
		class: className = "object-contain",
		onError,
		onPlaying,
	}: Props = $props();

	let videoEl = $state<HTMLVideoElement | null>(null);
	const endpoint = $derived(streamEndpoint(whepUrl));

	function waitForIceGathering(pc: RTCPeerConnection): Promise<void> {
		if (pc.iceGatheringState === "complete") {
			return Promise.resolve();
		}
		return new Promise((resolve) => {
			const check = () => {
				if (pc.iceGatheringState === "complete") {
					pc.removeEventListener("icegatheringstatechange", check);
					resolve();
				}
			};
			pc.addEventListener("icegatheringstatechange", check);
			// Safety timeout so we don't hang forever on broken ICE.
			window.setTimeout(() => {
				pc.removeEventListener("icegatheringstatechange", check);
				resolve();
			}, 5_000);
		});
	}

	$effect(() => {
		const url = withStreamToken(endpoint, untrack(() => token));
		const el = videoEl;
		if (!el) {
			return;
		}

		let cancelled = false;
		let pc: RTCPeerConnection | null = null;
		let resourceUrl: string | null = null;

		async function connect() {
			try {
				pc = new RTCPeerConnection({
					iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
				});

				pc.addTransceiver("video", { direction: "recvonly" });
				pc.addTransceiver("audio", { direction: "recvonly" });

				pc.ontrack = (event) => {
					if (cancelled || !el) {
						return;
					}
					const [stream] = event.streams;
					if (stream) {
						el.srcObject = stream;
					} else {
						el.srcObject = new MediaStream([event.track]);
					}
				};

				pc.onconnectionstatechange = () => {
					if (!pc || cancelled) {
						return;
					}
					if (pc.connectionState === "failed" || pc.connectionState === "disconnected") {
						onError?.(new Error(`WebRTC connection ${pc.connectionState}`));
					}
				};

				const offer = await pc.createOffer();
				await pc.setLocalDescription(offer);
				await waitForIceGathering(pc);

				if (cancelled || !pc.localDescription?.sdp) {
					return;
				}

				const response = await fetch(url, {
					method: "POST",
					headers: {
						Accept: "application/sdp",
						"Content-Type": "application/sdp",
					},
					body: pc.localDescription.sdp,
				});

				if (cancelled) {
					return;
				}

				if (!response.ok) {
					throw new Error(`WHEP request failed (${response.status})`);
				}

				const answerSdp = await response.text();
				if (!answerSdp.trim()) {
					throw new Error("Empty WHEP SDP answer");
				}

				const location = response.headers.get("Location");
				if (location) {
					try {
						resourceUrl = new URL(location, url).toString();
					} catch {
						resourceUrl = location;
					}
				}

				await pc.setRemoteDescription({ type: "answer", sdp: answerSdp });
			} catch (err) {
				if (cancelled) {
					return;
				}
				const error = err instanceof Error ? err : new Error("WHEP playback failed");
				onError?.(error);
			}
		}

		void connect();

		return () => {
			cancelled = true;
			if (el) {
				el.srcObject = null;
			}
			if (pc) {
				pc.ontrack = null;
				pc.onconnectionstatechange = null;
				pc.close();
				pc = null;
			}
			// Best-effort WHEP resource teardown (MediaMTX session).
			if (resourceUrl) {
				void fetch(resourceUrl, { method: "DELETE" }).catch(() => {
					/* ignore cleanup errors */
				});
			}
		};
	});
</script>

<video
	bind:this={videoEl}
	class="h-full w-full bg-black {className}"
	autoplay
	playsinline
	muted
	onplaying={() => onPlaying?.()}
></video>
