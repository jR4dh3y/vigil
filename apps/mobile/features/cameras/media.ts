import type { LiveStream, PlaybackSession } from "@nvr/api-client";
import { resolveMediaUrl } from "@/lib/api/config";

type ResolvedLiveStream =
	| { kind: "pending" }
	| { kind: "ready"; hlsUrl: string; whepUrl: string }
	| { kind: "error"; error: Error };

type ResolvedPlayback =
	| { kind: "pending" }
	| { kind: "ready"; url: string; startOffsetSec: number }
	| { kind: "error"; error: Error };

export function resolveLiveStream(stream?: LiveStream): ResolvedLiveStream {
	if (!stream) {
		return { kind: "pending" };
	}
	try {
		return {
			kind: "ready",
			hlsUrl: resolveMediaUrl(stream.hlsUrl, stream.token),
			whepUrl: resolveMediaUrl(stream.whepUrl, stream.token),
		};
	} catch (cause) {
		return { kind: "error", error: mediaResolutionError(cause) };
	}
}

export function resolvePlayback(session?: PlaybackSession): ResolvedPlayback {
	if (!session) {
		return { kind: "pending" };
	}
	try {
		return {
			kind: "ready",
			url: resolveMediaUrl(session.playbackUrl, session.token),
			startOffsetSec: session.startOffsetSec,
		};
	} catch (cause) {
		return { kind: "error", error: mediaResolutionError(cause) };
	}
}

function mediaResolutionError(cause: unknown): Error {
	return cause instanceof Error
		? cause
		: new Error("The recorder returned an invalid media address");
}
