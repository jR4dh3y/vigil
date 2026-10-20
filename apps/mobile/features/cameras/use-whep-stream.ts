import { fetch } from "expo/fetch";
import { type ComponentType, useEffect, useEffectEvent, useState } from "react";
import type { RTCVideoViewProps } from "react-native-webrtc";
import { startWhepSession } from "@/features/cameras/whep-session";
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
		let session: ReturnType<typeof startWhepSession> | null = null;
		let retryTimer: ReturnType<typeof setTimeout> | undefined;
		let finishRetry: (() => void) | undefined;
		const stop = () => {
			cancelled = true;
			clearTimeout(fallbackTimer);
			clearTimeout(retryTimer);
			finishRetry?.();
			session?.dispose();
		};
		const fail = () => {
			if (cancelled) return;
			stop();
			setState((current) => ({ ...current, failed: true }));
		};
		const fallbackTimer = setTimeout(fail, 8_000);

		const connect = async () => {
			const webRtc = await import("react-native-webrtc");
			if (cancelled) return;
			// Host candidates suffice for LAN/Tailscale; no public STUN dependency.
			const peer = new webRtc.RTCPeerConnection({ iceServers: [] });
			session = null;
			try {
				peer.addTransceiver("video", { direction: "recvonly" });
				peer.addTransceiver("audio", { direction: "recvonly" });
				peer.ontrack = (event: unknown) => {
					const streamUrl = getTrackStreamUrl(event);
					if (!cancelled && streamUrl) {
						clearTimeout(fallbackTimer);
						setState({ streamUrl, RtcView: webRtc.RTCView, failed: false });
					}
				};
				peer.onconnectionstatechange = () => {
					if (peer.connectionState === "failed") fail();
				};
				session = startWhepSession(peer, connectionUri, fetch);
				await session.ready;
			} catch (cause) {
				if (session) session.dispose();
				else peer.close();
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
					finishRetry = resolve;
					retryTimer = setTimeout(resolve, 700);
					await promise;
				}
			}
			fail();
		})();
		return stop;
	}, [endpoint]);

	return state;
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
