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
	const playbackKey = cameraKeys.playback(cameraId ?? "", event?.startedAt ?? "");
	const continuePlaybackMutation = useMutation({
		mutationFn: ({
			cameraId,
			nextStart,
		}: {
			cameraId: string;
			nextStart: string;
			playbackKey: ReturnType<typeof cameraKeys.playback>;
		}) => getPlayback(cameraId, nextStart),
		onSuccess: (nextSession, variables) => {
			queryClient.setQueryData(variables.playbackKey, nextSession);
		},
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
	const continuePlayback = () => {
		const session = playbackQuery.data;
		if (
			session?.source !== "gdrive" ||
			!session.nextRecordingStart ||
			continuePlaybackMutation.isPending
		) {
			return;
		}
		continuePlaybackMutation.mutate({
			cameraId: session.cameraId,
			nextStart: session.nextRecordingStart,
			playbackKey,
		});
	};
	const retryPlayback = () => {
		continuePlaybackMutation.reset();
		return playbackQuery.refetch();
	};

	return {
		event,
		camera: cameraQuery.data,
		playbackUrl,
		eventPending: eventQuery.isPending,
		eventError: eventQuery.error,
		playbackPending: playbackQuery.isPending || continuePlaybackMutation.isPending,
		playbackError: continuePlaybackMutation.error ?? playbackQuery.error,
		acknowledging: acknowledgeMutation.isPending,
		acknowledgeError: acknowledgeMutation.error,
		retryEvent: eventQuery.refetch,
		retryPlayback,
		continuePlayback,
		acknowledge: acknowledgeMutation.mutate,
	};
}
