import type { Camera } from "@nvr/api-client";
import { useQuery } from "@tanstack/react-query";
import { useIsFocused } from "expo-router/react-navigation";
import { useCallback, useRef, useState } from "react";
import {
	FlatList,
	type ListRenderItem,
	RefreshControl,
	StyleSheet,
	View,
	type ViewToken,
} from "react-native";
import { SectionHeading } from "@/components/section-heading";
import { StatePanel } from "@/components/state-panel";
import { cameraKeys, listCameras } from "@/features/cameras/api";
import { CameraCard } from "@/features/cameras/components/camera-card";
import { colors } from "@/theme/colors";

export default function LiveScreen() {
	const isFocused = useIsFocused();
	const [visibleCameraIds, setVisibleCameraIds] = useState<ReadonlySet<string>>(() => new Set());
	const viewabilityConfig = useRef({ itemVisiblePercentThreshold: 40 }).current;
	const onViewableItemsChanged = useRef(
		({ viewableItems }: { viewableItems: ViewToken<Camera>[] }) => {
			const next = new Set(
				viewableItems.flatMap(({ item, isViewable }) => (isViewable ? [item.id] : [])),
			);
			setVisibleCameraIds((current) => (sameIds(current, next) ? current : next));
		},
	).current;
	const camerasQuery = useQuery({
		queryKey: cameraKeys.all,
		queryFn: listCameras,
		enabled: isFocused,
		refetchInterval: 30_000,
	});
	const cameras = camerasQuery.data ?? [];
	const onlineCount = cameras.filter((camera) => camera.status === "online").length;
	const renderCamera = useCallback<ListRenderItem<Camera>>(
		({ item }) => <CameraCard active={isFocused && visibleCameraIds.has(item.id)} camera={item} />,
		[isFocused, visibleCameraIds],
	);

	return (
		<FlatList
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			data={cameras}
			ItemSeparatorComponent={CameraSeparator}
			keyExtractor={cameraKey}
			ListEmptyComponent={
				camerasQuery.isPending ? (
					<StatePanel detail="Checking camera availability…" loading title="Connecting" />
				) : camerasQuery.isError ? (
					<StatePanel
						actionLabel="Try again"
						detail={camerasQuery.error.message}
						onAction={() => camerasQuery.refetch()}
						title="Cameras are out of reach"
					/>
				) : (
					<StatePanel
						detail="Add and enable a camera from the web dashboard, then it will appear here."
						title="No cameras yet"
					/>
				)
			}
			ListHeaderComponent={
				<View style={styles.header}>
					<SectionHeading
						detail={cameras.length > 0 ? `${onlineCount} of ${cameras.length} online` : undefined}
						title="Cameras"
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
			onViewableItemsChanged={onViewableItemsChanged}
			style={styles.screen}
			viewabilityConfig={viewabilityConfig}
		/>
	);
}

function sameIds(current: ReadonlySet<string>, next: ReadonlySet<string>): boolean {
	if (current.size !== next.size) {
		return false;
	}
	for (const id of current) {
		if (!next.has(id)) {
			return false;
		}
	}
	return true;
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
		paddingBottom: 14,
	},
	separator: {
		height: 14,
	},
});
