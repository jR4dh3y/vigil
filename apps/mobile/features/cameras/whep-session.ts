import type { RTCPeerConnection } from "react-native-webrtc";

type WhepPeer = Pick<
	RTCPeerConnection,
	| "createOffer"
	| "setLocalDescription"
	| "setRemoteDescription"
	| "localDescription"
	| "iceGatheringState"
	| "onicegatheringstatechange"
	| "ontrack"
	| "onconnectionstatechange"
	| "close"
>;

// Owns one negotiation, including responses that arrive after disposal.
export function startWhepSession(peer: WhepPeer, uri: string, request: typeof fetch) {
	const controller = new AbortController();
	const { signal } = controller;
	let resourceUrl: string | null = null;
	const deleteResource = () => {
		if (!resourceUrl) return;
		const url = resourceUrl;
		resourceUrl = null;
		void request(url, { method: "DELETE" }).catch(() => undefined);
	};
	const dispose = () => {
		if (!signal.aborted) {
			controller.abort();
			peer.ontrack = null;
			peer.onconnectionstatechange = null;
			peer.onicegatheringstatechange = null;
			peer.close();
		}
		deleteResource();
	};
	const ready = (async () => {
		try {
			const offer = await peer.createOffer();
			signal.throwIfAborted();
			await peer.setLocalDescription(offer);
			signal.throwIfAborted();
			await waitForIceGathering(peer, signal);
			signal.throwIfAborted();
			const sdp = peer.localDescription?.sdp;
			if (!sdp) throw new Error("WHEP offer has no SDP");
			const response = await request(uri, {
				method: "POST",
				headers: { Accept: "application/sdp", "Content-Type": "application/sdp" },
				body: sdp,
				signal,
			});
			const location = response.headers.get("Location");
			if (location) {
				const resolved = new URL(location, uri);
				for (const [key, value] of new URL(uri).searchParams) {
					if (!resolved.searchParams.has(key)) resolved.searchParams.set(key, value);
				}
				resourceUrl = resolved.toString();
			}
			signal.throwIfAborted();
			if (!response.ok) throw new Error(`WHEP request failed (${response.status})`);
			const answer = await response.text();
			signal.throwIfAborted();
			if (!answer.trim()) throw new Error("WHEP returned an empty response");
			await peer.setRemoteDescription({ type: "answer", sdp: answer });
			signal.throwIfAborted();
		} catch (cause) {
			dispose();
			throw cause;
		}
	})();
	return { ready, dispose };
}

function waitForIceGathering(peer: WhepPeer, signal: AbortSignal): Promise<void> {
	if (peer.iceGatheringState === "complete") return Promise.resolve();
	return new Promise((resolve) => {
		const finish = () => {
			clearTimeout(timer);
			peer.onicegatheringstatechange = null;
			signal.removeEventListener("abort", finish);
			resolve();
		};
		const timer = setTimeout(finish, 5_000);
		signal.addEventListener("abort", finish, { once: true });
		peer.onicegatheringstatechange = () => {
			if (peer.iceGatheringState === "complete") finish();
		};
	});
}
