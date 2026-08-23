import { useEventListener } from "expo";
import { useVideoPlayer, VideoView } from "expo-video";
import { useCallback, useEffect, useRef, useState } from "react";
import {
	type GestureResponderEvent,
	Pressable,
	type StyleProp,
	StyleSheet,
	Text,
	View,
	type ViewStyle,
} from "react-native";
import { swatches } from "@/theme/colors";

type LivePlayerProps = {
	uri: string;
	/** Fills parent by default; pass styles to override. */
	style?: StyleProp<ViewStyle>;
	/** Show native transport controls (default off for mosaic tiles). */
	nativeControls?: boolean;
	/** Live streams loop. Recorded segments use false so the next segment can load. */
	loop?: boolean;
	/** Seek applied once after the source metadata loads. */
	startOffsetSec?: number;
	onPlaybackEnded?: () => void;
	/** Renews an expired signed source. Live streams retry the current URL by default. */
	onRetry?: () => unknown;
};

export function LivePlayer({
	uri,
	style,
	nativeControls = false,
	loop = true,
	startOffsetSec = 0,
	onPlaybackEnded,
	onRetry,
}: LivePlayerProps) {
	const [error, setError] = useState<string | null>(null);
	const [reloadAttempt, setReloadAttempt] = useState(0);
	const [retrying, setRetrying] = useState(false);
	const appliedSource = useRef<string | null>(null);
	const expectedSource = useRef<string | null>(null);
	const expectedSeek = useRef(false);
	const expectedStartOffset = useRef(startOffsetSec);
	const player = useVideoPlayer(null, (instance) => {
		instance.loop = loop;
		instance.muted = true;
	});
	const sourceKey = `${uri}\u0000${startOffsetSec}\u0000${reloadAttempt}`;
	const handlePlaybackEnded = useCallback(() => {
		onPlaybackEnded?.();
	}, [onPlaybackEnded]);
	const handleRetry = useCallback(
		(event: GestureResponderEvent) => {
			event.stopPropagation();
			if (retrying) {
				return;
			}
			if (onRetry) {
				setRetrying(true);
				void Promise.resolve()
					.then(onRetry)
					.finally(() => setRetrying(false));
				return;
			}
			setReloadAttempt((current) => current + 1);
		},
		[onRetry, retrying],
	);

	useEventListener(player, "playToEnd", handlePlaybackEnded);
	useEventListener(player, "sourceLoad", ({ duration, videoSource }) => {
		const expectedSourceKey = expectedSource.current;
		if (!videoSource || !expectedSourceKey || appliedSource.current === expectedSourceKey) {
			return;
		}

		appliedSource.current = expectedSourceKey;
		if (expectedSeek.current) {
			player.currentTime = clampPlaybackOffset(expectedStartOffset.current, duration);
		}
		player.play();
	});
	useEventListener(player, "statusChange", ({ error: playerError, status }) => {
		if (status === "error") {
			setError(playerError?.message.trim() || "The video could not be played.");
		}
	});

	useEffect(() => {
		appliedSource.current = null;
		expectedSource.current = sourceKey;
		expectedSeek.current = !loop || startOffsetSec > 0;
		expectedStartOffset.current = startOffsetSec;
		setError(null);
		setRetrying(false);
		player.loop = loop;
		player.replace(uri);
	}, [loop, player, sourceKey, startOffsetSec, uri]);

	return (
		<View
			style={[styles.container, style]}
			pointerEvents={nativeControls || error ? "auto" : "none"}
		>
			{error ? (
				<View style={styles.error}>
					<Text selectable numberOfLines={2} style={styles.errorMessage}>
						{error}
					</Text>
					<Pressable
						accessibilityRole="button"
						disabled={retrying}
						onPress={handleRetry}
						style={({ pressed }) => [styles.retry, pressed ? styles.retryPressed : null]}
					>
						<Text style={styles.retryLabel}>{retrying ? "Retrying…" : "Retry video"}</Text>
					</Pressable>
				</View>
			) : (
				<VideoView
					allowsPictureInPicture={nativeControls}
					contentFit="cover"
					fullscreenOptions={{ enable: nativeControls }}
					nativeControls={nativeControls}
					player={player}
					style={styles.video}
				/>
			)}
		</View>
	);
}

function clampPlaybackOffset(offset: number, duration: number): number {
	if (!Number.isFinite(offset) || offset <= 0) {
		return 0;
	}
	if (!Number.isFinite(duration) || duration <= 0) {
		return offset;
	}
	return Math.min(offset, duration);
}

const styles = StyleSheet.create({
	container: {
		...StyleSheet.absoluteFill,
		backgroundColor: swatches.black,
	},
	video: {
		flex: 1,
	},
	error: {
		...StyleSheet.absoluteFill,
		alignItems: "center",
		backgroundColor: swatches.preview,
		gap: 12,
		justifyContent: "center",
		padding: 16,
	},
	errorMessage: {
		color: swatches.white,
		fontSize: 13,
		lineHeight: 18,
		textAlign: "center",
	},
	retry: {
		backgroundColor: swatches.white,
		borderCurve: "continuous",
		borderRadius: 11,
		paddingHorizontal: 14,
		paddingVertical: 9,
	},
	retryPressed: {
		opacity: 0.7,
	},
	retryLabel: {
		color: swatches.black,
		fontSize: 13,
		fontWeight: "700",
	},
});
