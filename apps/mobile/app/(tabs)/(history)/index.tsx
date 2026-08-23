import type { Camera } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { useCallback } from "react";
import { FlatList, type ListRenderItem, RefreshControl, StyleSheet, View } from "react-native";
import { StatePanel } from "@/components/state-panel";
import { cameraKeys, listCameras } from "@/features/cameras/api";
import { HistoryCameraRow } from "@/features/recordings/components/history-camera-row";
import { colors } from "@/theme/colors";

export default function HistoryScreen() {
	const camerasQuery = useQuery({ queryKey: cameraKeys.all, queryFn: listCameras });
	const renderCamera = useCallback<ListRenderItem<Camera>>(
		({ item }) => (
			<HistoryCameraRow
				enabled={item.enabled}
				host={item.host}
				id={item.id}
				name={item.name}
				status={item.status}
			/>
		),
		[],
	);

	return (
		<FlatList
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			data={camerasQuery.data ?? []}
			ItemSeparatorComponent={CameraSeparator}
			keyExtractor={cameraKey}
			ListEmptyComponent={
				camerasQuery.isPending ? (
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
				)
			}
			ListHeaderComponent={
				<View style={styles.header}>
					<StatePanel
						detail="Choose a camera, then jump directly to any retained calendar day."
						title="Recording history"
					/>
				</View>
			}
			refreshControl={
				<RefreshControl
					refreshing={camerasQuery.isRefetching}
					onRefresh={() => camerasQuery.refetch()}
					tintColor={colors.accent}
				/>
			}
			renderItem={renderCamera}
			style={styles.screen}
		/>
	);
}

function cameraKey(camera: Camera): string {
	return camera.id;
}

function CameraSeparator() {
	return <View style={styles.separator} />;
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		padding: 16,
		paddingBottom: 40,
	},
	header: {
		paddingBottom: 18,
	},
	separator: {
		height: 12,
	},
});
