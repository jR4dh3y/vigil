import { fetch } from "expo/fetch";
import { type ComponentType, useEffect, useState } from "react";
import type {
	RTCPeerConnection as RTCPeerConnectionType,
	RTCVideoViewProps,
} from "react-native-webrtc";

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

	useEffect(() => {
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

				peer = new webRtc.RTCPeerConnection({
					iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
				});
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
					if (
						!cancelled &&
						peer &&
						(peer.connectionState === "failed" || peer.connectionState === "disconnected")
					) {
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

				const response = await fetch(uri, {
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
				resourceUrl = location ? new URL(location, uri).toString() : null;
				await peer.setRemoteDescription({ type: "answer", sdp: answer });
			} catch {
				if (!cancelled) {
					setState((current) => ({ ...current, failed: true }));
				}
			}
		};

		setState(initialState);
		void connect();
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
	}, [uri]);

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
