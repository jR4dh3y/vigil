import { fetch } from "expo/fetch";
import { type ComponentType, useEffect, useEffectEvent, useState } from "react";
import type {
	RTCPeerConnection as RTCPeerConnectionType,
	RTCVideoViewProps,
} from "react-native-webrtc";
import { streamEndpoint } from "@/lib/api/config";

type WhepStreamState = {
	streamUrl: string | null;
	RtcView: ComponentType<RTCVideoViewProps> | null;
	failed: boolean;
};

const initialState: WhepStreamState = {
	streamUrl: null,
	RtcView: null,
	failed: false,
};

export function useWhepStream(uri: string): WhepStreamState {
	const [state, setState] = useState<WhepStreamState>(initialState);
	const endpoint = streamEndpoint(uri);
	const getConnectionUri = useEffectEvent(() => uri);

	useEffect(() => {
		void endpoint;
		const connectionUri = getConnectionUri();
		let cancelled = false;
		let peer: RTCPeerConnectionType | null = null;
		let resourceUrl: string | null = null;
		const fallbackTimer = setTimeout(() => {
			peer?.close();
			setState((current) => (current.streamUrl ? current : { ...current, failed: true }));
		}, 8_000);

		const connect = async () => {
			try {
				const webRtc = await import("react-native-webrtc");
				if (cancelled) {
					return;
				}

				// No public STUN servers: recorders are reached over the local network,
				// where host candidates suffice and streaming must not depend on (or
				// leak to) third-party infrastructure.
				peer = new webRtc.RTCPeerConnection({ iceServers: [] });
				peer.addTransceiver("video", { direction: "recvonly" });
				peer.addTransceiver("audio", { direction: "recvonly" });
				peer.ontrack = (event: unknown) => {
					const streamUrl = getTrackStreamUrl(event);
					if (!cancelled && streamUrl) {
						clearTimeout(fallbackTimer);
						setState({
							streamUrl,
							RtcView: webRtc.RTCView,
							failed: false,
						});
					}
				};
				peer.onconnectionstatechange = () => {
					// Only "failed" is terminal; "disconnected" is often transient and
					// can recover on its own, so it must not trigger the HLS fallback.
					if (!cancelled && peer && peer.connectionState === "failed") {
						peer.close();
						setState((current) => ({ ...current, failed: true }));
					}
				};

				const offer = await peer.createOffer();
				await peer.setLocalDescription(offer);
				await waitForIceGathering(peer);
				const sdp = peer.localDescription?.sdp;
				if (cancelled || !sdp) {
					return;
				}

				const response = await fetch(connectionUri, {
					method: "POST",
					headers: {
						Accept: "application/sdp",
						"Content-Type": "application/sdp",
					},
					body: sdp,
				});
				if (!response.ok) {
					throw new Error(`WHEP request failed (${response.status})`);
				}
				const answer = await response.text();
				if (!answer.trim()) {
					throw new Error("WHEP returned an empty response");
				}

				const location = response.headers.get("Location");
				if (location) {
					const resolvedUrl = new URL(location, connectionUri);
					const baseUri = new URL(connectionUri);
					for (const [key, value] of baseUri.searchParams) {
						if (!resolvedUrl.searchParams.has(key)) {
							resolvedUrl.searchParams.set(key, value);
						}
					}
					resourceUrl = resolvedUrl.toString();
				} else {
					resourceUrl = null;
				}
				await peer.setRemoteDescription({ type: "answer", sdp: answer });
			} catch (cause) {
				peer?.close();
				peer = null;
				throw cause;
			}
		};

		setState(initialState);
		// One bounded retry: a scroll-flap or transient negotiation error must not
		// permanently demote a tile to the HLS fallback.
		void (async () => {
			for (let attempt = 0; ; attempt += 1) {
				try {
					await connect();
					return;
				} catch {
					if (cancelled || attempt >= 1) {
						break;
					}
					const { promise, resolve } = Promise.withResolvers<void>();
					setTimeout(resolve, 700);
					await promise;
				}
			}
			if (!cancelled) {
				setState((current) => ({ ...current, failed: true }));
			}
		})();
		return () => {
			cancelled = true;
			clearTimeout(fallbackTimer);
			if (peer) {
				peer.ontrack = null;
				peer.onconnectionstatechange = null;
			}
			peer?.close();
			if (resourceUrl) {
				const cleanupUrl = resourceUrl;
				void fetch(cleanupUrl, { method: "DELETE" }).catch(() => undefined);
			}
		};
	}, [endpoint]);

	return state;
}

function waitForIceGathering(peer: RTCPeerConnectionType): Promise<void> {
	if (peer.iceGatheringState === "complete") {
		return Promise.resolve();
	}
	return new Promise((resolve) => {
		let timer: ReturnType<typeof setTimeout>;
		const finish = () => {
			clearTimeout(timer);
			peer.onicegatheringstatechange = null;
			resolve();
		};
		const checkState = () => {
			if (peer.iceGatheringState === "complete") {
				finish();
			}
		};
		peer.onicegatheringstatechange = checkState;
		timer = setTimeout(finish, 5_000);
	});
}

function getTrackStreamUrl(event: unknown): string | null {
	if (!isRecord(event) || !Array.isArray(event.streams)) {
		return null;
	}
	const stream: unknown = event.streams[0];
	if (!isRecord(stream) || typeof stream.toURL !== "function") {
		return null;
	}
	const value: unknown = stream.toURL();
	return typeof value === "string" ? value : null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}
