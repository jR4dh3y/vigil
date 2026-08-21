import { useEventListener } from "expo";
import { useVideoPlayer, VideoView } from "expo-video";
import { useCallback, useEffect } from "react";
import { type StyleProp, StyleSheet, View, type ViewStyle } from "react-native";

type LivePlayerProps = {
	uri: string;
	/** Fills parent by default; pass styles to override. */
	style?: StyleProp<ViewStyle>;
	/** Show native transport controls (default off for mosaic tiles). */
	nativeControls?: boolean;
	/** Live streams loop. Recorded segments use false so the next segment can load. */
	loop?: boolean;
	onPlaybackEnded?: () => void;
};

export function LivePlayer({
	uri,
	style,
	nativeControls = false,
	loop = true,
	onPlaybackEnded,
}: LivePlayerProps) {
	const player = useVideoPlayer(uri, (instance) => {
		instance.loop = loop;
		instance.muted = true;
		instance.play();
	});
	const handlePlaybackEnded = useCallback(() => {
		onPlaybackEnded?.();
	}, [onPlaybackEnded]);

	useEventListener(player, "playToEnd", handlePlaybackEnded);

	useEffect(() => {
		player.loop = loop;
		player.replace(uri);
		player.play();
	}, [loop, player, uri]);

	return (
		<View style={[styles.container, style]} pointerEvents={nativeControls ? "auto" : "none"}>
			<VideoView
				allowsPictureInPicture={nativeControls}
				contentFit="cover"
				fullscreenOptions={{ enable: nativeControls }}
				nativeControls={nativeControls}
				player={player}
				style={styles.video}
			/>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		...StyleSheet.absoluteFill,
		backgroundColor: "#000000",
	},
	video: {
		flex: 1,
	},
});
