import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import { RefreshControl, ScrollView, StyleSheet, View } from "react-native";
import { SectionHeading } from "@/components/section-heading";
import { StatePanel } from "@/components/state-panel";
import { cameraKeys, listCameras } from "@/features/cameras/api";
import { ArmCard } from "@/features/cameras/components/arm-card";
import { CameraCard } from "@/features/cameras/components/camera-card";
import { colors } from "@/theme/colors";

export default function LiveScreen() {
	const camerasQuery = useQuery({
		queryKey: cameraKeys.all,
		queryFn: listCameras,
		refetchInterval: 30_000,
	});
	const cameras = camerasQuery.data ?? [];
	const onlineCount = cameras.filter((camera) => camera.status === "online").length;

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
			<ArmCard />

			<View style={styles.section}>
				<SectionHeading
					detail={cameras.length > 0 ? `${onlineCount} of ${cameras.length} online` : undefined}
					title="Cameras"
				/>

				{camerasQuery.isPending ? (
					<StatePanel detail="Checking camera availability…" loading title="Connecting" />
				) : camerasQuery.isError ? (
					<StatePanel
						actionLabel="Try again"
						detail={camerasQuery.error.message}
						onAction={() => camerasQuery.refetch()}
						title="Cameras are out of reach"
					/>
				) : cameras.length === 0 ? (
					<StatePanel
						detail="Add and enable a camera from the web dashboard, then it will appear here."
						title="No cameras yet"
					/>
				) : (
					<View style={styles.grid}>
						{cameras.map((camera) => (
							<CameraCard
								camera={camera}
								key={camera.id}
								onPress={() => router.push({ pathname: "/camera/[id]", params: { id: camera.id } })}
							/>
						))}
					</View>
				)}
			</View>
		</ScrollView>
	);
}

const styles = StyleSheet.create({
	screen: {
		backgroundColor: colors.background,
		flex: 1,
	},
	content: {
		gap: 28,
		padding: 16,
		paddingBottom: 40,
	},
	section: {
		gap: 14,
	},
	grid: {
		gap: 14,
	},
});
