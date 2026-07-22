import { useVideoPlayer, VideoView } from "expo-video";
import { StyleSheet, View } from "react-native";

type LivePlayerProps = {
	uri: string;
};

export function LivePlayer({ uri }: LivePlayerProps) {
	const player = useVideoPlayer(uri, (instance) => {
		instance.play();
	});

	return (
		<View style={styles.container}>
			<VideoView
				allowsPictureInPicture
				contentFit="contain"
				fullscreenOptions={{ enable: true }}
				nativeControls
				player={player}
				style={styles.video}
			/>
		</View>
	);
}

const styles = StyleSheet.create({
	container: {
		aspectRatio: 16 / 9,
		backgroundColor: "#000000",
		borderCurve: "continuous",
		borderRadius: 22,
		overflow: "hidden",
	},
	video: {
		flex: 1,
	},
});
