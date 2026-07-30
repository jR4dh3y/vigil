import { Stack, useLocalSearchParams } from "expo-router";
import { useIsFocused } from "expo-router/react-navigation";
import { EventDetail } from "@/features/events/components/event-detail";
import { useEventDetail } from "@/features/events/use-event-detail";

export default function EventDetailScreen() {
	const { id } = useLocalSearchParams<{ id: string }>();
	const isFocused = useIsFocused();
	const detail = useEventDetail(id, isFocused);

	return (
		<>
			<Stack.Title>{detail.event?.title ?? "Event"}</Stack.Title>
			<EventDetail
				acknowledgeError={detail.acknowledgeError}
				acknowledging={detail.acknowledging}
				camera={detail.camera}
				event={detail.event}
				eventError={detail.eventError}
				eventPending={detail.eventPending}
				onAcknowledge={detail.acknowledge}
				onRetryEvent={() => detail.retryEvent()}
				onRetryPlayback={() => detail.retryPlayback()}
				playbackError={detail.playbackError}
				playbackPending={detail.playbackPending}
				playbackUrl={isFocused ? detail.playbackUrl : undefined}
			/>
		</>
	);
}
