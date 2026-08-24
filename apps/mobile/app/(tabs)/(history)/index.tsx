import type { Camera } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { RefreshControl, ScrollView, StyleSheet } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { cameraKeys, listCameras } from "@/features/cameras/api";
import { HistoryCameraRow } from "@/features/recordings/components/history-camera-row";
import { colors } from "@/theme/colors";

export default function HistoryScreen() {
	const camerasQuery = useQuery({ queryKey: cameraKeys.all, queryFn: listCameras });
	const cameras = camerasQuery.data ?? [];

	return (
		<ScrollView
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			refreshControl={
				<RefreshControl
					refreshing={camerasQuery.isRefetching}
					onRefresh={() => camerasQuery.refetch()}
					tintColor={colors.accent}
				/>
			}
			style={styles.screen}
		>
			<StatePanel
				detail="Choose a camera, then jump directly to any retained calendar day."
				title="Recording history"
			/>
			{cameras.length > 0 ? (
				cameras.map((camera: Camera) => (
					<HistoryCameraRow
						enabled={camera.enabled}
						id={camera.id}
						key={camera.id}
						name={camera.name}
						status={camera.status}
					/>
				))
			) : camerasQuery.isPending ? (
				<StatePanel detail="Loading cameras…" loading title="Checking history" />
			) : camerasQuery.isError ? (
				<StatePanel
					actionLabel="Try again"
					detail={camerasQuery.error.message}
					onAction={() => camerasQuery.refetch()}
					title="History is out of reach"
				/>
			) : (
				<StatePanel
					detail="Add a camera from the web dashboard before browsing retained video."
					title="No cameras yet"
				/>
			)}
		</ScrollView>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		gap: 12,
		padding: 16,
		paddingBottom: 40,
	},
});
