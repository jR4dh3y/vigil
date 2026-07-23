import { useVideoPlayer, VideoView } from "expo-video";
import { useEffect } from "react";
import { type StyleProp, StyleSheet, View, type ViewStyle } from "react-native";

type LivePlayerProps = {
	uri: string;
	/** Fills parent by default; pass styles to override. */
	style?: StyleProp<ViewStyle>;
	/** Show native transport controls (default off for mosaic tiles). */
	nativeControls?: boolean;
};

export function LivePlayer({ uri, style, nativeControls = false }: LivePlayerProps) {
	const player = useVideoPlayer(uri, (instance) => {
		instance.loop = true;
		instance.muted = true;
		instance.play();
	});

	useEffect(() => {
		player.replace(uri);
		player.play();
	}, [player, uri]);

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
