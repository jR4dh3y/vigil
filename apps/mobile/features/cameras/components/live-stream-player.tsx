import { ActivityIndicator, StyleSheet, View } from "react-native";
import { LivePlayer } from "@/features/cameras/components/live-player";
import { useWhepStream } from "@/features/cameras/use-whep-stream";
import { swatches } from "@/theme/colors";

type LiveStreamPlayerProps = {
	whepUri: string;
	hlsUri: string;
	nativeControls?: boolean;
};

export function LiveStreamPlayer({
	whepUri,
	hlsUri,
	nativeControls = false,
}: LiveStreamPlayerProps) {
	const { streamUrl, RtcView, failed } = useWhepStream(whepUri);

	if (failed) {
		return <LivePlayer nativeControls={nativeControls} uri={hlsUri} />;
	}
	if (streamUrl && RtcView) {
		return (
			<RtcView
				objectFit={nativeControls ? "contain" : "cover"}
				streamURL={streamUrl}
				style={styles.video}
			/>
		);
	}
	return (
		<View style={styles.loading}>
			<ActivityIndicator color={swatches.white} />
		</View>
	);
}

const styles = StyleSheet.create({
	video: {
		...StyleSheet.absoluteFill,
		backgroundColor: swatches.black,
	},
	loading: {
		...StyleSheet.absoluteFill,
		alignItems: "center",
		backgroundColor: swatches.preview,
		justifyContent: "center",
	},
});
