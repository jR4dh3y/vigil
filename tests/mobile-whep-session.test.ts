import { describe, expect, mock, test } from "bun:test";
import { startWhepSession } from "../apps/mobile/features/cameras/whep-session";

type Peer = Parameters<typeof startWhepSession>[0];
const uri = "http://recorder/live/whep?token=signed";
const offer = { type: "offer", sdp: "offer" };

function createPeer() {
	return {
		createOffer: mock(async () => offer),
		setLocalDescription: mock(async () => undefined),
		setRemoteDescription: mock(async () => undefined),
		localDescription: { ...offer, _sdp: offer.sdp, _type: offer.type, toJSON: () => offer },
		iceGatheringState: "complete",
		onicegatheringstatechange: null,
		ontrack: null,
		onconnectionstatechange: null,
		close: mock(() => undefined),
	} satisfies Peer;
}

async function flushNegotiation() {
	for (let i = 0; i < 10; i += 1) await Promise.resolve();
}

describe("mobile WHEP cancellation", () => {
	test("never sets a local description after scroll-out during offer creation", async () => {
		const peer = createPeer();
		const pending = Promise.withResolvers<typeof offer>();
		peer.createOffer.mockImplementation(() => pending.promise);
		const request = mock<typeof fetch>(async () => new Response("answer"));
		const session = startWhepSession(peer, uri, request);
		const result = session.ready.catch(() => "cancelled");
		session.dispose();
		session.dispose();
		pending.resolve(offer);
		expect(await result).toBe("cancelled");
		expect(peer.setLocalDescription).not.toHaveBeenCalled();
		expect(request).not.toHaveBeenCalled();
		expect(peer.close).toHaveBeenCalledTimes(1);
	});

	test("stops after an in-flight local description completes", async () => {
		const peer = createPeer();
		const pending = Promise.withResolvers<void>();
		peer.setLocalDescription.mockImplementation(() => pending.promise);
		const request = mock<typeof fetch>(async () => new Response("answer"));
		const session = startWhepSession(peer, uri, request);
		const result = session.ready.catch(() => "cancelled");
		await flushNegotiation();
		session.dispose();
		pending.resolve();
		expect(await result).toBe("cancelled");
		expect(request).not.toHaveBeenCalled();
		expect(peer.close).toHaveBeenCalledTimes(1);
	});

	test("cancels ICE gathering immediately and removes its listener", async () => {
		const peer: Peer = { ...createPeer(), iceGatheringState: "gathering" };
		const request = mock<typeof fetch>(async () => new Response("answer"));
		const session = startWhepSession(peer, uri, request);
		const result = session.ready.catch(() => "cancelled");
		await flushNegotiation();
		expect(peer.onicegatheringstatechange).not.toBeNull();
		session.dispose();
		expect(await result).toBe("cancelled");
		expect(peer.onicegatheringstatechange).toBeNull();
		expect(request).not.toHaveBeenCalled();
	});

	test("deletes a late server session without applying it to the closed peer", async () => {
		const peer = createPeer();
		const pending = Promise.withResolvers<Response>();
		const request = mock<typeof fetch>(async (_url, options) =>
			options?.method === "POST" ? pending.promise : new Response(null, { status: 204 }),
		);
		const session = startWhepSession(peer, uri, request);
		const result = session.ready.catch(() => "cancelled");
		await flushNegotiation();
		expect(request).toHaveBeenCalledTimes(1);
		const signal = request.mock.calls[0]?.[1]?.signal;
		session.dispose();
		expect(signal?.aborted).toBe(true);
		pending.resolve(new Response("answer", { status: 201, headers: { Location: "/session/1" } }));
		expect(await result).toBe("cancelled");
		expect(peer.setRemoteDescription).not.toHaveBeenCalled();
		expect(request).toHaveBeenLastCalledWith("http://recorder/session/1?token=signed", {
			method: "DELETE",
		});
		expect(peer.close).toHaveBeenCalledTimes(1);
	});

	test("releases a session on failure before a retry starts", async () => {
		const peer = createPeer();
		const request = mock<typeof fetch>(async (_url, options) =>
			options?.method === "POST"
				? new Response("", { status: 201, headers: { Location: "/session/failed" } })
				: new Response(null, { status: 204 }),
		);
		const session = startWhepSession(peer, uri, request);
		await expect(session.ready).rejects.toThrow("empty response");
		session.dispose();
		expect(peer.close).toHaveBeenCalledTimes(1);
		expect(request).toHaveBeenCalledTimes(2);
	});

	test("stops when scroll-out happens while the answer body is loading", async () => {
		const peer = createPeer();
		const body = Promise.withResolvers<string>();
		class PendingAnswer extends Response {
			override text() {
				return body.promise;
			}
		}
		const request = mock<typeof fetch>(async (_url, options) =>
			options?.method === "POST"
				? new PendingAnswer(null, { status: 201, headers: { Location: "/session/body" } })
				: new Response(null, { status: 204 }),
		);
		const session = startWhepSession(peer, uri, request);
		const result = session.ready.catch(() => "cancelled");
		await flushNegotiation();
		session.dispose();
		body.resolve("answer");
		expect(await result).toBe("cancelled");
		expect(peer.setRemoteDescription).not.toHaveBeenCalled();
		expect(request).toHaveBeenCalledTimes(2);
		expect(peer.close).toHaveBeenCalledTimes(1);
	});

	test("keeps a negotiated peer alive until disposal and deletes its session once", async () => {
		const peer = createPeer();
		const request = mock<typeof fetch>(async (_url, options) =>
			options?.method === "POST"
				? new Response("answer", { status: 201, headers: { Location: "/session/ok" } })
				: new Response(null, { status: 204 }),
		);
		const session = startWhepSession(peer, uri, request);
		await session.ready;
		expect(peer.setRemoteDescription).toHaveBeenCalledWith({ type: "answer", sdp: "answer" });
		expect(peer.close).not.toHaveBeenCalled();
		session.dispose();
		session.dispose();
		expect(peer.close).toHaveBeenCalledTimes(1);
		expect(request).toHaveBeenCalledTimes(2);
	});
});
