import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { cameraKeys, getCamera, getPlayback } from "@/features/cameras/api";
import { acknowledgeEvent, eventKeys, getEvent } from "@/features/events/api";
import { resolveMediaUrl } from "@/lib/api/config";

export function useEventDetail(id: string, active: boolean) {
	const queryClient = useQueryClient();
	const eventQuery = useQuery({
		queryKey: eventKeys.detail(id),
		queryFn: () => getEvent(id),
		enabled: Boolean(id),
	});
	const event = eventQuery.data;
	const cameraId = event?.cameraId ?? undefined;
	const cameraQuery = useQuery({
		queryKey: cameraKeys.detail(cameraId ?? ""),
		queryFn: () => getCamera(cameraId ?? ""),
		enabled: Boolean(cameraId),
	});
	const playbackQuery = useQuery({
		queryKey: cameraKeys.playback(cameraId ?? "", event?.startedAt ?? ""),
		queryFn: () => getPlayback(cameraId ?? "", event?.startedAt ?? ""),
		enabled: active && Boolean(cameraId && event?.startedAt),
		staleTime: 45_000,
	});
	const acknowledgeMutation = useMutation({
		mutationFn: () => acknowledgeEvent(id),
		onSuccess: async (updated) => {
			queryClient.setQueryData(eventKeys.detail(id), updated);
			await queryClient.invalidateQueries({ queryKey: eventKeys.all });
		},
	});
	const playbackUrl = playbackQuery.data
		? resolveMediaUrl(playbackQuery.data.playbackUrl, playbackQuery.data.token)
		: undefined;

	return {
		event,
		camera: cameraQuery.data,
		playbackUrl,
		eventPending: eventQuery.isPending,
		eventError: eventQuery.error,
		playbackPending: playbackQuery.isPending,
		playbackError: playbackQuery.error,
		acknowledging: acknowledgeMutation.isPending,
		acknowledgeError: acknowledgeMutation.error,
		retryEvent: eventQuery.refetch,
		retryPlayback: playbackQuery.refetch,
		acknowledge: acknowledgeMutation.mutate,
	};
}
