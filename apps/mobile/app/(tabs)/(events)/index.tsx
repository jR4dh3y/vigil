import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { FlatList, RefreshControl, StyleSheet, View } from "react-native";
import { SectionHeading } from "@/components/section-heading";
import { StatePanel } from "@/components/state-panel";
import { acknowledgeEvent, eventKeys, listEvents } from "@/features/events/api";
import { EventCard } from "@/features/events/components/event-card";
import { EventFilter } from "@/features/events/components/event-filter";
import { colors } from "@/theme/colors";

export default function EventsScreen() {
	const [unacknowledgedOnly, setUnacknowledgedOnly] = useState(false);
	const [acknowledgingId, setAcknowledgingId] = useState<string | null>(null);
	const queryClient = useQueryClient();
	const eventsQuery = useQuery({
		queryKey: eventKeys.list(unacknowledgedOnly),
		queryFn: () => listEvents(unacknowledgedOnly),
		refetchInterval: 15_000,
	});
	const acknowledgeMutation = useMutation({
		mutationFn: acknowledgeEvent,
		onMutate: (id) => setAcknowledgingId(id),
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: eventKeys.all });
		},
		onSettled: () => setAcknowledgingId(null),
	});

	return (
		<FlatList
			contentInsetAdjustmentBehavior="automatic"
			contentContainerStyle={styles.content}
			data={eventsQuery.data ?? []}
			ItemSeparatorComponent={EventSeparator}
			keyExtractor={(event) => event.id}
			ListEmptyComponent={
				eventsQuery.isPending ? (
					<StatePanel
						detail="Loading the latest camera and system activity…"
						loading
						title="Checking events"
					/>
				) : eventsQuery.isError ? (
					<StatePanel
						actionLabel="Try again"
						detail={eventsQuery.error.message}
						onAction={() => eventsQuery.refetch()}
						title="Events are out of reach"
					/>
				) : (
					<StatePanel
						detail={
							unacknowledgedOnly
								? "Everything has been reviewed."
								: "Camera and system activity will appear here."
						}
						title={unacknowledgedOnly ? "You’re all caught up" : "No recent activity"}
					/>
				)
			}
			ListHeaderComponent={
				<View style={styles.header}>
					<SectionHeading detail="Auto-refreshes" title="Recent activity" />
					<EventFilter onChange={setUnacknowledgedOnly} unacknowledgedOnly={unacknowledgedOnly} />
				</View>
			}
			refreshControl={
				<RefreshControl
					refreshing={eventsQuery.isRefetching}
					onRefresh={() => eventsQuery.refetch()}
					tintColor={colors.accent}
				/>
			}
			renderItem={({ item }) => (
				<EventCard
					acknowledging={acknowledgingId === item.id}
					event={item}
					onAcknowledge={(id) => acknowledgeMutation.mutate(id)}
				/>
			)}
			style={styles.screen}
		/>
	);
}

function EventSeparator() {
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
		gap: 14,
		paddingBottom: 16,
	},
	separator: {
		height: 12,
	},
});
